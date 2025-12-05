// Copyright (c) 2025 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type clusterHibernatingStateCache struct {
	// A map from cluster name to its hibernating state.
	hibernatingStates             map[string]bool
	mutex                         sync.RWMutex
	onHibernatingStateChangeFuncs []func(clusterName string) error
}

func newClusterHibernatingStateCache() *clusterHibernatingStateCache {
	return &clusterHibernatingStateCache{
		hibernatingStates:             make(map[string]bool),
		onHibernatingStateChangeFuncs: []func(clusterName string) error{},
	}
}

func (c *clusterHibernatingStateCache) AddOnHibernatingStateChangeFunc(f func(clusterName string) error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.onHibernatingStateChangeFuncs = append(c.onHibernatingStateChangeFuncs, f)
}

func (c *clusterHibernatingStateCache) IsHibernating(key string) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hibernating, exists := c.hibernatingStates[key]
	if !exists {
		return false
	}
	return hibernating
}

func (c *clusterHibernatingStateCache) Get(obj interface{}) (interface{}, bool, error) {
	cd, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil, false, fmt.Errorf("unexpected object type: %T", obj)
	}
	if cd.GetNamespace() != cd.GetName() {
		return nil, false, nil
	}
	key := cd.GetName()
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hibernating, exists := c.hibernatingStates[key]
	return hibernating, exists, nil
}

func (c *clusterHibernatingStateCache) GetByKey(key string) (interface{}, bool, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	hibernating, exists := c.hibernatingStates[key]
	return hibernating, exists, nil
}

func (c *clusterHibernatingStateCache) Add(obj interface{}) error {
	cd, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}

	if cd.GetNamespace() != cd.GetName() {
		klog.V(4).Infof("ClusterDeployment %s/%s is ignored as its name and namespace are different", cd.GetNamespace(), cd.GetName())
		return nil
	}

	key := cd.GetName()
	newHibernatingState := isHibernating(cd)

	c.mutex.Lock()

	oldHibernatingState, exists := c.hibernatingStates[key]

	switch {
	case exists && oldHibernatingState == newHibernatingState:
		// No change
		c.mutex.Unlock()
		return nil
	case exists && oldHibernatingState != newHibernatingState:
		// Changed
		klog.Infof("ClusterDeployment %s hibernating state changed to %v", key, newHibernatingState)
	default:
		// Not exists
		klog.Infof("ClusterDeployment %s hibernating state is %v", key, newHibernatingState)
	}

	c.hibernatingStates[key] = newHibernatingState
	c.mutex.Unlock()

	for _, f := range c.onHibernatingStateChangeFuncs {
		if err := f(key); err != nil {
			klog.Errorf("failed to call onHibernatingStateChangeFunc for cluster %s: %v", key, err)
		}
	}

	return nil
}

func (c *clusterHibernatingStateCache) Update(obj interface{}) error {
	return c.Add(obj)
}

func (c *clusterHibernatingStateCache) Delete(obj interface{}) error {
	cd, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unexpected object type: %T", obj)
	}

	if cd.GetNamespace() != cd.GetName() {
		klog.V(4).Infof("ClusterDeployment %s/%s is ignored as its name and namespace are different", cd.GetNamespace(), cd.GetName())
		return nil
	}

	key := cd.GetName()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.hibernatingStates, key)
	klog.V(4).Infof("ClusterDeployment %s deleted from hibernating state cache", key)
	return nil
}

func (c *clusterHibernatingStateCache) List() []interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	list := make([]interface{}, 0, len(c.hibernatingStates))
	for _, state := range c.hibernatingStates {
		list = append(list, state)
	}
	return list
}

func (c *clusterHibernatingStateCache) ListKeys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	keys := make([]string, 0, len(c.hibernatingStates))
	for key := range c.hibernatingStates {
		keys = append(keys, key)
	}
	return keys
}

func (c *clusterHibernatingStateCache) Replace(list []interface{}, _ string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.hibernatingStates = make(map[string]bool)
	for _, obj := range list {
		cd, ok := obj.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("unexpected object type: %T", obj)
		}
		if cd.GetNamespace() != cd.GetName() {
			klog.V(4).Infof("ClusterDeployment %s/%s is ignored as its name and namespace are different", cd.GetNamespace(), cd.GetName())
			continue
		}
		key := cd.GetName()
		c.hibernatingStates[key] = isHibernating(cd)
	}
	return nil
}

func (c *clusterHibernatingStateCache) Resync() error {
	return nil
}

func isHibernating(cd *unstructured.Unstructured) bool {
	conditions, found, err := unstructured.NestedSlice(cd.Object, "status", "conditions")
	if err != nil || !found {
		return false
	}

	for _, condition := range conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}
		conditionType, found, err := unstructured.NestedString(conditionMap, "type")
		if err != nil || !found {
			continue
		}
		if conditionType == "Hibernating" {
			status, _, _ := unstructured.NestedString(conditionMap, "status")
			return status == "True"
		}
	}
	return false
}
