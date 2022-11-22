// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	mciv1beta1 "github.com/stolostron/cluster-lifecycle-api/clusterinfo/v1beta1"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned/fake"
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

	clusterClient := clusterclient.NewSimpleClientset()
	tests := []generateMetricsTestCase{
		{
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_labels"},
			Want:        `acm_managed_cluster_labels{cloud="Amazon",managed_cluster_id="managed_cluster_id",hub_cluster_id="hub_cluster_id",vendor="AKS"} 1`,
		},
		{
			Obj:         mc2,
			MetricNames: []string{"acm_managed_cluster_labels"},
			Want:        `acm_managed_cluster_labels{cloud="Amazon",managed_cluster_id="managed_cluster_id",hub_cluster_id="hub_cluster_id",vendor="OpenShift"} 1`,
		},
	}

	for i, c := range tests {
		_, err := clusterClient.ClusterV1().ManagedClusters().Create(context.Background(), c.Obj, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("failed to generate managedcluster CR: %s\ns", err)
		}

		c.Func = metric.ComposeMetricGenFuncs(
			getManagedClusterLabelMetricFamilies("hub_cluster_id", clusterClient),
		)
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}

		e := clusterClient.ClusterV1().ManagedClusters().Delete(context.Background(), c.Obj.Name, metav1.DeleteOptions{})
		if e != nil {
			t.Errorf("failed to delete cluster %v: %v", c.Obj.Name, e)
		}

		if err = c.run(); err == nil {
			t.Errorf("failed to trigger error response for cluster %v: %v", c.Obj.Name, err)
		}
	}
}
