// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	clusterclient "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	mcv1 "github.com/open-cluster-management/api/cluster/v1"
	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getManagedClusterMetricFamilies(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(mciv1beta1.GroupVersion, &mciv1beta1.ManagedClusterInfo{})
	s.AddKnownTypes(mcv1.GroupVersion, &mcv1.ManagedCluster{})

	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "hive",
			},
			Labels: map[string]string{
				OCPVersion:       "4.3.1",
				LabelKubeVendor:  "OpenShift",
				LabelCloudVendor: "Amazon",
				LabelClusterID:   "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
		},
	}

	mcDiscovery := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "discovery-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "discovery",
			},
			Labels: map[string]string{
				OCPVersion:       "4.3.1",
				LabelKubeVendor:  "OpenShift",
				LabelCloudVendor: "Amazon",
				LabelClusterID:   "managed_cluster_id",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
		},
	}

	mcOther := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-other",
			Labels: map[string]string{
				LabelKubeVendor:  "OpenShift",
				LabelCloudVendor: "Amazon",
				LabelClusterID:   "managed_cluster_id",
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
		},
	}

	mcMissingInfo := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster-2",
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(3, resource.DecimalSI),
			},
		},
	}

	envTest, _, _, _ := setupEnvTest(t)
	clusterClient, _ := clusterclient.NewForConfig(envTest.Config)
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
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", clusterClient))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}
	}
}
