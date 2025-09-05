// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"
)

var (
	descClusterCountName = "acm_managed_cluster_count"
	descClusterCountHelp = "Managed cluster count"
)

func GetManagedClusterCountMetricFamilies() metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descClusterCountName,
		Type: metric.Gauge,
		Help: descClusterCountHelp,
		GenerateFunc: func(obj interface{}) *metric.Family {
			count, ok := obj.(int)
			if !ok {
				klog.Infof("Invalid number of managed clusters: %v", obj)
				return &metric.Family{Metrics: []*metric.Metric{}}
			}
			f := &metric.Family{Metrics: []*metric.Metric{
				{
					Value: float64(count),
				},
			}}
			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
			return f
		},
	}
}
