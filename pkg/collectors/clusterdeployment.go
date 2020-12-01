package collectors

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"
)

var (
	descClusterDeploymentName                = "ocm_clusterdeployment_created"
	descClusterDeploymentHelp                = "Hive Cluster deployment"
	descClusterDeploymentLabelsDefaultLabels = []string{"hub_cluster_id", "namespace", "name"}
)

func getClusterDeploymentMetricFamilies(hubClusterID string) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: descClusterDeploymentName,
			Type: metric.MetricTypeGauge,
			Help: descClusterDeploymentHelp,
			GenerateFunc: wrapClusterDeploymentFunc(func(c *unstructured.Unstructured) metric.Family {
				labelsValues := []string{hubClusterID,
					c.GetName(),
					c.GetName()}
				return metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   descClusterDeploymentLabelsDefaultLabels,
						LabelValues: labelsValues,
						Value:       1,
					},
				}}
			}),
		},
	}
}

func wrapClusterDeploymentFunc(f func(*unstructured.Unstructured) metric.Family) func(interface{}) metric.Family {
	return func(obj interface{}) metric.Family {
		Cluster := obj.(*unstructured.Unstructured)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createClusterDeploymentListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return client.Resource(cdGVR).Namespace(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return client.Resource(cdGVR).Namespace(ns).Watch(context.TODO(), opts)
		},
	}
}
