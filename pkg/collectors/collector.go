// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"io"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	ScrapeErrorTotalMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ksm_scrape_error_total",
			Help: "Total scrape errors encountered when scraping a resource",
		},
		[]string{"resource"},
	)

	ResourcesPerScrapeMetric = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "ksm_resources_per_scrape",
			Help: "Number of resources returned per scrape",
		},
		[]string{"resource"},
	)
)

type MetricsCollector interface {
	WriteAll(w io.Writer)
}

// composedMetricsCollector is a collector that composes multiple
// MetricsCollectors into a single one.
type composedMetricsCollector struct {
	collectors []MetricsCollector
}

func newComposedMetricsCollector(collectors ...MetricsCollector) *composedMetricsCollector {
	return &composedMetricsCollector{
		collectors: collectors,
	}
}

func (c *composedMetricsCollector) WriteAll(w io.Writer) {
	for _, collector := range c.collectors {
		collector.WriteAll(w)
	}
}
