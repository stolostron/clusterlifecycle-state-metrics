package collectors

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getManagedClusterMetricFamilies(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(mciv1beta1.GroupVersion, &mciv1beta1.ManagedClusterInfo{})

	mcHive := &mciv1beta1.ManagedClusterInfo{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hive-cluster",
		},
		Status: mciv1beta1.ClusterInfoStatus{
			KubeVendor:  mciv1beta1.KubeVendorOpenShift,
			CloudVendor: mciv1beta1.CloudVendorAWS,
			Version:     "v1.16.2",
			ClusterID:   "managed_cluster_id",
		},
	}
	mcHiveU := &unstructured.Unstructured{}
	err := scheme.Scheme.Convert(mcHive, mcHiveU, nil)
	if err != nil {
		t.Error(err)
	}

	client := fake.NewSimpleDynamicClient(s, mcHiveU)
	tests := []generateMetricsTestCase{
		{
			Obj:         mcHiveU,
			MetricNames: []string{"ocm_managedcluster_info"},
			Want: `
			ocm_managedcluster_info{cloud="Amazon",cluster="hive-cluster",cluster_id="managed_cluster_id",hub_cluster_id="mycluster_id",vendor="OpenShift",version="v1.16.2"} 1
				`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterInfoMetricFamilies("mycluster_id", client))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
