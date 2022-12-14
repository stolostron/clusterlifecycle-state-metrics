// Copyright (c) 2022 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"regexp"

	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"

	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descManagedClusterLabelInfoName     = "acm_managed_cluster_labels"
	descManagedClusterLabelInfoHelp     = "Managed cluster labels"
	descManagedClusterLabelDefaultLabel = []string{
		"hub_cluster_id",
		"managed_cluster_id",
	}
)

func GetManagedClusterLabelMetricFamilies(hubClusterID string) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descManagedClusterLabelInfoName,
		Type: metric.Gauge,
		Help: descManagedClusterLabelInfoHelp,
		GenerateFunc: wrapManagedClusterLabelFunc(func(mc *mcv1.ManagedCluster) metric.Family {
			klog.Infof("Wrap %s", mc.GetName())

			mangedClusterID := getClusterID(mc)

			labelsKeys := descManagedClusterLabelDefaultLabel
			labelsValues := []string{
				hubClusterID,
				mangedClusterID,
			}

			regex := regexp.MustCompile(`[^\w]+`)
			for key, value := range mc.Labels {
				// Ignore the clusterID label since it is being set within the hub and managed cluster IDs
				if key != "clusterID" {
					labelsKeys = append(labelsKeys, regex.ReplaceAllString(key, "_"))
					labelsValues = append(labelsValues, value)
				}
			}

			f := metric.Family{Metrics: []*metric.Metric{
				{
					LabelKeys:   labelsKeys,
					LabelValues: labelsValues,
					Value:       1,
				},
			}}

			klog.Infof("Returning %v", string(f.ByteSlice()))
			return f
		}),
	}
}

func wrapManagedClusterLabelFunc(f func(obj *mcv1.ManagedCluster) metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		cluster := obj.(*mcv1.ManagedCluster)

		metricFamily := f(cluster)
		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return &metricFamily
	}
}
