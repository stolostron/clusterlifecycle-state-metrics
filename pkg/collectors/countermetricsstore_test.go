// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"bytes"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kube-state-metrics/pkg/metric"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

func Test_CounterMetricsStore_WriteAll(t *testing.T) {
	headers := []string{`# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge`}

	generateFunc := func(obj interface{}) []metricsstore.FamilyByteSlicer {
		count := obj.(int)
		return []metricsstore.FamilyByteSlicer{
			&metric.Family{
				Name: "acm_manifestwork_count",
				Metrics: []*metric.Metric{
					{
						Value: float64(count),
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		toAdd    []string
		toUpdate []string
		toDelete []string
		want     string
	}{
		{
			name: "empty",
			want: `# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
`,
		},
		{
			name:  "add",
			toAdd: []string{"a", "b"},
			want: `# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
acm_manifestwork_count 2
`,
		},
		{
			name:     "update",
			toAdd:    []string{"a", "b"},
			toUpdate: []string{"b", "c"},
			want: `# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
acm_manifestwork_count 3
`,
		},
		{
			name:     "delete",
			toAdd:    []string{"a", "b"},
			toDelete: []string{"b", "c"},
			want: `# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
acm_manifestwork_count 1
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCounterMetricsStore(headers, generateFunc)
			for _, obj := range tt.toAdd {
				store.Add(toObject(obj))
			}

			for _, obj := range tt.toUpdate {
				store.Update(toObject(obj))
			}

			for _, obj := range tt.toDelete {
				store.Delete(toObject(obj))
			}

			buf := new(bytes.Buffer)
			store.WriteAll(buf)
			if buf.String() != tt.want {
				t.Errorf("want\n%s\nbut got\n%v", tt.want, buf.String())
			}
		})
	}
}

func toObject(uid string) runtime.Object {
	cluster := &mcv1.ManagedCluster{}
	cluster.UID = types.UID(uid)
	return cluster
}
