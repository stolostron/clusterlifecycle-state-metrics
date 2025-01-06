// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/common"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
)

var (
	descWorkTimestampName = "acm_manifestwork_apply_timestamp"
	descWorkTimestampHelp = "The timestamp of the manifestwork appled"
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

			report, hostedcluster := common.FilterTimestampManifestwork(mw)
			if !report {
				return &metric.Family{Metrics: []*metric.Metric{}}
			}

			klog.Infof("Hanlde ManifestWork %s/%s", mw.Namespace, mw.Name)
			keys := []string{"manifestwork", "managed_cluster_name"}
			values := []string{mw.Name, mw.Namespace}
			if len(hostedcluster) != 0 {
				keys = append(keys, "hosted_cluster_name")
				values = append(values, hostedcluster)
			}
			if clusterId := getClusterIdFunc(mw.Namespace); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}

			family := metric.Family{}
			observedTimestamp := common.GetObservedTimestamp(mw)
			if observedTimestamp != nil {
				family.Metrics = append(family.Metrics,
					generators.BuildTimestampMetric(mw.CreationTimestamp, keys, values, generators.CreatedTimestamp))
				family.Metrics = append(family.Metrics, generators.BuildTimestampMetric(
					metav1.NewTime(observedTimestamp.AppliedTime),
					keys, values, generators.AppliedTimestamp))
			}

			klog.Infof("Returning %v", string(family.ByteSlice()))
			return &family
		},
	}
}
