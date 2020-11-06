package collectors

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"

	clientset "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
)

var (
	descClusterLabelsName          = "ocm_managedcluster_labels"
	descClusterLabelsHelp          = "Kubernetes labels converted to Prometheus labels."
	descClusterLabelsDefaultLabels = []string{"managedcluster"}

	cdGVR = schema.GroupVersionResource{
		Group:    "hive.openshift.io",
		Version:  "v1",
		Resource: "clusterdeployments",
	}
)

func getManagedClusterrMetricFamilies(client dynamic.Interface) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: "ocm_cluster_created",
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
			Name: descClusterLabelsName,
			Type: metric.MetricTypeGauge,
			Help: descClusterLabelsHelp,
			GenerateFunc: wrapManagedClusterFunc(func(d *managedclusterv1.ManagedCluster) metric.Family {
				labelKeys, labelValues := kubeLabelsToPrometheusLabels(d.Labels)
				createdVia := "hive"
				_, err := client.Resource(cdGVR).Namespace(d.GetName()).Get(context.TODO(), d.GetName(), metav1.GetOptions{})
				if errors.IsNotFound(err) {
					createdVia = "imported"
				}
				labelKeys = append(labelKeys, "created_via")
				labelValues = append(labelValues, createdVia)
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
}

func wrapManagedClusterFunc(f func(*managedclusterv1.ManagedCluster) metric.Family) func(interface{}) metric.Family {
	return func(obj interface{}) metric.Family {
		Cluster := obj.(*managedclusterv1.ManagedCluster)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append(descClusterLabelsDefaultLabels, m.LabelKeys...)
			m.LabelValues = append([]string{Cluster.Name}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createManagedClusterListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	managedclusterclient, err := createManagedClusterClient(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create ManagedCluster client: %v", err)
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
