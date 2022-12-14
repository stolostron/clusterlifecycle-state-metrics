// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"io"
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// CounterMetricsStore implements the k8s.io/client-go/tools/cache.Store
// interface. Instead of storing entire Kubernetes objects, it stores UID of
// those objects, which can be used to caculate the number of the stored objects.
type CounterMetricsStore struct {
	// Protects metrics
	mutex sync.RWMutex

	// metricFamilies is a slice of metric families, containing a slice of metrics.
	metricFamilies [][]byte

	// data is a map indexed by Kubernetes object id with empty structs
	data map[types.UID]struct{}

	// headers contains the header (TYPE and HELP) of each metric family.
	headers []string

	// generateMetricsFunc generates metrics based on a given int (the number of the
	// stored objects) and returns them grouped by metric family.
	generateMetricsFunc func(interface{}) []metricsstore.FamilyByteSlicer
}

// newCounterMetricsStore returns a new CounterMetricsStore
func newCounterMetricsStore(headers []string, generateFunc func(interface{}) []metricsstore.FamilyByteSlicer) *CounterMetricsStore {
	return &CounterMetricsStore{
		generateMetricsFunc: generateFunc,
		headers:             headers,
		data:                map[types.UID]struct{}{},
	}
}

// Add implements the Add method of the store interface.
func (s *CounterMetricsStore) Add(obj interface{}) error {
	o, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	uid := o.GetUID()
	if _, ok := s.data[uid]; ok {
		return nil
	}

	s.data[uid] = struct{}{}
	families := s.generateMetricsFunc(len(s.data))
	s.metricFamilies = make([][]byte, len(families))

	for i, f := range families {
		s.metricFamilies[i] = f.ByteSlice()
	}

	return nil
}

// Update implements the Update method of the store interface.
func (s *CounterMetricsStore) Update(obj interface{}) error {
	// TODO: For now, just call Add, in the future one could check if the resource version changed?
	return s.Add(obj)
}

// Delete implements the Delete method of the store interface.
func (s *CounterMetricsStore) Delete(obj interface{}) error {

	o, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	uid := o.GetUID()
	if _, ok := s.data[uid]; !ok {
		return nil
	}

	delete(s.data, uid)
	families := s.generateMetricsFunc(len(s.data))
	s.metricFamilies = make([][]byte, len(families))

	for i, f := range families {
		s.metricFamilies[i] = f.ByteSlice()
	}

	return nil
}

// List implements the List method of the store interface.
func (s *CounterMetricsStore) List() []interface{} {
	return nil
}

// ListKeys implements the ListKeys method of the store interface.
func (s *CounterMetricsStore) ListKeys() []string {
	return nil
}

// Get implements the Get method of the store interface.
func (s *CounterMetricsStore) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// GetByKey implements the GetByKey method of the store interface.
func (s *CounterMetricsStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// Add implements the Add method of the store interface.
func (s *CounterMetricsStore) Replace(list []interface{}, _ string) error {
	s.mutex.Lock()
	s.data = map[types.UID]struct{}{}
	s.mutex.Unlock()

	for _, o := range list {
		err := s.Add(o)
		if err != nil {
			return err
		}
	}

	return nil
}

// Resync implements the Resync method of the store interface.
func (s *CounterMetricsStore) Resync() error {
	return nil
}

// WriteAll writes all metrics of the store into the given writer, zipped with the
// help text of each metric family.
func (s *CounterMetricsStore) WriteAll(w io.Writer) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for i, help := range s.headers {
		write(w, []byte(help))
		write(w, []byte{'\n'})

		// only output the header if the metric families are initialized
		if len(s.metricFamilies) > i {
			write(w, s.metricFamilies[i])
		}
	}
}

func write(w io.Writer, data []byte) {
	if _, err := w.Write(data); err != nil {
		klog.Errorf("cannot write data: %v", string(data))
	}
}
