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

			// Regex patterns which is following the Prometheus metric label validation rules.
			nonWordRegex := regexp.MustCompile(`[^\w]+`) // Regex to check non-word characters
			firstDigitRegex := regexp.MustCompile(`^\d`) // Regex to check if the first character is a digit

			for key, value := range mc.Labels {
				// Ignore the clusterID label since it is being set within the hub and managed cluster IDs
				if key != "clusterID" {
					// Preserve the original key for logging
					originalKey := key

					// Replace non-word characters with underscores
					// For example, label key velero.io/exclude-from-backup will be replaced with velero_io_exclude_from_backup,
					modifiedKey := nonWordRegex.ReplaceAllString(key, "_")

					// If the first character is a digit, prepend an underscore
					// For example, label key 5g-dev01 will be replaced with _5g_dev01.
					if firstDigitRegex.MatchString(modifiedKey) {
						modifiedKey = "_" + modifiedKey
					}

					// If the key was converted, log a warning
					if originalKey != modifiedKey {
						klog.Infof("Label key '%s' was converted to '%s' since it contains non-word characters or a first digit", originalKey, modifiedKey)
					}

					// Add the modified key and value to the label slices
					labelsKeys = append(labelsKeys, modifiedKey)
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

			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
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
