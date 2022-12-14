// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	"k8s.io/klog/v2"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descClusterStatusName           = "acm_managed_cluster_status_condition"
	descClusterStatusHelp           = "Managed cluster status condition"
	requiredClusterStatusConditions = []string{
		"ManagedClusterConditionAvailable",
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
			)
			klog.Infof("Returning %v", string(f.ByteSlice()))
			return f
		}),
	}
}
