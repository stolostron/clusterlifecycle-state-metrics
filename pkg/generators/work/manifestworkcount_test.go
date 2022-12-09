// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package work

import (
	"testing"

	testcommon "github.com/stolostron/clusterlifecycle-state-metrics/test/unit/common"
	"k8s.io/kube-state-metrics/pkg/metric"
)

func Test_getManifestWorkCountMetricFamilies(t *testing.T) {
	tests := []testcommon.GenerateMetricsTestCase{
		{
			Name:        "test manifestwork count",
			Obj:         21,
			MetricNames: []string{"acm_manifestwork_count"},
			Want:        `acm_manifestwork_count 21`,
		},
		{
			Name:        "test manifestwork count with invalid input",
			Obj:         "abc",
			MetricNames: []string{"acm_manifestwork_count"},
			Want:        ``,
		},
	}

	for i, c := range tests {
		t.Run(c.Name, func(t *testing.T) {
			c.Func = metric.ComposeMetricGenFuncs(
				[]metric.FamilyGenerator{GetManifestWorkCountMetricFamilies()},
			)
			if err := c.Run(); err != nil {
				t.Errorf("unexpected collecting result in %v run:\n%s", i, err)
			}
		})
	}
}
