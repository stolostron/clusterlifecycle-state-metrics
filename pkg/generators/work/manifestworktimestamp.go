// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"

	"k8s.io/klog/v2"
)

var (
	descWorkTimestampName           = "acm_manifestwork_apply_timestamp"
	descWorkTimestampHelp           = "The timestamp of the manifestwork appled"
	requiredWorkTimestampConditions = []string{
		workv1.WorkApplied,
		workv1.WorkAvailable,
	}
)

func GetManifestWorkTimestampMetricFamilies(getClusterIdFunc func(string) string) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descWorkTimestampName,
		Type: metric.Gauge,
		Help: descWorkTimestampHelp,
		GenerateFunc: func(obj interface{}) *metric.Family {
			mw, ok := obj.(*workv1.ManifestWork)
			if !ok {
				klog.Infof("Invalid ManifestWork: %v", obj)
				return &metric.Family{Metrics: []*metric.Metric{}}
			}

			klog.Infof("Hanlde ManifestWork %s/%s", mw.Namespace, mw.Name)
			keys := []string{"manifestwork"}
			values := []string{mw.Name}
			if clusterId := getClusterIdFunc(mw.Namespace); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}
			keys = append(keys, "managed_cluster_name")
			values = append(values, mw.Namespace)

			family := metric.Family{}
			appliedCond := meta.FindStatusCondition(mw.Status.Conditions, workv1.WorkApplied)
			if appliedCond != nil && appliedCond.Status == metav1.ConditionTrue {
				family.Metrics = append(family.Metrics,
					buildAppliedTimeMetric(appliedCond.LastTransitionTime, keys, values),
				)
			}
			klog.Infof("Returning %v", string(family.ByteSlice()))
			return &family
		},
	}
}

func buildAppliedTimeMetric(joinedTime metav1.Time, keys, values []string) *metric.Metric {
	labelKeys := append(keys, "status")

	metric := &metric.Metric{
		LabelKeys:   labelKeys,
		LabelValues: append(values, "Joined"),
		Value:       float64(joinedTime.Unix()),
	}

	return metric
}
