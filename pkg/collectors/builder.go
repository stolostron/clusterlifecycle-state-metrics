// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"sort"
	"strings"

	ocpclient "github.com/openshift/client-go/config/clientset/versioned"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"

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
	return b
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
		klog.Fatalf("cannot create config: %v", err)
	}
	clusterclient, err := clusterclient.NewForConfig(config)
	if err != nil {
		klog.Fatalf("cannot create clusterclient: %v", err)
	}
	ocpclient, err := ocpclient.NewForConfig(config)
	if err != nil {
		klog.Fatalf("cannot create ocpclient: %v", err)
	}
	return b.buildManagedClusterInfoCollectorWithClient(ocpclient, clusterclient)
}

func (b *Builder) buildManagedClusterInfoCollectorWithClient(ocpclient *ocpclient.Clientset, clusterclient *clusterclient.Clientset) *metricsstore.MetricsStore {
	hubClusterID := getHubClusterID(ocpclient)
	filteredMetricFamilies := metric.FilterMetricFamilies(b.whiteBlackList,
		getManagedClusterInfoMetricFamilies(hubClusterID, clusterclient))
	composedMetricGenFuncs := metric.ComposeMetricGenFuncs(filteredMetricFamilies)

	familyHeaders := metric.ExtractMetricFamilyHeaders(filteredMetricFamilies)

	store := metricsstore.NewMetricsStore(
		familyHeaders,
		composedMetricGenFuncs,
	)

	createManagedClusterInformer(b.apiserver, b.kubeconfig, store)

	return store
}
