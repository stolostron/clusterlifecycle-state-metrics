// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"testing"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterStatusMetricFamilies(t *testing.T) {
	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
		},
		Status: mcv1.ManagedClusterStatus{
			Conditions: []metav1.Condition{
				testcommon.NewCondition("HubAcceptedManagedCluster", metav1.ConditionTrue),
				testcommon.NewCondition("ExternalManagedKubeconfigCreatedSucceeded", metav1.ConditionTrue),
				testcommon.NewCondition("ManagedClusterJoined", metav1.ConditionTrue),
				testcommon.NewCondition("ManagedClusterConditionAvailable", metav1.ConditionTrue),
			},
		},
	}

	mcWithoutCondition := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test cluster status",
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_status_condition"},
			Want: `acm_managed_cluster_status_condition{managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="HubAcceptedManagedCluster",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="ExternalManagedKubeconfigCreatedSucceeded",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="ManagedClusterJoined",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="ManagedClusterConditionAvailable",status="true"} 1`,
		},
		{
			Name:        "test cluster status without condition",
			Obj:         mcWithoutCondition,
			MetricNames: []string{"acm_managed_cluster_status_condition"},
			Want:        `acm_managed_cluster_status_condition{managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="ManagedClusterConditionAvailable",status="unknown"} 1`,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManagedClusterStatusMetricFamilies()},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
