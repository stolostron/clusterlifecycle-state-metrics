// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	mciv1beta1 "github.com/stolostron/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterMetricFamilies(t *testing.T) {
	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "hive",
			},
			Labels: map[string]string{
				mciv1beta1.OCPVersion:       "4.3.1",
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOpenShift),
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
				mciv1beta1.LabelClusterID:   "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
			ClusterClaims: []mcv1.ManagedClusterClaim{
				{
					Name:  "kubeversion.open-cluster-management.io",
					Value: "v1.16.2",
				},
			},
			Conditions: []metav1.Condition{},
		},
	}

	mcDiscovery := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "discovery-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "discovery",
			},
			Labels: map[string]string{
				mciv1beta1.OCPVersion:       "4.3.1",
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOpenShift),
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
				mciv1beta1.LabelClusterID:   "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
			ClusterClaims: []mcv1.ManagedClusterClaim{
				{
					Name:  "kubeversion.open-cluster-management.io",
					Value: "v1.16.2",
				},
			},
			Conditions: []metav1.Condition{},
		},
	}

	mcOther := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-other",
			Labels: map[string]string{
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOther),
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
			ClusterClaims: []mcv1.ManagedClusterClaim{
				{
					Name:  "kubeversion.open-cluster-management.io",
					Value: "v1.16.2",
				},
			},
			Conditions: []metav1.Condition{},
		},
	}

	mcMissingInfo := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster-2",
			Labels: map[string]string{
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOther),
				mciv1beta1.LabelCloudVendor: string(mciv1beta1.CloudVendorAWS),
				mciv1beta1.LabelClusterID:   "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(3, resource.DecimalSI),
			},
			Conditions: []metav1.Condition{},
		},
	}

	envTest, _, _, _ := setupEnvTest(t)
	clusterClient, err := clusterclient.NewForConfig(envTest.Config)
	if err != nil {
		t.Errorf("Failed to create clusterclient: %s", err)
	}

	tests := []generateMetricsTestCase{
		{
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Hive",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Obj:         mcDiscovery,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Discovery",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Obj:         mcMissingInfo,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        "",
		},
		{
			Obj:         mcOther,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="cluster-other",created_via="Other",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="Other",version="v1.16.2"} 1`,
		},
	}
	for i, c := range tests {
		mc, err := clusterClient.ClusterV1().ManagedClusters().Create(context.Background(), c.Obj, metav1.CreateOptions{})
		if err != nil {
			t.Errorf("failed to generate managedcluster CR: %s\ns", err)
		}
		mc.Status = c.Obj.DeepCopy().Status
		_, err = clusterClient.ClusterV1().ManagedClusters().UpdateStatus(context.Background(), mc, metav1.UpdateOptions{})
		if err != nil {
			t.Errorf("failed to update managedcluster statue: %s\ns", err)
		}
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", clusterClient))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}
		clusterClient.ClusterV1().ManagedClusters().Delete(context.Background(), c.Obj.Name, metav1.DeleteOptions{})
	}
}
