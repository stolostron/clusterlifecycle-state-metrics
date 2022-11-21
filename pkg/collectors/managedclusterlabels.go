package collectors

import (
	"context"
	"regexp"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"

	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descManagedClusterLabelInfoName     = "managed_cluster_labels"
	descManagedClusterLabelInfoHelp     = "Managed cluster labels"
	descManagedClusterLabelDefaultLabel = []string{
		"hub_cluster_id",
		"managed_cluster_id",
	}
)

func getManagedClusterLabelMetricFamilies(hubClusterID string, clusterclient *clusterclient.Clientset) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: descManagedClusterLabelInfoName,
			Type: metric.Gauge,
			Help: descManagedClusterLabelInfoHelp,
			GenerateFunc: wrapManagedClusterLabelFunc(func(obj *mcv1.ManagedCluster) metric.Family {
				klog.Infof("Wrap %s", obj.GetName())
				mc, err := clusterclient.ClusterV1().ManagedClusters().Get(context.Background(), obj.GetName(),
					metav1.GetOptions{})

				if err != nil {
					klog.Warningf("Failed to get managedcluster resource %s: %v", obj.GetName(), err)
					return metric.Family{Metrics: []*metric.Metric{}}
				}

				mangedClusterID := getClusterID(mc)

				labelsKeys := descManagedClusterLabelDefaultLabel
				labelsValues := []string{
					hubClusterID,
					mangedClusterID,
				}

				regex := regexp.MustCompile(`[^\w]+`)
				for key, value := range mc.Labels {
					// Ignore the clusterID label since it is being set within the hub and managed cluster IDs
					if key != "clusterID" {
						labelsKeys = append(labelsKeys, regex.ReplaceAllString(key, "_"))
						labelsValues = append(labelsValues, value)
					}
				}

				f := metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   labelsKeys,
						LabelValues: labelsValues,
						Value:       1,
					},
				}}

				klog.Infof("Returning %v", string(f.ByteSlice()))
				return f
			}),
		},
	}
}

func wrapManagedClusterLabelFunc(f func(obj *mcv1.ManagedCluster) metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		cluster := obj.(*mcv1.ManagedCluster)

		metricFamily := f(cluster)
		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return &metricFamily
	}
}
