// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	operatorv1 "open-cluster-management.io/api/operator/v1"
	workv1 "open-cluster-management.io/api/work/v1"
)

const (
	HostedClusterLabel                     = "import.open-cluster-management.io/hosted-cluster"
	StatusManagedClusterKubeconfigProvided = "ManagedClusterKubeconfigProvided"
	StatusStartToApplyKlusterletResources  = "StartToApplyKlusterletResources"
)

// clusterTimestampCache implements the k8s.io/client-go/tools/cache.Store interface.
// It stores timestamps for the cluster importing phases of ManagedCluster objects.
// Note the cached value is for ManagedCluster, but the input obj should be a ManifestWork.
type clusterTimestampCache struct {
	// Protects metrics
	mutex sync.RWMutex

	// data is a map indexed by cluster name with cluster IDs
	data map[string]map[string]float64
}

func newClusterTimestampCache() *clusterTimestampCache {
	return &clusterTimestampCache{
		data: map[string]map[string]float64{},
	}
}

func (s *clusterTimestampCache) GetClusterTimestamps(clusterName string) map[string]float64 {
	return s.data[clusterName]
}

// Add implements the Add method of the store interface.
func (s *clusterTimestampCache) Add(obj interface{}) error {
	return s.replace(obj, false)
}

func (s *clusterTimestampCache) replace(obj interface{}, replace bool) error {
	mw, ok := obj.(*workv1.ManifestWork)
	if !ok {
		return fmt.Errorf("invalid ManifestWork: %v", obj)
	}

	clusterName, ok := mw.GetLabels()[HostedClusterLabel]
	if !ok {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	timestamps, err := getClusterTimestamps(clusterName, mw.Status)
	if err != nil {
		return err
	}
	if len(timestamps) == 0 {
		if replace {
			s.data[clusterName] = timestamps
		}
		return nil
	}

	s.data[clusterName] = timestamps
	return nil
}

func getClusterTimestamps(clusterName string, workStatus workv1.ManifestWorkStatus) (map[string]float64, error) {
	for _, manifest := range workStatus.ResourceStatus.Manifests {
		if manifest.ResourceMeta.Group != operatorv1.GroupName ||
			(manifest.ResourceMeta.Kind != "Klusterlet" && manifest.ResourceMeta.Resource != "klusterlets") ||
			manifest.ResourceMeta.Name != hostedKlusterletCRName(clusterName) {
			continue
		}

		return getClusterTimestampsByFeedbackRules(manifest.StatusFeedbacks.Values)
	}

	return nil, nil
}

func getClusterTimestampsByFeedbackRules(feedbackValues []workv1.FeedbackValue) (map[string]float64, error) {
	var status bool = false
	var message, lastTransitionTime string
	for _, fb := range feedbackValues {
		if fb.Name == "ReadyToApply-status" && fb.Value.String != nil {
			status = strings.EqualFold(*fb.Value.String, "True")
		}
		if fb.Name == "ReadyToApply-message" && fb.Value.String != nil {
			message = *fb.Value.String
		}
		if fb.Name == "ReadyToApply-lastTransitionTime" && fb.Value.String != nil {
			lastTransitionTime = *fb.Value.String
		}
	}

	if !status {
		return nil, nil
	}

	timestamps := make(map[string]float64)
	transition, err := time.Parse(time.RFC3339, lastTransitionTime)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last transition time %q: %v", lastTransitionTime, err)
	}
	timestamps[StatusStartToApplyKlusterletResources] = float64(transition.Unix())

	// parse the message to get the kubeconfig secret creation time, the message format is:
	// "Klusterlet is ready to apply, the external managed kubeconfig secret was created at: 2021-07-01T00:00:00Z"
	parts := strings.SplitN(message, "the external managed kubeconfig secret was created at:", 2)
	if len(parts) == 2 {
		creationTime, err := time.Parse(time.RFC3339, strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubeconfig secret creation time %q: %v", parts[1], err)
		}
		timestamps[StatusManagedClusterKubeconfigProvided] = float64(creationTime.Unix())
	}
	return timestamps, nil
}

func hostedKlusterletCRName(managedClusterName string) string {
	return fmt.Sprintf("klusterlet-%s", managedClusterName)
}

// Update implements the Update method of the store interface.
func (s *clusterTimestampCache) Update(obj interface{}) error {
	return s.replace(obj, true)
}

// Delete implements the Delete method of the store interface.
func (s *clusterTimestampCache) Delete(obj interface{}) error {
	mw, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	clusterName, ok := mw.GetLabels()[HostedClusterLabel]
	if !ok {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, clusterName)
	return nil
}

// List implements the List method of the store interface.
func (s *clusterTimestampCache) List() []interface{} {
	return nil
}

// ListKeys implements the ListKeys method of the store interface.
func (s *clusterTimestampCache) ListKeys() []string {
	return nil
}

// Get implements the Get method of the store interface.
func (s *clusterTimestampCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// GetByKey implements the GetByKey method of the store interface.
func (s *clusterTimestampCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	timestamps, ok := s.data[key]
	return timestamps, ok, nil
}

// Replace implements the Replace method of the store interface.
func (s *clusterTimestampCache) Replace(list []interface{}, _ string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data = map[string]map[string]float64{}

	for _, obj := range list {
		mw, ok := obj.(*workv1.ManifestWork)
		if !ok {
			return fmt.Errorf("invalid ManifestWork: %v", obj)
		}
		clusterName, ok := mw.GetLabels()[HostedClusterLabel]
		if !ok {
			return nil
		}
		timestamps, err := getClusterTimestamps(clusterName, mw.Status)
		if err != nil {
			return err
		}
		s.data[clusterName] = timestamps
	}

	return nil
}

// Resync implements the Resync method of the store interface.
func (s *clusterTimestampCache) Resync() error {
	return nil
}
