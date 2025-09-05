// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descWorkerCoresName          = "acm_managed_cluster_worker_cores"
	descWorkerCoresHelp          = "The number of worker CPU cores of ACM managed clusters"
	descWorkerCoresDefaultLabels = []string{"hub_cluster_id",
		"managed_cluster_id",
	}
)

func GetManagedClusterWorkerCoresMetricFamilies(hubClusterID string, isHibernatingFn func(string) bool) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descWorkerCoresName,
		Type: metric.Gauge,
		Help: descWorkerCoresHelp,
		GenerateFunc: func(obj interface{}) *metric.Family {
			mc := obj.(*mcv1.ManagedCluster)
			clusterID := getClusterID(mc)

			core_worker, _ := getCapacity(mc)

			// set worker cores to 0 if the ClusterDeployment is in hibernating state.
			if isHibernatingFn(mc.GetName()) {
				core_worker = 0
			}

			if clusterID == "" {
				klog.Infof("Not enough information available for %s", mc.GetName())
				klog.Infof(`\tClusterID=%s,

core_worker=%d`,
					clusterID,
					core_worker,
				)
				return &metric.Family{Metrics: []*metric.Metric{}}
			}
			labelsValues := []string{hubClusterID,
				clusterID,
			}

			f := &metric.Family{Metrics: []*metric.Metric{
				{
					LabelKeys:   descWorkerCoresDefaultLabels,
					LabelValues: labelsValues,
					Value:       float64(core_worker),
				},
			}}
			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
			return f
		},
	}
}
