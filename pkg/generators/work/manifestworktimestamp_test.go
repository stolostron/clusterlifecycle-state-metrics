// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"fmt"
	"testing"
	"time"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"
)

func Test_getManifestWorkTimestampMetricFamilies(t *testing.T) {

	t1, err := time.Parse(time.RFC3339, "2021-09-01T00:01:01Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t2, err := time.Parse(time.RFC3339, "2021-09-01T00:01:02Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	work := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster2-hosted-klusterlet",
			Namespace:         "local-cluster",
			CreationTimestamp: metav1.Time{Time: t1},
			Labels: map[string]string{
				importHostedClusterLabel: "cluster2",
			},
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				testcommon.NewConditionWithTime("Applied", metav1.ConditionTrue, t2),
			},
		},
	}

	workWithoutCondition := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster2-hosted-klusterlet",
			Namespace: "local-cluster",
			CreationTimestamp: metav1.Time{
				Time: t1,
			},
			Labels: map[string]string{
				importHostedClusterLabel: "cluster2",
			},
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test work status",
			Obj:         work,
			MetricNames: []string{"acm_manifestwork_apply_timestamp"},
			Want: fmt.Sprintf(`acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Created"} %.9e
acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Applied"} %.9e`, float64(t1.Unix()), float64(t2.Unix())),
		},
		{
			Name:        "test work status without condition",
			Obj:         workWithoutCondition,
			MetricNames: []string{"acm_manifestwork_apply_timestamp"},
			Want:        fmt.Sprintf(`acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Created"} %.9e`, float64(t1.Unix())),
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManifestWorkTimestampMetricFamilies(func(clusterName string) string {
					return clusterName
				})},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
