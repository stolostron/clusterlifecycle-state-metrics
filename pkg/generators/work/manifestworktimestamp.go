// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"

	"k8s.io/klog/v2"
)

var (
	descWorkTimestampName            = "acm_manifestwork_apply_timestamp"
	descWorkTimestampHelp            = "The timestamp of the manifestwork appled"
	clusterServiceHostedClusterLabel = "api.openshift.com/management-cluster"
	importHostedClusterLabel         = "import.open-cluster-management.io/hosted-cluster"
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

			hostedcluster := getHostedClusterName(mw)
			if len(hostedcluster) == 0 {
				return &metric.Family{Metrics: []*metric.Metric{}}
			}

			klog.Infof("Hanlde ManifestWork %s/%s", mw.Namespace, mw.Name)
			keys := []string{"manifestwork", "managed_cluster_name", "hosted_cluster_name"}
			values := []string{mw.Name, mw.Namespace, hostedcluster}
			if clusterId := getClusterIdFunc(mw.Namespace); len(clusterId) > 0 {
				keys = append(keys, "managed_cluster_id")
				values = append(values, clusterId)
			}

			family := metric.Family{}
			family.Metrics = append(family.Metrics,
				generators.BuildTimestampMetric(mw.CreationTimestamp, keys, values, generators.CreatedTimestamp))

			appliedCond := meta.FindStatusCondition(mw.Status.Conditions, workv1.WorkApplied)
			if appliedCond != nil && appliedCond.Status == metav1.ConditionTrue {
				family.Metrics = append(family.Metrics, generators.BuildTimestampMetric(
					appliedCond.LastTransitionTime, keys, values, generators.AppliedTimestamp))
			}
			klog.Infof("Returning %v", string(family.ByteSlice()))
			return &family
		},
	}
}

func getHostedClusterName(mw *workv1.ManifestWork) string {
	if len(mw.GetLabels()) == 0 {
		return ""
	}
	if hostedcluster, ok := mw.GetLabels()[importHostedClusterLabel]; ok {
		return hostedcluster
	}
	if hostedcluster, ok := mw.GetLabels()[clusterServiceHostedClusterLabel]; ok {
		return hostedcluster
	}

	return ""
}
