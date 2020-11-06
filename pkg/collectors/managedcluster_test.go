package collectors

import (
	"testing"
	"time"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	hivev1 "github.com/openshift/hive/pkg/apis/hive/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getManagedClusterrMetricFamilies(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(managedclusterv1.SchemeGroupVersion, &managedclusterv1.ManagedCluster{})
	s.AddKnownTypes(hivev1.SchemeGroupVersion, &hivev1.ClusterDeployment{})

	mcImported := &managedclusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "imported-cluster",
			CreationTimestamp: metav1.Time{Time: time.Unix(1500000000, 0)},
			Labels: map[string]string{
				"cloud":  "aws",
				"vendor": "OpneShift",
			},
		},
		Status: managedclusterv1.ManagedClusterStatus{
			Version: managedclusterv1.ManagedClusterVersion{
				Kubernetes: "v1.16.2",
			},
		},
	}
	mcHive := &managedclusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "hive-cluster",
			CreationTimestamp: metav1.Time{Time: time.Unix(1500000000, 0)},
			Labels: map[string]string{
				"cloud":  "aws",
				"vendor": "OpneShift",
			},
		},
		Status: managedclusterv1.ManagedClusterStatus{
			Version: managedclusterv1.ManagedClusterVersion{
				Kubernetes: "v1.16.2",
			},
		},
	}
	cd := &hivev1.ClusterDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hive-cluster",
			Namespace: "hive-cluster",
		},
	}

	client := fake.NewSimpleDynamicClient(s, mcImported, mcHive, cd)
	tests := []generateMetricsTestCase{
		{
			Obj:         mcImported,
			MetricNames: []string{"ocm_managedcluster_labels"},
			Want: `
			ocm_managedcluster_labels{cloud="aws",created_via="imported",managedcluster="imported-cluster",vendor="OpneShift",version="v1.16.2"} 1
				`,
		},
		{
			Obj:         mcHive,
			MetricNames: []string{"ocm_managedcluster_labels"},
			Want: `
			ocm_managedcluster_labels{cloud="aws",created_via="hive",managedcluster="hive-cluster",vendor="OpneShift",version="v1.16.2"} 1
				`,
		},
	}
	for i, c := range tests {
		c.Func = metric.ComposeMetricGenFuncs(getManagedClusterMetricFamilies(client))
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
