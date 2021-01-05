// Copyright (c) 2020 Red Hat, Inc.

package collectors

import (
	ocinfrav1 "github.com/openshift/api/config/v1"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
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

func getHubClusterID(c dynamic.Interface) string {

	cvObj, errCv := c.Resource(cvGVR).Get(context.TODO(), "version", metav1.GetOptions{})
	if errCv != nil {
		klog.Fatalf("Error getting cluster version %v \n", errCv)
	}
	cv := &ocinfrav1.ClusterVersion{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(cvObj.UnstructuredContent(), &cv)
	if err != nil {
		klog.Fatalf("Error unmarshal cluster version object%v \n", err)
	}
	return string(cv.Spec.ClusterID)
}
