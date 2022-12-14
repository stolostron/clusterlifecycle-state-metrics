// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"k8s.io/client-go/tools/cache"
)

// composedStore implements the k8s.io/client-go/tools/cache.Store
// interface. It composes multiple Store into a single one.
type composedStore struct {
	stores []cache.Store
}

// newComposedStore returns a new composedStore
func newComposedStore(stores ...cache.Store) *composedStore {
	return &composedStore{
		stores: stores,
	}
}

func (s *composedStore) AddStore(store cache.Store) {
	s.stores = append(s.stores, store)
}

func (s *composedStore) Size() int {
	return len(s.stores)
}

// Add implements the Add method of the store interface.
func (s *composedStore) Add(obj interface{}) error {
	for _, store := range s.stores {
		if err := store.Add(obj); err != nil {
			return err
		}
	}

	return nil
}

// Update implements the Update method of the store interface.
func (s *composedStore) Update(obj interface{}) error {
	for _, store := range s.stores {
		if err := store.Update(obj); err != nil {
			return err
		}
	}

	return nil
}

// Delete implements the Delete method of the store interface.
func (s *composedStore) Delete(obj interface{}) error {
	for _, store := range s.stores {
		if err := store.Delete(obj); err != nil {
			return err
		}
	}

	return nil
}

// List implements the List method of the store interface.
func (s *composedStore) List() []interface{} {
	return nil
}

// ListKeys implements the ListKeys method of the store interface.
func (s *composedStore) ListKeys() []string {
	return nil
}

// Get implements the Get method of the store interface.
func (s *composedStore) Get(obj interface{}) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// GetByKey implements the GetByKey method of the store interface.
func (s *composedStore) GetByKey(key string) (item interface{}, exists bool, err error) {
	return nil, false, nil
}

// Replace implements the Replace method of the store interface.
func (s *composedStore) Replace(list []interface{}, str string) error {
	for _, store := range s.stores {
		if err := store.Replace(list, str); err != nil {
			return err
		}
	}

	return nil
}

// Resync implements the Resync method of the store interface.
func (s *composedStore) Resync() error {
	for _, store := range s.stores {
		if err := store.Resync(); err != nil {
			return err
		}
	}

	return nil
}
