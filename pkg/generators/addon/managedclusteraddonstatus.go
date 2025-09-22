// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package addon

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
)

var (
	descAddOnStatusName           = "acm_managed_cluster_addon_status_condition"
	descAddOnStatusHelp           = "Managed cluster add-on status condition"
	requiredAddOnStatusConditions = []string{
		addonv1alpha1.ManagedClusterAddOnConditionAvailable,
	}
)

func GetManagedClusterAddOnStatusMetricFamilies(getClusterIdFunc func(string) string) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descAddOnStatusName,
		Type: metric.Gauge,
		Help: descAddOnStatusHelp,
		GenerateFunc: func(obj interface{}) *metric.Family {
			addon, ok := obj.(*addonv1alpha1.ManagedClusterAddOn)
			if !ok {
				klog.Errorf("Invalid ManagedClusterAddOn: %v", obj)
				return &metric.Family{Metrics: []*metric.Metric{}}
			}

			klog.Infof("Hanlde ManagedClusterAddOn %s/%s", addon.Namespace, addon.Name)
			keys := []string{"addon_name"}
			values := []string{addon.Name}
			if clusterId := getClusterIdFunc(addon.Namespace); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}
			keys = append(keys, "managed_cluster_name")
			values = append(values, addon.Namespace)

			f := generators.BuildStatusConditionMetricFamily(
				addon.Status.Conditions,
				keys,
				values,
				requiredAddOnStatusConditions,
				getAllowedAddOnConditionStatuses,
			)
			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
			return &f
		},
	}
}

func getAllowedAddOnConditionStatuses(conditionType string) []metav1.ConditionStatus {
	switch conditionType {
	case "RegistrationApplied", "ManifestApplied", "ClusterCertificateRotated", "UnsupportedConfiguration":
		return []metav1.ConditionStatus{
			metav1.ConditionTrue,
			metav1.ConditionFalse,
		}
	default:
		return []metav1.ConditionStatus{
			metav1.ConditionTrue,
			metav1.ConditionFalse,
			metav1.ConditionUnknown,
		}
	}
}
