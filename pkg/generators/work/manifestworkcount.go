// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/klog/v2"
)

var (
	descWorkCountName = "acm_manifestwork_count"
	descWorkCountHelp = "ManifestWork count"
)

func GetManifestWorkCountMetricFamilies() metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descWorkCountName,
		Type: metric.Gauge,
		Help: descWorkCountHelp,
		GenerateFunc: func(obj interface{}) *metric.Family {
			count, ok := obj.(int)
			if !ok {
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
