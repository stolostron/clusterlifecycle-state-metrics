// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package addon

import (
	"testing"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	addonv1alphal1 "open-cluster-management.io/api/addon/v1alpha1"
)

func Test_getManagedClusterAddOnStatusMetricFamilies(t *testing.T) {
	addOn := &addonv1alphal1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "work-manager",
			Namespace: "local-cluster",
		},
		Status: addonv1alphal1.ManagedClusterAddOnStatus{
			Conditions: []metav1.Condition{
				testcommon.NewCondition("UnsupportedConfiguration", metav1.ConditionTrue),
				testcommon.NewCondition("ManifestApplied", metav1.ConditionTrue),
				testcommon.NewCondition("RegistrationApplied", metav1.ConditionTrue),
				testcommon.NewCondition("ClusterCertificateRotated", metav1.ConditionTrue),
				testcommon.NewCondition("Available", metav1.ConditionTrue),
			},
		},
	}

	addOnWithoutCondition := &addonv1alphal1.ManagedClusterAddOn{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "work-manager",
			Namespace: "cluster1",
		},
	}

	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test addon status",
			Obj:         addOn,
			MetricNames: []string{"acm_managed_cluster_addon_status_condition"},
			Want: `acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="UnsupportedConfiguration",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="UnsupportedConfiguration",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="ManifestApplied",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="ManifestApplied",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="RegistrationApplied",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="RegistrationApplied",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="ClusterCertificateRotated",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="ClusterCertificateRotated",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local-cluster",managed_cluster_name="local-cluster",condition="Available",status="unknown"} 0`,
		},
		{
			Name:        "test addon status without condition",
			Obj:         addOnWithoutCondition,
			MetricNames: []string{"acm_managed_cluster_addon_status_condition"},
			Want: `acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="true"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="false"} 0
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="cluster1",managed_cluster_name="cluster1",condition="Available",status="unknown"} 1`,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManagedClusterAddOnStatusMetricFamilies(func(clusterName string) string {
					return clusterName
				})},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
