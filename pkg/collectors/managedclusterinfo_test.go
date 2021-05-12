// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"reflect"
	"testing"

	mcv1 "github.com/open-cluster-management/api/cluster/v1"
	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/metric"
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

	cdU := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "ClusterDeployment",
			"apiVersion": "hive.openshift.io/v1",
			"metadata": map[string]interface{}{
				"name":      "hive-cluster",
				"namespace": "hive-cluster",
			},
		},
	}

	client := fake.NewSimpleDynamicClient(s, mciU, mciUMissingInfo, mciUOther, mcU, mcUOther, mcUMissingInfo)
	clientHive := fake.NewSimpleDynamicClient(s, mciU, cdU, mcU, mcUOther, mcUMissingInfo)
	tests := []generateMetricsTestCase{
		{
			Obj:         mciU,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Other",hub_cluster_id="mycluster_id",socket_worker="2",status="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
		{
			Obj:         mciUMissingInfo,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        "",
		},
		{
			Obj:         mciUOther,
			MetricNames: []string{"acm_managed_cluster_info"},
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="cluster-other",created_via="Other",hub_cluster_id="mycluster_id",socket_worker="2",status="Unknown",vendor="Other",version="v1.16.2"} 1`,
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
			Want:        `acm_managed_cluster_info{cloud="Amazon",core_worker="4",managed_cluster_id="managed_cluster_id",created_via="Hive",hub_cluster_id="mycluster_id",socket_worker="2",status="Unknown",vendor="OpenShift",version="4.3.1"} 1`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", clientHive))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}

func Test_createManagedClusterInfoListWatchWithClient(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(mciv1beta1.GroupVersion, &mciv1beta1.ManagedClusterInfo{})
	s.AddKnownTypes(mciv1beta1.GroupVersion, &mciv1beta1.ManagedClusterInfoList{})

	mc := &mciv1beta1.ManagedClusterInfo{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ManagedClusterInfo",
			APIVersion: "internal.open-cluster-management.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hive-cluster",
			Namespace: "hive-cluster",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOpenShift,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
			ClusterID:   "managed_cluster_id",
		},
	}
	mcU := &unstructured.Unstructured{}
	err := scheme.Scheme.Convert(mc, mcU, nil)
	if err != nil {
		t.Error(err)
	}

	client := fake.NewSimpleDynamicClient(s, mc)
	type args struct {
		client dynamic.Interface
		ns     string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name: "succeed",
			args: args{
				client: client,
				ns:     "hive-cluster",
			},
			want:    1,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createManagedClusterInfoListWatchWithClient(tt.args.client, tt.args.ns)
			l, err := got.ListFunc(metav1.ListOptions{})
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}
			lU := l.(*unstructured.UnstructuredList)

			if len(lU.Items) != tt.want {
				t.Errorf("expected a list of %d elements got %d", tt.want, len(lU.Items))
			}
			if !reflect.DeepEqual(lU.Items[0], *mcU) {
				t.Errorf("expected of %v got %v", *mcU, lU.Items[0])
			}
			w, err := got.WatchFunc(metav1.ListOptions{})
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}
			if w == nil {
				t.Errorf("expected the watch to be not nil")
			}
		})
	}
}
