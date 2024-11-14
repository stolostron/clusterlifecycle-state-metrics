// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descTimestampName          = "acm_managed_cluster_import_timestamp"
	descTimestampHelp          = "The timestamp of different status when importing an ACM managed clusters"
	descTimestampDefaultLabels = []string{"hub_cluster_id",
		"managed_cluster_id",
		"managed_cluster_name",
	}
)

func GetManagedClusterTimestampMetricFamilies(hubClusterID string) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descTimestampName,
		Type: metric.Gauge,
		Help: descTimestampHelp,
		GenerateFunc: wrapManagedClusterTimestampFunc(func(mc *mcv1.ManagedCluster) metric.Family {
			klog.Infof("Wrap %s", mc.GetName())
			keys := []string{}
			values := []string{}
			if clusterId := getClusterID(mc); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}
			keys = append(keys, "managed_cluster_name")
			values = append(values, mc.GetName())

			f := buildManagedClusterTimestampMetricFamily(
				mc,
				keys,
				values,
				requiredClusterStatusConditions,
				getAllowedClusterConditionStatuses,
			)
			klog.Infof("Returning %v", string(f.ByteSlice()))
			return f
		}),
	}
}

func wrapManagedClusterTimestampFunc(f func(obj *mcv1.ManagedCluster) metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		Cluster := obj.(*mcv1.ManagedCluster)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return &metricFamily
	}
}

func buildManagedClusterTimestampMetricFamily(mc *mcv1.ManagedCluster, labelKeys, labelValues, requiredConditionTypes []string, getAllowedConditionStatuses func(conditionType string) []metav1.ConditionStatus) metric.Family {
	family := metric.Family{}

	family.Metrics = append(family.Metrics, buildCreationTimeMetric(mc.CreationTimestamp, labelKeys, labelValues))

	// handle existing conditions
	joinedCond := meta.FindStatusCondition(mc.Status.Conditions, mcv1.ManagedClusterConditionJoined)
	if joinedCond != nil && joinedCond.Status == metav1.ConditionTrue {
		family.Metrics = append(family.Metrics, buildJoinedTimeMetric(joinedCond.LastTransitionTime, labelKeys, labelValues))
	}
	return family
}

func buildCreationTimeMetric(creationTime metav1.Time, keys, values []string) *metric.Metric {
	labelKeys := append(keys, "status")

	metric := &metric.Metric{
		LabelKeys:   labelKeys,
		LabelValues: append(values, "Created"),
		Value:       float64(creationTime.Unix()),
	}

	return metric
}

func buildJoinedTimeMetric(joinedTime metav1.Time, keys, values []string) *metric.Metric {
	labelKeys := append(keys, "status")

	metric := &metric.Metric{
		LabelKeys:   labelKeys,
		LabelValues: append(values, "Joined"),
		Value:       float64(joinedTime.Unix()),
	}

	return metric
}
