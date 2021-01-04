// Copyright (c) 2021 Red Hat, Inc.

package collectors

import (
	"reflect"
	"testing"

	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
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

	mc := &mciv1beta1.ManagedClusterInfo{
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
		},
	}
	mcU := &unstructured.Unstructured{}
	err := scheme.Scheme.Convert(mc, mcU, nil)
	if err != nil {
		t.Error(err)
	}

	mcOther := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-other",
			Namespace: "cluster-other",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOther,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
		},
	}
	mcUOther := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(mcOther, mcUOther, nil)
	if err != nil {
		t.Error(err)
	}

	mcMissingInfo := &mciv1beta1.ManagedClusterInfo{
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

	client := fake.NewSimpleDynamicClient(s, mcU, mcUMissingInfo, mcUOther)
	clientHive := fake.NewSimpleDynamicClient(s, mcU, cdU)
	tests := []generateMetricsTestCase{
		{
			Obj:         mcU,
			MetricNames: []string{"clc_managedcluster_info"},
			Want: `
			clc_managedcluster_info{cloud="Amazon",cluster_id="managed_cluster_id",created_via="Other",hub_cluster_id="mycluster_id",vendor="OpenShift",version="4.3.1"} 1
				`,
		},
		{
			Obj:         mcUMissingInfo,
			MetricNames: []string{"clc_managedcluster_info"},
			Want:        "",
		},
		{
			Obj:         mcUOther,
			MetricNames: []string{"clc_managedcluster_info"},
			Want: `
			clc_managedcluster_info{cloud="Amazon",cluster_id="cluster-other",created_via="Other",hub_cluster_id="mycluster_id",vendor="Other",version="v1.16.2"} 1
				`,
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
			Obj:         mcU,
			MetricNames: []string{"clc_managedcluster_info"},
			Want: `
			clc_managedcluster_info{cloud="Amazon",cluster_id="managed_cluster_id",created_via="Hive",hub_cluster_id="mycluster_id",vendor="OpenShift",version="4.3.1"} 1
				`,
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

	mc := &mciv1beta1.ManagedClusterInfo{
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

	client := fake.NewSimpleDynamicClient(s, mcU)
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
