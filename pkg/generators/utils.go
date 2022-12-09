// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package generators

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"

	"k8s.io/apimachinery/pkg/util/sets"
)

func BuildStatusConditionMetricFamily(conditions []metav1.Condition, labelKeys, labelValues, requiredConditionTypes []string) metric.Family {
	family := metric.Family{}
	keys := append(labelKeys, "condition", "status")
	missingConditionTypes := sets.NewString(requiredConditionTypes...)

	// handle existing conditions
	for _, condition := range conditions {
		values := append(labelValues, condition.Type, strings.ToLower(string(condition.Status)))
		family.Metrics = append(family.Metrics, &metric.Metric{
			LabelKeys:   keys,
			LabelValues: values,
			Value:       1,
		})

		missingConditionTypes.Delete(condition.Type)
	}

	// add missing required conditions
	for _, conditionType := range missingConditionTypes.List() {
		values := append(labelValues, conditionType, "unknown")
		family.Metrics = append(family.Metrics, &metric.Metric{
			LabelKeys:   keys,
			LabelValues: values,
			Value:       1,
		})
	}
	return family
}
