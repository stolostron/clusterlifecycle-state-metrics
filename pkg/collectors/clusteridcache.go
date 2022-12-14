// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mciv1beta1 "github.com/stolostron/cluster-lifecycle-api/clusterinfo/v1beta1"
)

// clusterIdCache implements the k8s.io/client-go/tools/cache.Store
// interface. Instead of storing entire ManagedCluster objects, it
// stores cluster IDs of ManagedCluster objects.
type clusterIdCache struct {
	// Protects metrics
	mutex sync.RWMutex

	// data is a map indexed by cluster name with cluster IDs
	data map[string]string
}

// newCounterMetricsStore returns a new CounterMetricsStore
func newClusterIdCache() *clusterIdCache {
	return &clusterIdCache{
		data: map[string]string{},
	}
}

func (s *clusterIdCache) GetClusterId(clusterName string) string {
	return s.data[clusterName]
}

// Add implements the Add method of the store interface.
func (s *clusterIdCache) Add(obj interface{}) error {
	o, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[o.GetName()] = getClusterID(o)
	return nil
}

// Update implements the Update method of the store interface.
func (s *clusterIdCache) Update(obj interface{}) error {
	return s.Add(obj)
}

// Delete implements the Delete method of the store interface.
func (s *clusterIdCache) Delete(obj interface{}) error {

	o, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.data, o.GetName())
	return nil
}

// List implements the List method of the store interface.
func (s *clusterIdCache) List() []interface{} {
	return nil
}

// ListKeys implements the ListKeys method of the store interface.
func (s *clusterIdCache) ListKeys() []string {
	return nil
}

// Get implements the Get method of the store interface.
func (s *clusterIdCache) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// GetByKey implements the GetByKey method of the store interface.
func (s *clusterIdCache) GetByKey(key string) (item interface{}, exists bool, err error) {
	cluserId, ok := s.data[key]
	return cluserId, ok, nil
}

// Add implements the Add method of the store interface.
func (s *clusterIdCache) Replace(list []interface{}, _ string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data = map[string]string{}

	for _, o := range list {
		obj, err := meta.Accessor(o)
		if err != nil {
			return err
		}
		s.data[obj.GetName()] = getClusterID(obj)
	}

	return nil
}

// Resync implements the Resync method of the store interface.
func (s *clusterIdCache) Resync() error {
	return nil
}

func getClusterID(obj metav1.Object) string {
	labels := obj.GetLabels()
	kubeVendor := labels[mciv1beta1.LabelKubeVendor]
	clusterID := labels[mciv1beta1.LabelClusterID]

	// Cluster ID is not available on non-OCP thus use the name
	if clusterID == "" && (kubeVendor != string(mciv1beta1.KubeVendorOpenShift)) {
		clusterID = obj.GetName()
	}
	// ClusterID is not available on OCP 3.x thus use the name
	if clusterID == "" && (kubeVendor == string(mciv1beta1.KubeVendorOpenShift)) && labels[mciv1beta1.OCPVersion] == "3" {
		clusterID = obj.GetName()
	}

	return clusterID
}
