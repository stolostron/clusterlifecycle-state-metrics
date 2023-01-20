// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package generators

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/apimachinery/pkg/util/sets"
)

func BuildStatusConditionMetricFamily(conditions []metav1.Condition, labelKeys, labelValues, requiredConditionTypes []string, getAllowedConditionStatuses func(conditionType string) []metav1.ConditionStatus) metric.Family {
	family := metric.Family{}
	missingConditionTypes := sets.NewString(requiredConditionTypes...)

	// handle existing conditions
	for _, condition := range conditions {
		metrics := buildStatusConditionMetrics(condition.Type, condition.Status, labelKeys, labelValues, getAllowedConditionStatuses(condition.Type))
		family.Metrics = append(family.Metrics, metrics...)
		missingConditionTypes.Delete(condition.Type)
	}

	// add missing required conditions
	for _, conditionType := range missingConditionTypes.List() {
		metrics := buildStatusConditionMetrics(conditionType, metav1.ConditionUnknown, labelKeys, labelValues, getAllowedConditionStatuses(conditionType))
		family.Metrics = append(family.Metrics, metrics...)
	}

	return family
}

func buildStatusConditionMetrics(conditionType string, conditionStatus metav1.ConditionStatus, keys, values []string, allowedConditionStatuses []metav1.ConditionStatus) []*metric.Metric {
	labelKeys := append(keys, "condition", "status")

	metrics := []*metric.Metric{}
	for _, status := range allowedConditionStatuses {
		labelValues := append(values, conditionType, strings.ToLower(string(status)))
		metric := &metric.Metric{
			LabelKeys:   labelKeys,
			LabelValues: labelValues,
			Value:       0,
		}
		if status == conditionStatus {
			metric.Value = 1
		}
		metrics = append(metrics, metric)
	}

	return metrics
}
