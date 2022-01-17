// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	mciv1beta1 "github.com/stolostron/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_getManagedClusterMetricFamilies(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(mciv1beta1.GroupVersion, &mciv1beta1.ManagedClusterInfo{})
	s.AddKnownTypes(mcv1.GroupVersion, &mcv1.ManagedCluster{})

	mci := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hive-cluster",
			Namespace: "hive-cluster",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOpenShift,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
			ClusterID:   "managed_cluster_id",
			DistributionInfo: mciv1beta1.DistributionInfo{
				Type: mciv1beta1.DistributionTypeOCP,
				OCP: mciv1beta1.OCPDistributionInfo{
					Version: "4.3.1",
				},
			},
			NodeList: []mciv1beta1.NodeStatus{
				//Label worker no vCPU
				{
					Name: "worker-2",
					Labels: map[string]string{
						workerLabel: "",
					},
					Capacity: mciv1beta1.ResourceList{
						mciv1beta1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
					},
				},
			},
		},
	}
	mciU := &unstructured.Unstructured{}
	err := scheme.Scheme.Convert(mci, mciU, nil)
	if err != nil {
		t.Error(err)
	}

	mc := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "hive",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
		},
	}

	mcU := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mc, mcU, nil)
	if err != nil {
		t.Error(err)
	}

	mciDiscovery := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "discovery-cluster",
			Namespace: "discovery-cluster",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOpenShift,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
			ClusterID:   "managed_cluster_id",
			DistributionInfo: mciv1beta1.DistributionInfo{
				Type: mciv1beta1.DistributionTypeOCP,
				OCP: mciv1beta1.OCPDistributionInfo{
					Version: "4.3.1",
				},
			},
			NodeList: []mciv1beta1.NodeStatus{
				//Label worker no vCPU
				{
					Name: "worker-2",
					Labels: map[string]string{
						workerLabel: "",
					},
					Capacity: mciv1beta1.ResourceList{
						mciv1beta1.ResourceMemory: *resource.NewQuantity(100, resource.DecimalSI),
					},
				},
			},
		},
	}

	mciUDiscovery := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mciDiscovery, mciUDiscovery, nil)
	if err != nil {
		t.Error(err)
	}

	mcDiscovery := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "discovery-cluster",
			Annotations: map[string]string{
				"open-cluster-management/created-via": "discovery",
			},
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
		},
	}

	mcUDiscovery := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mcDiscovery, mcUDiscovery, nil)
	if err != nil {
		t.Error(err)
	}

	mciOther := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-other",
			Namespace: "cluster-other",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOther,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
			NodeList: []mciv1beta1.NodeStatus{
				// Label worker with vCPU
				{
					Name: "worker-3",
					Labels: map[string]string{
						workerLabel: "",
					},
				},
			},
		},
	}
	mciUOther := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mciOther, mciUOther, nil)
	if err != nil {
		t.Error(err)
	}

	mcOther := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-other",
		},
		Status: mcv1.ManagedClusterStatus{
			Capacity: mcv1.ResourceList{
				resourceCoreWorker:   *resource.NewQuantity(4, resource.DecimalSI),
				resourceSocketWorker: *resource.NewQuantity(2, resource.DecimalSI),
			},
		},
	}

	mcUOther := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mcOther, mcUOther, nil)
	if err != nil {
		t.Error(err)
	}

	mciMissingInfo := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hive-cluster-2",
			Namespace: "hive-cluster-2",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOpenShift,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			ClusterID:   "managed_cluster_id",
		},
	}

	mciUMissingInfo := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mciMissingInfo, mciUMissingInfo, nil)
	if err != nil {
		t.Error(err)
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

	mcUMissingInfo := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mcMissingInfo, mcUMissingInfo, nil)
	if err != nil {
		t.Error(err)
	}

	client := fake.NewSimpleDynamicClient(s, mciU, mciUDiscovery, mciUMissingInfo, mciUOther, mcU, mcDiscovery, mcUOther, mcUMissingInfo)
	clientHive := fake.NewSimpleDynamicClient(s, mciU, mciDiscovery, mcU, mcUOther, mcUMissingInfo)
	tests := []generateMetricsTestCase{
		{
			Obj:         mciU,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Hive",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Obj:         mciUDiscovery,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Discovery",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Obj:         mciUMissingInfo,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        "",
		},
		{
			Obj:         mciUOther,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="cluster-other",created_via="Other",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="Other",version="v1.16.2"} 1`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", client))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
		}
	}
	tests = []generateMetricsTestCase{
		{
			Obj:         mciU,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Hive",hub_cluster_id="mycluster_id",socket_worker="2",available="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", clientHive))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
