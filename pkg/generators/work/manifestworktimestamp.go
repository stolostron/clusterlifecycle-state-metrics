// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"encoding/json"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kube-state-metrics/pkg/metric"
	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/generators"
)

var (
	descWorkTimestampName                = "acm_manifestwork_apply_timestamp"
	descWorkTimestampHelp                = "The timestamp of the manifestwork appled"
	clusterServiceManagementClusterLabel = "api.openshift.com/management-cluster"
	importHostedClusterLabel             = "import.open-cluster-management.io/hosted-cluster"
	hostedClusterLabel                   = "api.open-cluster-management.io/hosted-cluster"
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

			report, hostedcluster := filterManifestwork(mw)
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

			generationTime := getGenerationTime(mw)
			if generationTime != nil {
				gkeys := append(keys, "generation")
				gvalues := append(values, fmt.Sprint(generationTime.Generation))
				family.Metrics = append(family.Metrics,
					generators.BuildTimestampMetric(
						metav1.NewTime(generationTime.CreatedTime),
						gkeys, gvalues, generators.CreatedTimestamp),
				)
				family.Metrics = append(family.Metrics,
					generators.BuildTimestampMetric(
						metav1.NewTime(generationTime.AppliedTime),
						gkeys, gvalues, generators.AppliedTimestamp),
				)

				if generationTime.Generation != 1 {
					gkeys := append(keys, "generation")
					gvalues := append(values, "1")
					family.Metrics = append(family.Metrics,
						generators.BuildTimestampMetric(
							mw.CreationTimestamp, gkeys, gvalues, generators.CreatedTimestamp),
					)
					family.Metrics = append(family.Metrics,
						generators.BuildTimestampMetric(
							metav1.NewTime(generationTime.FirstGenerationAppliedTime),
							gkeys, gvalues, generators.AppliedTimestamp),
					)
				}
			} else {
				family.Metrics = append(family.Metrics,
					generators.BuildTimestampMetric(mw.CreationTimestamp, keys, values, generators.CreatedTimestamp))

				appliedCond := meta.FindStatusCondition(mw.Status.Conditions, workv1.WorkApplied)
				if appliedCond != nil && appliedCond.Status == metav1.ConditionTrue {
					family.Metrics = append(family.Metrics, generators.BuildTimestampMetric(
						appliedCond.LastTransitionTime, keys, values, generators.AppliedTimestamp))
				}
			}

			klog.Infof("Returning %v", string(family.ByteSlice()))
			return &family
		},
	}
}

func filterManifestwork(mw *workv1.ManifestWork) (bool, string) {
	if len(mw.GetLabels()) == 0 {
		return false, ""
	}
	if hostedcluster, ok := mw.GetLabels()[importHostedClusterLabel]; ok {
		return true, hostedcluster
	}
	// currently, the service delivery team uses the clusterServiceManagementClusterLabel that can not indicate the
	// hosted cluster, here we reserve a label hostedClusterLabel for them to pass to the hosted cluster in the future
	if hostedcluster, ok := mw.GetLabels()[hostedClusterLabel]; ok {
		return true, hostedcluster
	}
	if _, ok := mw.GetLabels()[clusterServiceManagementClusterLabel]; ok {
		return true, ""
	}

	return false, ""
}

const (
	// TODO: move to the api repo
	annotationKeyGenerationTime = "metrics.open-cluster-management.io/observed-generation-time"
)

type GenerationTimestamp struct {
	Generation                 int64     `json:"generation"`
	CreatedTime                time.Time `json:"createdTime"`
	AppliedTime                time.Time `json:"appliedTime"`
	FirstGenerationAppliedTime time.Time `json:"firstGenerationAppliedTime"`
}

func getGenerationTime(mw *workv1.ManifestWork) *GenerationTimestamp {
	value, ok := mw.Annotations[annotationKeyGenerationTime]
	if !ok || len(value) == 0 {
		return nil
	}

	generationTime := &GenerationTimestamp{}
	err := json.Unmarshal([]byte(value), generationTime)
	if err != nil {
		return nil
	}
	return generationTime
}
