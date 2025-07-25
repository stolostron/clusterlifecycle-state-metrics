// Copyright (c) 2025 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func newTestClusterDeployment(namespace, name string, hibernating bool) *unstructured.Unstructured {
	cd := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "hive.openshift.io/v1",
			"kind":       "ClusterDeployment",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{},
		},
	}

	if hibernating {
		unstructured.SetNestedField(cd.Object, []interface{}{
			map[string]interface{}{
				"type":   "Hibernating",
				"status": "True",
			},
		}, "status", "conditions")
	}

	return cd
}

func TestClusterHibernatingStateCache(t *testing.T) {
	cases := []struct {
		name                   string
		existing               []interface{}
		add                    []interface{}
		update                 []interface{}
		delete                 []interface{}
		validate               func(t *testing.T, c *clusterHibernatingStateCache)
		expectedErrorOnAdd     bool
		expectedErrorOnDelete  bool
		expectedErrorOnReplace bool
	}{
		{
			name: "add cluster",
			add:  []interface{}{newTestClusterDeployment("c1", "c1", true)},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				state, exists, _ := c.GetByKey("c1")
				if !exists || !state.(bool) {
					t.Errorf("expected cluster c1 to be hibernating")
				}
			},
		},
		{
			name: "add cluster with different name and namespace",
			add:  []interface{}{newTestClusterDeployment("ns1", "c1", true)},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				_, exists, _ := c.GetByKey("c1")
				if exists {
					t.Errorf("expected cluster c1 to be ignored")
				}
			},
		},
		{
			name: "add two clusters",
			existing: []interface{}{
				newTestClusterDeployment("c1", "c1", false),
			},
			add: []interface{}{newTestClusterDeployment("c2", "c2", true)},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				state, exists, _ := c.GetByKey("c1")
				if !exists || state.(bool) {
					t.Errorf("expected cluster c1 to not be hibernating")
				}
				state, exists, _ = c.GetByKey("c2")
				if !exists || !state.(bool) {
					t.Errorf("expected cluster c2 to be hibernating")
				}
			},
		},
		{
			name:     "update cluster",
			existing: []interface{}{newTestClusterDeployment("c1", "c1", false)},
			update:   []interface{}{newTestClusterDeployment("c1", "c1", true)},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				state, exists, _ := c.GetByKey("c1")
				if !exists || !state.(bool) {
					t.Errorf("expected cluster c1 to be hibernating after update")
				}
			},
		},
		{
			name:     "delete cluster",
			existing: []interface{}{newTestClusterDeployment("c1", "c1", true)},
			delete:   []interface{}{newTestClusterDeployment("c1", "c1", true)},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				_, exists, _ := c.GetByKey("c1")
				if exists {
					t.Errorf("expected cluster c1 to not be in cache after deletion")
				}
			},
		},
		{
			name: "replace clusters",
			existing: []interface{}{
				newTestClusterDeployment("c1", "c1", false),
				newTestClusterDeployment("c2", "c2", true),
			},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				state, exists, _ := c.GetByKey("c1")
				if !exists || state.(bool) {
					t.Errorf("expected c1 to not be hibernating")
				}
				state, exists, _ = c.GetByKey("c2")
				if !exists || !state.(bool) {
					t.Errorf("expected c2 to be hibernating")
				}
			},
		},
		{
			name: "list keys",
			existing: []interface{}{
				newTestClusterDeployment("c1", "c1", false),
				newTestClusterDeployment("c2", "c2", true),
			},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				keys := c.ListKeys()
				if len(keys) != 2 {
					t.Errorf("expected 2 keys, got %d", len(keys))
				}
			},
		},
		{
			name: "list",
			existing: []interface{}{
				newTestClusterDeployment("c1", "c1", false),
				newTestClusterDeployment("c2", "c2", true),
			},
			validate: func(t *testing.T, c *clusterHibernatingStateCache) {
				items := c.List()
				if len(items) != 2 {
					t.Errorf("expected 2 items, got %d", len(items))
				}
			},
		},
		{
			name:               "add invalid type",
			add:                []interface{}{"invalid"},
			expectedErrorOnAdd: true,
			validate:           func(t *testing.T, c *clusterHibernatingStateCache) {},
		},
		{
			name:                  "delete invalid type",
			delete:                []interface{}{"invalid"},
			expectedErrorOnDelete: true,
			validate:              func(t *testing.T, c *clusterHibernatingStateCache) {},
		},
		{
			name:                   "replace with invalid type",
			existing:               []interface{}{"invalid"},
			expectedErrorOnReplace: true,
			validate:               func(t *testing.T, c *clusterHibernatingStateCache) {},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cache := newClusterHibernatingStateCache()

			err := cache.Replace(tc.existing, "")
			if (err != nil) != tc.expectedErrorOnReplace {
				t.Errorf("unexpected error on Replace: %v", err)
			}

			for _, item := range tc.add {
				err := cache.Add(item)
				if (err != nil) != tc.expectedErrorOnAdd {
					t.Errorf("unexpected error on Add: %v", err)
				}
			}

			for _, item := range tc.update {
				err := cache.Update(item)
				if err != nil {
					t.Errorf("unexpected error on Update: %v", err)
				}
			}

			for _, item := range tc.delete {
				err := cache.Delete(item)
				if (err != nil) != tc.expectedErrorOnDelete {
					t.Errorf("unexpected error on Delete: %v", err)
				}
			}

			tc.validate(t, cache)
		})
	}
}

func TestIsHibernating(t *testing.T) {
	cases := []struct {
		name     string
		cd       *unstructured.Unstructured
		expected bool
	}{
		{
			name:     "is hibernating",
			cd:       newTestClusterDeployment("c1", "c1", true),
			expected: true,
		},
		{
			name:     "is not hibernating",
			cd:       newTestClusterDeployment("c1", "c1", false),
			expected: false,
		},
		{
			name: "hibernating status is false",
			cd: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "hive.openshift.io/v1",
					"kind":       "ClusterDeployment",
					"metadata": map[string]interface{}{
						"name":      "c1",
						"namespace": "c1",
					},
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Hibernating",
								"status": "False",
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "no status",
			cd: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "hive.openshift.io/v1",
					"kind":       "ClusterDeployment",
					"metadata": map[string]interface{}{
						"name":      "c1",
						"namespace": "c1",
					},
				},
			},
			expected: false,
		},
		{
			name: "no conditions",
			cd: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "hive.openshift.io/v1",
					"kind":       "ClusterDeployment",
					"metadata": map[string]interface{}{
						"name":      "c1",
						"namespace": "c1",
					},
					"status": map[string]interface{}{},
				},
			},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if isHibernating(tc.cd) != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, !tc.expected)
			}
		})
	}
}

func TestClusterHibernatingStateCache_IsHibernating(t *testing.T) {
	cache := newClusterHibernatingStateCache()
	cache.Replace([]interface{}{
		newTestClusterDeployment("c1", "c1", true),
		newTestClusterDeployment("c2", "c2", false),
	}, "")

	if !cache.IsHibernating("c1") {
		t.Errorf("expected c1 to be hibernating")
	}

	if cache.IsHibernating("c2") {
		t.Errorf("expected c2 to not be hibernating")
	}

	if cache.IsHibernating("c3") {
		t.Errorf("expected non-existent cluster to not be hibernating")
	}
}

func TestClusterHibernatingStateCache_Callbacks(t *testing.T) {
	cache := newClusterHibernatingStateCache()
	callbackCounter := 0
	cache.AddOnHibernatingStateChangeFunc(func(clusterName string) error {
		callbackCounter++
		return nil
	})

	// Add a new hibernating cluster, should trigger callback
	cache.Add(newTestClusterDeployment("c1", "c1", true))
	if callbackCounter != 1 {
		t.Errorf("expected callback to be called on add, counter is %d", callbackCounter)
	}

	// Add it again with same state, should not trigger callback
	cache.Add(newTestClusterDeployment("c1", "c1", true))
	if callbackCounter != 1 {
		t.Errorf("expected callback not to be called on same state update, counter is %d", callbackCounter)
	}

	// Update to not hibernating, should trigger callback
	cache.Add(newTestClusterDeployment("c1", "c1", false))
	if callbackCounter != 2 {
		t.Errorf("expected callback to be called on state change, counter is %d", callbackCounter)
	}

	// Add back as hibernating
	cache.Add(newTestClusterDeployment("c1", "c1", true))
	if callbackCounter != 3 {
		t.Errorf("expected callback to be called on add, counter is %d", callbackCounter)
	}
}
