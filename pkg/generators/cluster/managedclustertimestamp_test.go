// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"fmt"
	"testing"
	"time"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterTimestampMetricFamilies(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2021-09-01T00:01:01Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t2, err := time.Parse(time.RFC3339, "2021-09-01T00:01:02Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster1",
			CreationTimestamp: metav1.Time{Time: t1},
			Annotations: map[string]string{
				"import.open-cluster-management.io/hosting-cluster-name": "cluster2",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Conditions: []metav1.Condition{
				testcommon.NewConditionWithTime("ManagedClusterJoined", metav1.ConditionTrue, t2),
			},
		},
	}

	mcWithoutCondition := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster1",
			CreationTimestamp: metav1.Time{Time: t1},
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test cluster status",
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_import_timestamp"},
			Want: fmt.Sprintf(`acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",hosting_cluster_name="cluster2",status="Created"} %.9e
acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",hosting_cluster_name="cluster2",status="Joined"} %.9e
acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",hosting_cluster_name="cluster2",status="ManagedClusterKubeconfigProvided"} %.9e
acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",hosting_cluster_name="cluster2",status="StartToApplyKlusterletResources"} %.9e`,
				float64(t1.Unix()), float64(t2.Unix()), float64(t1.Unix()), float64(t2.Unix())),
		},
		{
			Name:        "test cluster status without condition",
			Obj:         mcWithoutCondition,
			MetricNames: []string{"acm_managed_cluster_import_timestamp"},
			Want: fmt.Sprintf(`acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",status="Created"} %.9e
			acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",status="ManagedClusterKubeconfigProvided"} %.9e
			acm_managed_cluster_import_timestamp{managed_cluster_name="cluster1",managed_cluster_id="cluster1",status="StartToApplyKlusterletResources"} %.9e`,
				float64(t1.Unix()), float64(t1.Unix()), float64(t2.Unix())),
		},
	}

	hubID := "hub"
	clusterTimestampCache := map[string]map[string]float64{
		"cluster1": {
			"ManagedClusterKubeconfigProvided": float64(t1.Unix()),
			"StartToApplyKlusterletResources":  float64(t2.Unix()),
		},
	}
	getClusterTimestamps := func(clusterName string) map[string]float64 {
		return clusterTimestampCache[clusterName]
	}
	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManagedClusterTimestampMetricFamilies(hubID, getClusterTimestamps)},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
