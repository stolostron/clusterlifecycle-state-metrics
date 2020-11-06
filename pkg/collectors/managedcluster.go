package collectors

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"

	clientset "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
)

var (
	descRouteLabelsName          = "ocm_managedcluster_labels"
	descRouteLabelsHelp          = "Kubernetes labels converted to Prometheus labels."
	descRouteLabelsDefaultLabels = []string{"managedcluster"}

	managedClusterMetricFamilies = []metric.FamilyGenerator{
		{
			Name: "openshift_route_created",
			Type: metric.MetricTypeGauge,
			Help: "Unix creation timestamp",
			GenerateFunc: wrapManagedClusterFunc(func(c *managedclusterv1.ManagedCluster) metric.Family {
				f := metric.Family{}

				if !c.CreationTimestamp.IsZero() {
					f.Metrics = append(f.Metrics, &metric.Metric{
						Value: float64(c.CreationTimestamp.Unix()),
					})
				}

				return f
			}),
		},
		{
			Name: descRouteLabelsName,
			Type: metric.MetricTypeGauge,
			Help: descRouteLabelsHelp,
			GenerateFunc: wrapManagedClusterFunc(func(d *managedclusterv1.ManagedCluster) metric.Family {
				labelKeys, labelValues := kubeLabelsToPrometheusLabels(d.Labels)
				return metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   labelKeys,
						LabelValues: labelValues,
						Value:       1,
					},
				}}
			}),
		},
	}
)

func wrapManagedClusterFunc(f func(*managedclusterv1.ManagedCluster) metric.Family) func(interface{}) metric.Family {
	return func(obj interface{}) metric.Family {
		Cluster := obj.(*managedclusterv1.ManagedCluster)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descRouteLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{Cluster.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createManagedClusterListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	managedclusterclient, err := createManagedClusterClient(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Route client: %v", err)
	}
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return managedclusterclient.ClusterV1().ManagedClusters().List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return managedclusterclient.ClusterV1().ManagedClusters().Watch(context.TODO(), opts)
		},
	}
}

func createManagedClusterClient(apiserver string, kubeconfig string) (*clientset.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	return client, err

}
