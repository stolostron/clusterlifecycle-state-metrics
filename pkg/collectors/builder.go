// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"sort"
	"strings"
	"time"

	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
	addonclient "open-cluster-management.io/api/client/addon/clientset/versioned"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	workclient "open-cluster-management.io/api/client/work/clientset/versioned"
	mcv1 "open-cluster-management.io/api/cluster/v1"
	workv1 "open-cluster-management.io/api/work/v1"

	ocpclient "github.com/openshift/client-go/config/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-state-metrics/pkg/metric"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
	"k8s.io/kube-state-metrics/pkg/options"

	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators/addon"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators/cluster"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators/work"
)

var ResyncPeriod = 60 * time.Minute

type whiteBlackLister interface {
	IsIncluded(string) bool
	IsExcluded(string) bool
}

// Builder helps to build collectors. It follows the builder pattern
// (https://en.wikipedia.org/wiki/Builder_pattern).
type Builder struct {
	hubType           string
	namespaces        options.NamespaceList
	ctx               context.Context
	enabledCollectors []string
	whiteBlackList    whiteBlackLister
	restConfig        *rest.Config
	kubeclient        kubernetes.Interface

	clusterIdCache               *clusterIdCache
	clusterHibernatingStateCache *clusterHibernatingStateCache
	clusterTimestampCache        *clusterTimestampCache
	composedClusterStore         *composedStore
	composedAddOnStore           *composedStore
	composedManifestWorkStore    *composedStore

	timestampMetricsEnabled bool
}

// NewBuilder returns a new builder.
func NewBuilder(ctx context.Context) *Builder {
	clusterIdCache := newClusterIdCache()
	clusterHibernatingStateCache := newClusterHibernatingStateCache()
	return &Builder{
		ctx:                          ctx,
		clusterIdCache:               clusterIdCache,
		clusterHibernatingStateCache: clusterHibernatingStateCache,
		composedClusterStore:         newComposedStore(clusterIdCache),
		composedAddOnStore:           newComposedStore(),
		composedManifestWorkStore:    newComposedStore(),
	}
}

func (b *Builder) WithKubeclient(kubeclient kubernetes.Interface) *Builder {
	b.kubeclient = kubeclient
	return b
}

func (b *Builder) WithRestConfig(restConfig *rest.Config) *Builder {
	b.restConfig = restConfig
	return b
}

func (b *Builder) WithTimestampMetricsEnabled(timestampMetricsEnabled bool) *Builder {
	b.timestampMetricsEnabled = timestampMetricsEnabled
	klog.InfoS("set timetamp metrics enabled", "enabled", timestampMetricsEnabled)
	if timestampMetricsEnabled && b.clusterTimestampCache == nil {
		clusterTimestampCache := newClusterTimestampCache()
		b.composedManifestWorkStore.AddStore(clusterTimestampCache)
		b.clusterTimestampCache = clusterTimestampCache
	}
	return b
}

func (b *Builder) WithHubType(hubType string) *Builder {
	b.hubType = hubType
	return b
}

// WithEnabledCollectors sets the enabledCollectors property of a Builder.
func (b *Builder) WithEnabledCollectors(c []string) *Builder {
	copy := []string{}
	copy = append(copy, c...)
	sort.Strings(copy)

	b.enabledCollectors = copy
	return b
}

// WithNamespaces sets the namespaces property of a Builder.
func (b *Builder) WithNamespaces(n options.NamespaceList) *Builder {
	b.namespaces = n
	return b
}

// WithWhiteBlackList configures the white or blacklisted metrics to be exposed
// by the collectors build by the Builder
func (b *Builder) WithWhiteBlackList(l whiteBlackLister) *Builder {
	b.whiteBlackList = l
	return b
}

// Build initializes and registers all enabled collectors.
func (b *Builder) Build() []MetricsCollector {
	if b.whiteBlackList == nil {
		panic("whiteBlackList should not be nil")
	}

	collectors := []MetricsCollector{}
	activeCollectorNames := []string{}

	for _, c := range b.enabledCollectors {
		constructor, ok := availableCollectors[c]
		if !ok {
			klog.Fatalf("collector %s is not correct", c)
		}

		collector := constructor(b)
		activeCollectorNames = append(activeCollectorNames, c)
		collectors = append(collectors, collector)

	}

	klog.Infof("Active collectors: %s", strings.Join(activeCollectorNames, ","))

	// start watching resources
	b.startWatchingManagedClusters()
	b.startWatchingClusterDeployments()
	b.startWatchingManagedClusterAddOns()
	b.startWatchingManifestWorks()

	return collectors
}

var availableCollectors = map[string]func(f *Builder) MetricsCollector{
	"managedclusters":      func(b *Builder) MetricsCollector { return b.buildManagedClusterCollector() },
	"managedclusteraddons": func(b *Builder) MetricsCollector { return b.buildManagedClusterAddOnCollector() },
	"manifestworks":        func(b *Builder) MetricsCollector { return b.buildManifestWorkCollector() },
}

func (b *Builder) buildManagedClusterCollector() MetricsCollector {
	// build metrics store
	ocpClient, err := ocpclient.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create ocpclient: %v", err)
	}

	hubClusterID := getHubClusterID(ocpClient, b.kubeclient)

	clusterFamilies := []metric.FamilyGenerator{
		cluster.GetManagedClusterInfoMetricFamilies(hubClusterID, b.hubType),
		cluster.GetManagedClusterLabelMetricFamilies(hubClusterID),
		cluster.GetManagedClusterStatusMetricFamilies(),
		cluster.GetManagedClusterWorkerCoresMetricFamilies(hubClusterID, b.clusterHibernatingStateCache.IsHibernating),
	}
	if b.timestampMetricsEnabled && b.clusterTimestampCache != nil {
		clusterFamilies = append(clusterFamilies,
			cluster.GetManagedClusterTimestampMetricFamilies(hubClusterID, b.clusterTimestampCache.GetClusterTimestamps))
	}
	filteredMetricFamilies := metric.FilterMetricFamilies(b.whiteBlackList, clusterFamilies)
	composedMetricGenFuncs := metric.ComposeMetricGenFuncs(filteredMetricFamilies)
	familyHeaders := metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)
	metricsStore := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)

	// register to the composed cluster store
	b.composedClusterStore.AddStore(metricsStore)

	// build counter metrics store
	filteredMetricFamilies = metric.FilterMetricFamilies(b.whiteBlackList,
		[]metric.FamilyGenerator{
			cluster.GetManagedClusterCountMetricFamilies(),
		})
	composedMetricGenFuncs = metric.ComposeMetricGenFuncs(filteredMetricFamilies)
	familyHeaders = metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)
	counterMetricsStore := newCounterMetricsStore(familyHeaders, composedMetricGenFuncs)

	// register to the composed cluster store
	b.composedClusterStore.AddStore(counterMetricsStore)

	// return a composed collector
	return newComposedMetricsCollector(metricsStore, counterMetricsStore)
}

func (b *Builder) buildManagedClusterAddOnCollector() MetricsCollector {
	filteredMetricFamilies := metric.FilterMetricFamilies(b.whiteBlackList,
		[]metric.FamilyGenerator{
			addon.GetManagedClusterAddOnStatusMetricFamilies(b.clusterIdCache.GetClusterId),
		})
	composedMetricGenFuncs := metric.ComposeMetricGenFuncs(filteredMetricFamilies)
	familyHeaders := metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)
	metricsStore := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)

	// register to the composed addon store
	b.composedAddOnStore.AddStore(metricsStore)

	return metricsStore
}

func (b *Builder) buildManifestWorkCollector() MetricsCollector {
	// build metrics store
	workFamilies := []metric.FamilyGenerator{
		work.GetManifestWorkStatusMetricFamilies(b.clusterIdCache.GetClusterId),
	}
	if b.timestampMetricsEnabled {
		workFamilies = append(workFamilies, work.GetManifestWorkTimestampMetricFamilies(b.clusterIdCache.GetClusterId))
	}
	filteredMetricFamilies := metric.FilterMetricFamilies(b.whiteBlackList, workFamilies)
	composedMetricGenFuncs := metric.ComposeMetricGenFuncs(filteredMetricFamilies)
	familyHeaders := metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)
	metricsStore := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)

	// register to the composed manifestwork store
	b.composedManifestWorkStore.AddStore(metricsStore)

	// build counter metrics store
	filteredMetricFamilies = metric.FilterMetricFamilies(b.whiteBlackList,
		[]metric.FamilyGenerator{
			work.GetManifestWorkCountMetricFamilies(),
		})
	composedMetricGenFuncs = metric.ComposeMetricGenFuncs(filteredMetricFamilies)
	familyHeaders = metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)
	counterMetricsStore := newCounterMetricsStore(familyHeaders, composedMetricGenFuncs)

	// register to the composed manifestwork store
	b.composedManifestWorkStore.AddStore(counterMetricsStore)

	// return a composed collector
	return newComposedMetricsCollector(metricsStore, counterMetricsStore)
}

func (b *Builder) startWatchingManagedClusters() {
	clusterClient, err := clusterclient.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create clusterclient: %v", err)
	}

	// initialize clusterID cache
	clusterList, err := clusterClient.ClusterV1().ManagedClusters().List(b.ctx, metav1.ListOptions{})
	if errors.IsNotFound(err) {
		klog.Errorf("cannot list managed clusters: %v", err)
	} else if err != nil {
		klog.Fatalf("cannot list managed clusters: %v", err)
	}

	clusters := []interface{}{}
	for index := range clusterList.Items {
		clusters = append(clusters, &clusterList.Items[index])
	}
	if err := b.clusterIdCache.Replace(clusters, ""); err != nil {
		klog.Fatalf("cannot initialize clusterID cache: %v", err)
	}
	klog.Infof("Cluster ID cached for %d managed clusters", len(clusters))

	if b.timestampMetricsEnabled && b.clusterTimestampCache != nil {
		// refresh the managed cluster store once the timestamp of a certian cluster is changed
		b.clusterTimestampCache.AddOnTimestampChangeFunc(func(clusterName string) error {
			klog.Infof("Refresh the managed cluster metrics since the timestamp of cluster %q is changed", clusterName)
			cluster, err := clusterClient.ClusterV1().ManagedClusters().Get(b.ctx, clusterName, metav1.GetOptions{})
			if err != nil {
				return err
			}

			return b.composedClusterStore.Update(cluster)
		})
	}

	// start watching managed clusters
	lw := cache.NewListWatchFromClient(clusterClient.ClusterV1().RESTClient(), "managedclusters", metav1.NamespaceAll, fields.Everything())
	reflector := cache.NewReflector(lw, &mcv1.ManagedCluster{}, b.composedClusterStore, ResyncPeriod)

	klog.Infof("Start watching ManagedClusters")
	go reflector.Run(b.ctx.Done())
}

func (b *Builder) startWatchingManagedClusterAddOns() {
	if b.composedAddOnStore.Size() == 0 {
		return
	}

	addOnClient, err := addonclient.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create addonclient: %v", err)
	}

	// refresh the addon store once the cluster ID of a certian cluster is changed
	b.clusterIdCache.AddOnClusterIdChangeFunc(func(clusterName string) error {
		klog.Infof("Refresh the addon metrics since the cluster ID of cluster %q is changed", clusterName)
		addons, err := addOnClient.AddonV1alpha1().ManagedClusterAddOns(clusterName).List(b.ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		errs := []error{}
		for index := range addons.Items {
			if err = b.composedAddOnStore.Update(&addons.Items[index]); err != nil {
				errs = append(errs, err)
			}
		}

		return utilerrors.NewAggregate(errs)
	})

	lw := cache.NewListWatchFromClient(addOnClient.AddonV1alpha1().RESTClient(), "managedclusteraddons",
		metav1.NamespaceAll, fields.Everything())
	reflector := cache.NewReflector(lw, &addonv1alpha1.ManagedClusterAddOn{}, b.composedAddOnStore, ResyncPeriod)

	klog.Infof("Start watching ManagedClusterAddOns")
	go reflector.Run(b.ctx.Done())
}

func (b *Builder) startWatchingManifestWorks() {
	if b.composedManifestWorkStore.Size() == 0 {
		return
	}

	workClient, err := workclient.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create workclient: %v", err)
	}

	// refresh the manifestwork store once the cluster ID of a certian cluster is changed
	b.clusterIdCache.AddOnClusterIdChangeFunc(func(clusterName string) error {
		klog.Infof("Refresh the manifestwork metrics since the cluster ID of cluster %q is changed", clusterName)
		works, err := workClient.WorkV1().ManifestWorks(clusterName).List(b.ctx, metav1.ListOptions{})
		if err != nil {
			return err
		}

		errs := []error{}
		for index := range works.Items {
			if err = b.composedManifestWorkStore.Update(&works.Items[index]); err != nil {
				errs = append(errs, err)
			}
		}

		return utilerrors.NewAggregate(errs)
	})

	lw := cache.NewListWatchFromClient(workClient.WorkV1().RESTClient(), "manifestworks", metav1.NamespaceAll, fields.Everything())
	reflector := cache.NewReflector(lw, &workv1.ManifestWork{}, b.composedManifestWorkStore, ResyncPeriod)

	klog.Infof("Start watching ManifestWorks")
	go reflector.Run(b.ctx.Done())
}

func (b *Builder) startWatchingClusterDeployments() {
	dynamicClient, err := dynamic.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create dynamic client: %v", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    "hive.openshift.io",
		Version:  "v1",
		Resource: "clusterdeployments",
	}

	// initialize hibernating state cache
	clusterDeploymentList, err := dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(b.ctx, metav1.ListOptions{})
	if err != nil {
		klog.Fatalf("cannot list clusterdeployments: %v", err)
	}

	clusterDeployments := []interface{}{}
	for index := range clusterDeploymentList.Items {
		clusterDeployments = append(clusterDeployments, &clusterDeploymentList.Items[index])
	}
	if err := b.clusterHibernatingStateCache.Replace(clusterDeployments, ""); err != nil {
		klog.Fatalf("cannot initialize hibernating state cache: %v", err)
	}
	klog.Infof("Hibernating state cached for %d clusterdeployments", len(clusterDeployments))

	clusterClient, err := clusterclient.NewForConfig(b.restConfig)
	if err != nil {
		klog.Fatalf("cannot create clusterclient: %v", err)
	}
	b.clusterHibernatingStateCache.AddOnHibernatingStateChangeFunc(func(clusterName string) error {
		klog.Infof("Refresh the managed cluster metrics since the hibernating state of cluster %q is changed", clusterName)
		cluster, err := clusterClient.ClusterV1().ManagedClusters().Get(b.ctx, clusterName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		return b.composedClusterStore.Update(cluster)
	})

	// start watching clusterdeployments
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).List(b.ctx, options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return dynamicClient.Resource(gvr).Namespace(metav1.NamespaceAll).Watch(b.ctx, options)
		},
	}
	reflector := cache.NewReflector(lw, &unstructured.Unstructured{}, b.clusterHibernatingStateCache, ResyncPeriod)

	klog.Infof("Start watching ClusterDeployments")
	go reflector.Run(b.ctx.Done())
}

func getHubClusterID(ocpClient ocpclient.Interface, kubeClient kubernetes.Interface) string {
	cv, err := ocpClient.ConfigV1().ClusterVersions().Get(context.TODO(), "version", metav1.GetOptions{})
	if err == nil {
		return string(cv.Spec.ClusterID)
	}

	if errors.IsNotFound(err) {
		ns, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), "kube-system", metav1.GetOptions{})
		if err != nil {
			klog.Fatalf("Error getting namespace %v \n", err)
		}
		return string(ns.GetUID())
	}

	klog.Fatalf("Error getting cluster version %v \n,", err)
	return ""
}
