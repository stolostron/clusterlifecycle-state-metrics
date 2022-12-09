// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package cluster

import (
	"testing"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getManagedClusterCountMetricFamilies(t *testing.T) {
	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test cluster count",
			Obj:         10,
			MetricNames: []string{"acm_managed_cluster_count"},
			Want:        `acm_managed_cluster_count 10`,
		},
		{
			Name:        "test cluster count with invalid input",
			Obj:         "abc",
			MetricNames: []string{"acm_managed_cluster_count"},
			Want:        ``,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManagedClusterCountMetricFamilies()},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
