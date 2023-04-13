// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"testing"

	mciv1beta1 "github.com/stolostron/cluster-lifecycle-api/clusterinfo/v1beta1"
	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterMetricFamilies(t *testing.T) {
	hubType := "mce"

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
			Annotations: map[string]string{
				"open-cluster-management/service-name": "compute",
			},
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

	mcWithoutClusterId := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "hive",
			},
			Labels: map[string]string{
				mciv1beta1.OCPVersion:       "4.3.1",
				mciv1beta1.LabelKubeVendor:  string(mciv1beta1.KubeVendorOpenShift),
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
				mciv1beta1.LabelClusterID: "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(0, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(0, resource.DecimalSI),
			},
			Conditions: []metav1.Condition{},
		},
	}

	mcZeroInfo := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster-3",
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
				resourceCoreWorker:   *resource.NewQuantity(0, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(0, resource.DecimalSI),
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

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "cluster info",
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",service_name="Other",created_via="Hive",hub_cluster_id="mycluster_id",hub_type="mce",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Name:        "cluster info discovery",
			Obj:         mcDiscovery,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",service_name="Other",created_via="Discovery",hub_cluster_id="mycluster_id",hub_type="mce",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Name:        "no cluster id",
			Obj:         mcWithoutClusterId,
			MetricNames: []string{"acm_managed_cluster_info"},
		},
		{
			Name:        "missing info",
			Obj:         mcMissingInfo,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="",core_worker="0",managed_cluster_id="managed_cluster_id",service_name="Other",created_via="Other",hub_cluster_id="mycluster_id",hub_type="mce",socket_worker="0",available="Unknown",vendor="",version=""} 1`,
		},
		{
			Name:        "zero resource",
			Obj:         mcZeroInfo,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="0",managed_cluster_id="managed_cluster_id",service_name="Other",created_via="Hive",hub_cluster_id="mycluster_id",hub_type="mce",socket_worker="0",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Name:        "others resource",
			Obj:         mcOther,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="cluster-other",service_name="Compute",created_via="Other",hub_cluster_id="mycluster_id",hub_type="mce",socket_worker="2",available="Unknown",vendor="Other",version="v1.16.2"} 1`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(
			[]metric.FamilyGenerator{GetManagedClusterInfoMetricFamilies("mycluster_id", hubType)},
		)
		if err := c.Run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}
	}
}
