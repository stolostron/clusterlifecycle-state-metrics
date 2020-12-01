package collectors

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getClusterDeploymentMetricFamilies(t *testing.T) {

	cd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"name":      "hive-cluster",
				"namespace": "hive-cluster",
			},
		},
	}

	cd.SetCreationTimestamp(metav1.Time{Time: time.Unix(1500000000, 0)})
	tests := []generateMetricsTestCase{
		{
			Obj:         cd,
			MetricNames: []string{"ocm_clusterdeployment_created"},
			Want: `
			ocm_clusterdeployment_created{hub_cluster_id="",name="hive-cluster",namespace="hive-cluster"} 1
				`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(clusterDeploymentrMetricFamilies)
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
