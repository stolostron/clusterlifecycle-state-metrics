// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"testing"

	mciv1beta1 "github.com/stolostron/cluster-lifecycle-api/clusterinfo/v1beta1"
	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterLabelMetricFamilies(t *testing.T) {
	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID:   "managed_cluster_id",
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorAKS),
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity:      mcv1.ResourceList{},
			ClusterClaims: []mcv1.ManagedClusterClaim{},
			Conditions:    []metav1.Condition{},
		},
	}

	mc2 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster-2",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID:   "managed_cluster_id",
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOpenShift),
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity:      mcv1.ResourceList{},
			ClusterClaims: []mcv1.ManagedClusterClaim{},
			Conditions:    []metav1.Condition{},
		},
	}

	mc3 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-cluster-3",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "managed_cluster_id",
				"5g-dev01-cluster":        "value1",
				"55g//dev01__cluster":     "value2",
				"_5g-dev01_cluster":       "value3",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity:      mcv1.ResourceList{},
			ClusterClaims: []mcv1.ManagedClusterClaim{},
			Conditions:    []metav1.Condition{},
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test cluster label",
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_labels"},
			Want:        `acm_managed_cluster_labels{cloud="Amazon",managed_cluster_id="managed_cluster_id",hub_cluster_id="hub_cluster_id",vendor="AKS"} 1`,
		},
		{
			Name:        "test cluster2 label",
			Obj:         mc2,
			MetricNames: []string{"acm_managed_cluster_labels"},
			Want:        `acm_managed_cluster_labels{cloud="Amazon",managed_cluster_id="managed_cluster_id",hub_cluster_id="hub_cluster_id",vendor="OpenShift"} 1`,
		},
		{
			Name:        "test cluster3 label",
			Obj:         mc3,
			MetricNames: []string{"acm_managed_cluster_labels"},
			Want:        `acm_managed_cluster_labels{managed_cluster_id="managed_cluster_id",hub_cluster_id="hub_cluster_id",_5g_dev01_cluster="value1",_55g_dev01__cluster="value2",_5g_dev01_cluster="value3"} 1`,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManagedClusterLabelMetricFamilies("hub_cluster_id")},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
