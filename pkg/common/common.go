package common

import (
	"time"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	workv1 "open-cluster-management.io/api/work/v1"
)

// Manifestworks with any of these 3 labels will be senet timestamp metrics
const (
	// LabelClusterServiceManagementCluster is the label key used by SD
	// to indicate the management cluster of a manifestwork
	LabelClusterServiceManagementCluster = "api.openshift.com/management-cluster"
	// LabelImportHostedCluster is the label key used by import controller
	// to indicate the manifestwork is to deploy a hosted mode klusterlet
	LabelImportHostedCluster = "import.open-cluster-management.io/hosted-cluster"
	// LabelHostedCluster is a reserved label key for manifestwork to record
	// the hosted cluster
	LabelHostedCluster = "api.open-cluster-management.io/hosted-cluster"
)

const (
	// AnnotationAppliedTime is the annotation key used to recored the applied time for resources
	AnnotationAppliedTime = "metrics.open-cluster-management.io/applied-time"
)

type AppliedTimestamp struct {
	AppliedTime time.Time `json:"appliedTime"`
}

// FilterTimestampManifestwork check if the manifestwork should be sent timestamp metrics,
// if should, it also returns the hosted cluster name.
func FilterTimestampManifestwork(mw *workv1.ManifestWork) (bool, string) {
	if len(mw.GetLabels()) == 0 {
		return false, ""
	}
	if hostedcluster, ok := mw.GetLabels()[LabelImportHostedCluster]; ok {
		return true, hostedcluster
	}
	// currently, the service delivery team uses the clusterServiceManagementClusterLabel that can not indicate the
	// hosted cluster, here we reserve a label hostedClusterLabel for them to pass to the hosted cluster in the future
	if hostedcluster, ok := mw.GetLabels()[LabelHostedCluster]; ok {
		return true, hostedcluster
	}
	if _, ok := mw.GetLabels()[LabelClusterServiceManagementCluster]; ok {
		return true, ""
	}

	return false, ""
}

func TimestampManifestworkLabelSelector() (labels.Selector, error) {
	r1, err := labels.NewRequirement(LabelClusterServiceManagementCluster, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	r2, err := labels.NewRequirement(LabelClusterServiceManagementCluster, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	r3, err := labels.NewRequirement(LabelClusterServiceManagementCluster, selection.Exists, nil)
	if err != nil {
		return nil, err
	}
	return labels.NewSelector().Add(*r1, *r2, *r3), nil
}
