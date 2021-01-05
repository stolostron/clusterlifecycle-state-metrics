// Copyright (c) 2021 Red Hat, Inc.

package collectors

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-state-metrics/pkg/metric"

	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/klog/v2"
)

const (
	createdViaHive  = "Hive"
	createdViaOther = "Other"
)

var (
	descClusterInfoName          = "acm_managedcluster_info"
	descClusterInfoHelp          = "Managed cluster information"
	descClusterInfoDefaultLabels = []string{"hub_cluster_id",
		"cluster_id",
		"vendor",
		"cloud",
		"version",
		"created_via"}

	cdGVR = schema.GroupVersionResource{
		Group:    "hive.openshift.io",
		Version:  "v1",
		Resource: "clusterdeployments",
	}

	cvGVR = schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "clusterversions",
	}

	mciGVR = schema.GroupVersionResource{
		Group:    "internal.open-cluster-management.io",
		Version:  "v1beta1",
		Resource: "managedclusterinfos",
	}
)

func getManagedClusterInfoMetricFamilies(hubClusterID string, client dynamic.Interface) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: descClusterInfoName,
			Type: metric.MetricTypeGauge,
			Help: descClusterInfoHelp,
			GenerateFunc: wrapManagedClusterInfoFunc(func(mciObj *unstructured.Unstructured) metric.Family {
				klog.Infof("Wrap %s", mciObj.GetName())
				mci := &mciv1beta1.ManagedClusterInfo{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(mciObj.UnstructuredContent(), &mci)
				if err != nil {
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				createdVia := createdViaHive
				cd, errCD := client.Resource(cdGVR).Namespace(mci.GetName()).Get(context.TODO(), mci.GetName(), metav1.GetOptions{})
				if errCD != nil {
					createdVia = createdViaOther
					klog.Infof("Cluster Deployment %s not found, err: %s", mci.GetName(), errCD)
				} else {
					klog.Infof("Cluster Deployment: %v,", cd.Object)
				}
				clusterID := mci.Status.ClusterID
				if clusterID == "" && mci.Status.KubeVendor != mciv1beta1.KubeVendorOpenShift {
					clusterID = mci.GetName()
				}
				version := getVersion(mci)
				if clusterID == "" ||
					mci.Status.KubeVendor == "" ||
					mci.Status.CloudVendor == "" ||
					version == "" {
					klog.Infof("Not enough information available for %s", mci.GetName())
					klog.Infof("\tClusterID=%s,KubeVendor=%s,CloudVendor=%s,Version=%s",
						clusterID,
						mci.Status.KubeVendor,
						mci.Status.CloudVendor,
						mci.Status.Version)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				labelsValues := []string{hubClusterID,
					clusterID,
					string(mci.Status.KubeVendor),
					string(mci.Status.CloudVendor),
					version,
					createdVia}

				f := metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   descClusterInfoDefaultLabels,
						LabelValues: labelsValues,
						Value:       1,
					},
				}}
				klog.Infof("Posting %v", f)
				return f
			}),
		},
	}
}

func getVersion(mci *mciv1beta1.ManagedClusterInfo) string {
	if mci.Status.KubeVendor == "" {
		return ""
	}
	switch mci.Status.KubeVendor {
	case mciv1beta1.KubeVendorOpenShift:
		return mci.Status.DistributionInfo.OCP.Version
	default:
		return mci.Status.Version
	}

}

func wrapManagedClusterInfoFunc(f func(*unstructured.Unstructured) metric.Family) func(interface{}) metric.Family {
	return func(obj interface{}) metric.Family {
		Cluster := obj.(*unstructured.Unstructured)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createManagedClusterInfoListWatchWithClient(client dynamic.Interface, ns string) cache.ListWatch {
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return client.Resource(mciGVR).Namespace(ns).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return client.Resource(mciGVR).Namespace(ns).Watch(context.TODO(), opts)
		},
	}
}
