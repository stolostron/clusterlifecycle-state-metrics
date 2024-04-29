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

func Test_getManagedClusterWorkerCoresMetricFamilies(t *testing.T) {
	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker: *resource.NewQuantity(4, resource.DecimalSI),
			},
			Conditions: []metav1.Condition{},
		},
	}

	mcWithZeroCapacity := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster2",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker: *resource.NewQuantity(0, resource.DecimalSI),
			},
		},
	}

	mcWithoutCapacity := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster3",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "managed_cluster_id",
			},
		},
	}

	ocpWithoutClusterId := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster4",
			Labels: map[string]string{
				mciv1beta1.OCPVersion:      "4.3.1",
				mciv1beta1.LabelKubeVendor: string(mciv1beta1.KubeVendorOpenShift),
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker: *resource.NewQuantity(4, resource.DecimalSI),
			},
		},
	}

	nonOCPWithoutClusterId := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster5",
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker: *resource.NewQuantity(4, resource.DecimalSI),
			},
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "cluster with core_worker",
			Obj:         mc,
			MetricNames: []string{"acm_managed_cluster_worker_cores"},
			Want:        `acm_managed_cluster_worker_cores{managed_cluster_id="managed_cluster_id",hub_cluster_id="mycluster_id"} 4`,
		},
		{
			Name:        "cluster with zero core_worker",
			Obj:         mcWithZeroCapacity,
			MetricNames: []string{"acm_managed_cluster_worker_cores"},
			Want:        `acm_managed_cluster_worker_cores{managed_cluster_id="managed_cluster_id",hub_cluster_id="mycluster_id"} 0`,
		},
		{
			Name:        "cluster without core_worker",
			Obj:         mcWithoutCapacity,
			MetricNames: []string{"acm_managed_cluster_worker_cores"},
			Want:        `acm_managed_cluster_worker_cores{managed_cluster_id="managed_cluster_id",hub_cluster_id="mycluster_id"} 0`,
		},
		{
			Name:        "ocp cluster without cluster id",
			Obj:         ocpWithoutClusterId,
			MetricNames: []string{"acm_managed_cluster_worker_cores"},
		},
		{
			Name:        "non-ocp cluster",
			Obj:         nonOCPWithoutClusterId,
			MetricNames: []string{"acm_managed_cluster_worker_cores"},
			Want:        `acm_managed_cluster_worker_cores{managed_cluster_id="cluster5",hub_cluster_id="mycluster_id"} 4`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(
			[]metric.FamilyGenerator{GetManagedClusterWorkerCoresMetricFamilies("mycluster_id")},
		)
		if err := c.Run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}
	}
}
