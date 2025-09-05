// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

var (
	descTimestampName                   = "acm_managed_cluster_import_timestamp"
	descTimestampHelp                   = "The timestamp of different status when importing an ACM managed clusters"
	hostingClusterNameAnnotation string = "import.open-cluster-management.io/hosting-cluster-name"
)

func GetManagedClusterTimestampMetricFamilies(hubClusterID string,
	getClusterTimestamps func(clusterName string) map[string]float64) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descTimestampName,
		Type: metric.Gauge,
		Help: descTimestampHelp,
		GenerateFunc: wrapManagedClusterTimestampFunc(func(mc *mcv1.ManagedCluster) metric.Family {
			klog.Infof("Wrap %s", mc.GetName())
			keys := []string{"managed_cluster_name"}
			values := []string{mc.GetName()}
			if clusterId := getClusterID(mc); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}
			if hostingcluster, ok := mc.GetAnnotations()[hostingClusterNameAnnotation]; ok {
				keys = append(keys, "hosting_cluster_name")
				values = append(values, hostingcluster)
			}

			f := buildManagedClusterTimestampMetricFamily(
				mc,
				keys,
				values,
				getClusterTimestamps,
			)
			klog.V(4).Infof("Returning %v", string(f.ByteSlice()))
			return f
		}),
	}
}

func wrapManagedClusterTimestampFunc(f func(obj *mcv1.ManagedCluster) metric.Family) func(interface{}) *metric.Family {
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

func buildManagedClusterTimestampMetricFamily(mc *mcv1.ManagedCluster, labelKeys, labelValues []string,
	getClusterTimestamps func(clusterName string) map[string]float64) metric.Family {
	family := metric.Family{}

	family.Metrics = append(family.Metrics,
		generators.BuildTimestampMetric(mc.CreationTimestamp, labelKeys, labelValues, generators.CreatedTimestamp))

	// handle existing conditions
	joinedCond := meta.FindStatusCondition(mc.Status.Conditions, mcv1.ManagedClusterConditionJoined)
	if joinedCond != nil && joinedCond.Status == metav1.ConditionTrue {
		family.Metrics = append(family.Metrics, generators.BuildTimestampMetric(
			joinedCond.LastTransitionTime, labelKeys, labelValues, generators.JoinedTimestamp))
	}

	timestamps := getClusterTimestamps(mc.GetName())
	if len(timestamps) == 0 {
		return family
	}

	for status, timestamp := range timestamps {
		family.Metrics = append(family.Metrics,
			&metric.Metric{
				// do not use 'LabelValues: append(labelValues, status),',
				// prevent from using the shared backing array
				LabelKeys:   append([]string{"status"}, labelKeys...),
				LabelValues: append([]string{status}, labelValues...),
				Value:       timestamp,
			})
	}
	return family
}
