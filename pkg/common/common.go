package common

import (
	"encoding/json"
	"time"

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
	// AnnotationObservedTimestamp is the annotation key used to recored the timestamp
	// for resources that the controller observed
	AnnotationObservedTimestamp = "metrics.open-cluster-management.io/observed-timestamp"
)

type ObservedTimestamp struct {
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

// GetObservedTimestamp gets the observed timestamp from annotation
func GetObservedTimestamp(mw *workv1.ManifestWork) *ObservedTimestamp {
	value, ok := mw.Annotations[AnnotationObservedTimestamp]
	if !ok || len(value) == 0 {
		return nil
	}

	timestamp := &ObservedTimestamp{}
	err := json.Unmarshal([]byte(value), timestamp)
	if err != nil {
		return nil
	}
	return timestamp
}
