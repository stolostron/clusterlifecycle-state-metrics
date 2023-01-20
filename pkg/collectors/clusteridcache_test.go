// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	mciv1beta1 "github.com/stolostron/cluster-lifecycle-api/clusterinfo/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_ClusterIDCache_GetClusterId(t *testing.T) {

	cluster1 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "cluster1-id",
			},
		},
	}
	cluster1Modified := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "cluster1-id-modified",
			},
		},
	}
	cluster2 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster2",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "cluster2-id",
			},
		},
	}

	tests := []struct {
		name              string
		clusterName       string
		existing          []interface{}
		toAdd             []interface{}
		toUpdate          []interface{}
		toDelete          []interface{}
		want              string
		numberOfIdChanged int
	}{
		{
			name:        "empty",
			clusterName: "cluster1",
			want:        "",
		},
		{
			name:        "existing",
			clusterName: "cluster1",
			existing:    []interface{}{cluster1},
			want:        "cluster1-id",
		},
		{
			name:              "add",
			clusterName:       "cluster1",
			toAdd:             []interface{}{cluster1},
			want:              "cluster1-id",
			numberOfIdChanged: 1,
		},
		{
			name:              "update cluster1",
			clusterName:       "cluster1",
			existing:          []interface{}{cluster1},
			toUpdate:          []interface{}{cluster1Modified},
			want:              "cluster1-id-modified",
			numberOfIdChanged: 1,
		},
		{
			name:              "update cluster2",
			clusterName:       "cluster2",
			toAdd:             []interface{}{cluster1},
			toUpdate:          []interface{}{cluster2},
			want:              "cluster2-id",
			numberOfIdChanged: 2,
		},
		{
			name:        "delete",
			clusterName: "cluster1",
			existing:    []interface{}{cluster1},
			toDelete:    []interface{}{cluster1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numberOfIdChanged := 0
			cache := newClusterIdCache()
			cache.AddOnClusterIdChangeFunc(func(clusterName string) error {
				numberOfIdChanged += 1
				return nil
			})

			err := cache.Replace(tt.existing, "")
			if err != nil {
				t.Errorf("caught unexpected err: %c", err)
			}

			for _, obj := range tt.toAdd {
				cache.Add(obj)
			}

			for _, obj := range tt.toUpdate {
				cache.Update(obj)
			}

			for _, obj := range tt.toDelete {
				cache.Delete(obj)
			}

			if actual := cache.GetClusterId(tt.clusterName); actual != tt.want {
				t.Errorf("want\n%s\nbut got\n%v", tt.want, actual)
			}

			if numberOfIdChanged != tt.numberOfIdChanged {
				t.Errorf("want numberOfIdChanged %d\nbut got%d", tt.numberOfIdChanged, numberOfIdChanged)
			}
		})
	}
}

func Test_getClusterId(t *testing.T) {

	cluster1 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster1",
			Labels: map[string]string{
				mciv1beta1.LabelClusterID: "cluster1-id",
			},
		},
	}
	nonOcpCluster := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster2",
		},
	}

	ocp311Cluster := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster3",
			Labels: map[string]string{
				mciv1beta1.LabelKubeVendor: string(mciv1beta1.KubeVendorOpenShift),
				mciv1beta1.OCPVersion:      "3",
			},
		},
	}

	cluster4 := &mcv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster4",
			Labels: map[string]string{
				mciv1beta1.LabelKubeVendor: string(mciv1beta1.KubeVendorOpenShift),
			},
		},
	}

	tests := []struct {
		name    string
		cluster *mcv1.ManagedCluster
		want    string
	}{
		{
			name:    "cluster with cluster id label",
			cluster: cluster1,
			want:    "cluster1-id",
		},
		{
			name:    "non-ocp cluster",
			cluster: nonOcpCluster,
			want:    "cluster2",
		},
		{
			name:    "ocp 311 cluster",
			cluster: ocp311Cluster,
			want:    "cluster3",
		},
		{
			name:    "ocp cluster without id label",
			cluster: cluster4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if actual := getClusterID(tt.cluster); actual != tt.want {
				t.Errorf("want\n%s\nbut got\n%v", tt.want, actual)
			}
		})
	}
}
