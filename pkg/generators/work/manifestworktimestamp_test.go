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
	t3, err := time.Parse(time.RFC3339, "2021-09-01T00:01:03Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t4, err := time.Parse(time.RFC3339, "2021-09-01T00:01:04Z")
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

	sdwork := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster2-hosted-klusterlet",
			Namespace:         "local-cluster",
			CreationTimestamp: metav1.Time{Time: t1},
			Labels: map[string]string{
				clusterServiceManagementClusterLabel: "local-cluster",
			},
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				testcommon.NewConditionWithTime("Applied", metav1.ConditionTrue, t2),
			},
		},
	}

	sdreservedwork := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster2-hosted-klusterlet",
			Namespace:         "local-cluster",
			CreationTimestamp: metav1.Time{Time: t1},
			Labels: map[string]string{
				hostedClusterLabel: "cluster2",
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

	generationTimeValue1 := `{"generation":1,"createdTime":"2021-09-01T00:01:01Z","appliedTime":"2021-09-01T00:01:02.799199048Z","firstGenerationAppliedTime":"2021-09-01T00:01:02.799199048Z"}`
	workWithGenerationTime1 := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster2-hosted-klusterlet",
			Namespace:         "local-cluster",
			CreationTimestamp: metav1.Time{Time: t1},
			Labels: map[string]string{
				importHostedClusterLabel: "cluster2",
			},
			Annotations: map[string]string{
				annotationKeyGenerationTime: generationTimeValue1,
			},
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				testcommon.NewConditionWithTime("Applied", metav1.ConditionTrue, t2),
			},
		},
	}
	generationTimeValue2 := `{"generation":2,"createdTime":"2021-09-01T00:01:03Z","appliedTime":"2021-09-01T00:01:04Z","firstGenerationAppliedTime":"2021-09-01T00:01:02Z"}`
	workWithGenerationTime2 := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "cluster2-hosted-klusterlet",
			Namespace:         "local-cluster",
			CreationTimestamp: metav1.Time{Time: t1},
			Labels: map[string]string{
				importHostedClusterLabel: "cluster2",
			},
			Annotations: map[string]string{
				annotationKeyGenerationTime: generationTimeValue2,
			},
		},
		Status: workv1.ManifestWorkStatus{
			Conditions: []metav1.Condition{
				testcommon.NewConditionWithTime("Applied", metav1.ConditionTrue, t2),
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
			Name:        "test sd work status",
			Obj:         sdwork,
			MetricNames: []string{"acm_manifestwork_apply_timestamp"},
			Want: fmt.Sprintf(`acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",managed_cluster_id="local-cluster",status="Created"} %.9e
acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",managed_cluster_id="local-cluster",status="Applied"} %.9e`, float64(t1.Unix()), float64(t2.Unix())),
		},
		{
			Name:        "test sd reserved work status",
			Obj:         sdreservedwork,
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
		{
			Name:        "test work with generation time annotation generation 1",
			Obj:         workWithGenerationTime1,
			MetricNames: []string{"acm_manifestwork_apply_timestamp"},
			Want: fmt.Sprintf(`acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Created",generation="1"} %.9e
			acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Applied",generation="1"} %.9e`, float64(t1.Unix()), float64(t2.Unix())),
		},
		{
			Name:        "test work with generation time annotation generation 2",
			Obj:         workWithGenerationTime2,
			MetricNames: []string{"acm_manifestwork_apply_timestamp"},
			Want: fmt.Sprintf(`acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Created",generation="2"} %.9e
			acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Applied",generation="2"} %.9e
			acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Created",generation="1"} %.9e
			acm_manifestwork_apply_timestamp{manifestwork="cluster2-hosted-klusterlet",managed_cluster_name="local-cluster",hosted_cluster_name="cluster2",managed_cluster_id="local-cluster",status="Applied",generation="1"} %.9e`, float64(t3.Unix()), float64(t4.Unix()), float64(t1.Unix()), float64(t2.Unix())),
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
