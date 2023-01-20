// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"testing"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"
)

func Test_getManifestWorkStatusMetricFamilies(t *testing.T) {
	work := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "local-cluster-klusterlet",
			Namespace: "local-cluster",
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				testcommon.NewCondition("Available", metav1.ConditionTrue),
				testcommon.NewCondition("Applied", metav1.ConditionTrue),
			},
		},
	}

	workWithoutCondition := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello-work",
			Namespace: "cluster1",
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test work status",
			Obj:         work,
			MetricNames: []string{"acm_manifestwork_status_condition"},
			Want: `acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="true"} 1
acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="false"} 0
acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="unknown"} 0
acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Applied",status="true"} 1
acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="local-cluster-klusterlet",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Applied",status="unknown"} 0`,
		},
		{
			Name:        "test work status without condition",
			Obj:         workWithoutCondition,
			MetricNames: []string{"acm_manifestwork_status_condition"},
			Want: `acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Applied",status="true"} 0
acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Applied",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="true"} 0
acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="false"} 0
acm_manifestwork_status_condition{manifestwork="hello-work",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="unknown"} 1`,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManifestWorkStatusMetricFamilies(func(clusterName string) string {
					return clusterName
				})},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
