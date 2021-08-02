// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
	"k8s.io/kube-state-metrics/pkg/options"

	"golang.org/x/net/context"
	"k8s.io/klog/v2"
)

type whiteBlackLister interface {
	IsIncluded(string) bool
	IsExcluded(string) bool
}

// Builder helps to build collectors. It follows the builder pattern
// (https://en.wikipedia.org/wiki/Builder_pattern).
type Builder struct {
	apiserver         string
	kubeconfig        string
	namespaces        options.NamespaceList
	ctx               context.Context
	enabledCollectors []string
	whiteBlackList    whiteBlackLister
}

// NewBuilder returns a new builder.
func NewBuilder(
	ctx context.Context,
) *Builder {
	return &Builder{
		ctx: ctx,
	}
}

func (b *Builder) WithApiserver(apiserver string) *Builder {
	b.apiserver = apiserver
	return bb
}

func (b *Builder) WithKubeConfig(kubeconfig string) *Builder {
	b.kubeconfig = kubeconfig
	return b
}

// WithEnabledCollectors sets the enabledCollectors property of a Builder.
func (b *Builder) WithEnabledCollectors(c []string) *Builder {
	copy := []string{}
	for _, s := range c {
		copy = append(copy, s)
	}

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
func (b *Builder) Build() []*metricsstore.MetricsStore {
	if b.whiteBlackList == nil {
		panic("whiteBlackList should not be nil")
	}

	collectors := []*metricsstore.MetricsStore{}
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

	return collectors
}

var availableCollectors = map[string]func(f *Builder) *metricsstore.MetricsStore{
	"managedclusterinfos": func(b *Builder) *metricsstore.MetricsStore { return b.buildManagedClusterInfoCollector() },
}

func (b *Builder) buildManagedClusterInfoCollector() *metricsstore.MetricsStore {
	config, err := clientcmd.BuildConfigFromFlags(b.apiserver, b.kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)
	return b.buildManagedClusterInfoCollectorWithClient(client)
}

func (b *Builder) buildManagedClusterInfoCollectorWithClient(client dynamic.Interface) *metricsstore.MetricsStore {
	hubClusterID := getHubClusterID(client)
	filteredMetricFamilies := metric.FilterMetricFamilies(b.whiteBlackList,
		getManagedClusterInfoMetricFamilies(hubClusterID, client))
	composedMetricGenFuncs := metric.ComposeMetricGenFuncs(filteredMetricFamilies)

	familyHeaders := metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)

	store := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)
	reflectorPerNamespace(b.ctx, &unstructured.Unstructured{}, store,
		b.apiserver, b.kubeconfig, b.namespaces, createManagedClusterInfoListWatch)
	reflectorClusterScoped(b.ctx, &unstructured.Unstructured{}, store,
		b.apiserver, b.kubeconfig, createManagedClusterListWatch)

	return store
}

// reflectorPerNamespace creates a Kubernetes client-go reflector with the given
// listWatchFunc for each given namespace and registers it with the given store.
func reflectorPerNamespace(
	ctx context.Context,
	expectedType interface{},
	store cache.Store,
	apiserver string,
	kubeconfig string,
	namespaces []string,
	listWatchFunc func(apiserver string, kubeconfig string, ns string) cache.ListWatch,
) {
	for _, ns := range namespaces {
		lw := listWatchFunc(apiserver, kubeconfig, ns)
		reflector := cache.NewReflector(&lw, expectedType, store, 0)
		go reflector.Run(ctx.Done())
	}
}

// reflectorClusterScoped creates a Kubernetes client-go reflectorClusterScoped with the given
// listWatchFunc for each given namespace and registers it with the given store.
func reflectorClusterScoped(
	ctx context.Context,
	expectedType interface{},
	store cache.Store,
	apiserver string,
	kubeconfig string,
	listWatchFunc func(apiserver string, kubeconfig string) cache.ListWatch,
) {
	lw := listWatchFunc(apiserver, kubeconfig)
	reflector := cache.NewReflector(&lw, expectedType, store, 0)
	go reflector.Run(ctx.Done())
}
