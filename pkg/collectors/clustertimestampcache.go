// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/common"
	"k8s.io/apimachinery/pkg/api/meta"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	operatorv1 "open-cluster-management.io/api/operator/v1"
	workv1 "open-cluster-management.io/api/work/v1"
)

const (
	StatusManagedClusterKubeconfigProvided = "ManagedClusterKubeconfigProvided"
	StatusStartToApplyKlusterletResources  = "StartToApplyKlusterletResources"
)

type onTimestampChangeFunc func(clusterName string) error

// clusterTimestampCache implements the k8s.io/client-go/tools/cache.Store interface.
// It stores timestamps for the cluster importing phases of ManagedCluster objects.
// Note the cached value is for ManagedCluster, but the input obj should be a ManifestWork.
type clusterTimestampCache struct {
	// Protects metrics
	mutex sync.RWMutex

	// data is a map indexed by cluster name with timestamps of different status
	data map[string]map[string]float64

	onTimestampChangeFuncs []onTimestampChangeFunc
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
	mw, ok := obj.(*workv1.ManifestWork)
	if !ok {
		return fmt.Errorf("invalid ManifestWork: %v", obj)
	}

	clusterName, ok := mw.GetLabels()[common.LabelImportHostedCluster]
	if !ok {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	newTimestamps, err := getClusterTimestamps(clusterName, mw.Status)
	if err != nil {
		return err
	}

	timestamps := s.data[clusterName]
	if reflect.DeepEqual(newTimestamps, timestamps) {
		return nil
	}

	s.data[clusterName] = newTimestamps
	klog.InfoS("Timestamps of cluster is changed", "clusterName", clusterName,
		"oldTimestamps", timestamps, "newTimestamps", newTimestamps)

	// run callback funcs once cluster ID is changed
	errs := []error{}
	for _, callback := range s.onTimestampChangeFuncs {
		if err := callback(clusterName); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
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
	return s.Add(obj)
}

// Delete implements the Delete method of the store interface.
func (s *clusterTimestampCache) Delete(obj interface{}) error {
	mw, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	clusterName, ok := mw.GetLabels()[common.LabelImportHostedCluster]
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
	errs := []error{}

	for _, obj := range list {
		mw, ok := obj.(*workv1.ManifestWork)
		if !ok {
			errs = append(errs, fmt.Errorf("invalid ManifestWork: %v", obj))
			continue
		}
		clusterName, ok := mw.GetLabels()[common.LabelImportHostedCluster]
		if !ok {
			continue
		}
		timestamps, err := getClusterTimestamps(clusterName, mw.Status)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.data[clusterName] = timestamps

		// run callback funcs once cluster ID is changed
		for _, callback := range s.onTimestampChangeFuncs {
			if err := callback(clusterName); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return utilerrors.NewAggregate(errs)
}

// Resync implements the Resync method of the store interface.
func (s *clusterTimestampCache) Resync() error {
	return nil
}

func (s *clusterTimestampCache) AddOnTimestampChangeFunc(callback onTimestampChangeFunc) {
	s.onTimestampChangeFuncs = append(s.onTimestampChangeFuncs, callback)
}
