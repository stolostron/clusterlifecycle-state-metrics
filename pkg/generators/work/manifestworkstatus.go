// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"k8s.io/kube-state-metrics/pkg/metric"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	"k8s.io/klog/v2"
	workv1 "open-cluster-management.io/api/work/v1"
)

var (
	descWorkStatusName           = "acm_manifestwork_status_condition"
	descWorkStatusHelp           = "ManifestWork status condition"
	requiredWorkStatusConditions = []string{
		"Applied",
		"Available",
	}
)

func GetManifestWorkStatusMetricFamilies(getClusterIdFunc func(string) string) metric.FamilyGenerator {
	return metric.FamilyGenerator{
		Name: descWorkStatusName,
		Type: metric.Gauge,
		Help: descWorkStatusHelp,
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

			f := generators.BuildStatusConditionMetricFamily(
				mw.Status.Conditions,
				keys,
				values,
				requiredWorkStatusConditions,
			)
			klog.Infof("Returning %v", string(f.ByteSlice()))
			return &f
		},
	}
}
