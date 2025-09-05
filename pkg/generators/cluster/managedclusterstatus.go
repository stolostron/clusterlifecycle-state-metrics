// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descClusterStatusName           = "acm_managed_cluster_status_condition"
	descClusterStatusHelp           = "Managed cluster status condition"
	requiredClusterStatusConditions = []string{
		mcv1.ManagedClusterConditionAvailable,
	}
)

func GetManagedClusterStatusMetricFamilies() metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descClusterStatusName,
		Type: metric.Gauge,
		Help: descClusterStatusHelp,
		GenerateFunc: wrapManagedClusterInfoFunc(func(mc *mcv1.ManagedCluster) metric.Family {
			klog.Infof("Wrap %s", mc.GetName())
			keys := []string{}
			values := []string{}
			if clusterId := getClusterID(mc); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}
			keys = append(keys, "managed_cluster_name")
			values = append(values, mc.GetName())

			f := generators.BuildStatusConditionMetricFamily(
				mc.Status.Conditions,
				keys,
				values,
				requiredClusterStatusConditions,
				getAllowedClusterConditionStatuses,
			)
			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
			return f
		}),
	}
}

func getAllowedClusterConditionStatuses(conditionType string) []metav1.ConditionStatus {
	switch conditionType {
	case mcv1.ManagedClusterConditionHubAccepted, mcv1.ManagedClusterConditionJoined, mcv1.ManagedClusterConditionHubDenied,
		"ManagedClusterImportSucceeded", "ExternalManagedKubeconfigCreatedSucceeded":
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
