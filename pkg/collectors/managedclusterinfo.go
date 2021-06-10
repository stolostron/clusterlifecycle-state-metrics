// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"context"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-state-metrics/pkg/metric"

	mcv1 "github.com/open-cluster-management/api/cluster/v1"
	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/klog/v2"
)

const (
	createdViaHive      = "Hive"
	createdViaOther     = "Other"
	createdViaDiscovery = "Discovery"

	workerLabel = "node-role.kubernetes.io/worker"

	resourceCoreWorker   mcv1.ResourceName = "core_worker"
	resourceSocketWorker mcv1.ResourceName = "socket_worker"

	createdViaAnnotation          = "open-cluster-management/created-via"
	createdViaAnnotationDiscovery = "discovery"
)

var (
	descClusterInfoName          = "acm_managed_cluster_info"
	descClusterInfoHelp          = "Managed cluster information"
	descClusterInfoDefaultLabels = []string{"hub_cluster_id",
		"managed_cluster_id",
		"vendor",
		"cloud",
		"version",
		"available",
		"created_via",
		"core_worker",
		"socket_worker"}

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

	mcGVR = schema.GroupVersionResource{
		Group:    "cluster.open-cluster-management.io",
		Version:  "v1",
		Resource: "managedclusters",
	}
)

func getManagedClusterInfoMetricFamilies(hubClusterID string, client dynamic.Interface) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: descClusterInfoName,
			Type: metric.Gauge,
			Help: descClusterInfoHelp,
			GenerateFunc: wrapManagedClusterInfoFunc(func(obj *unstructured.Unstructured) metric.Family {
				klog.Infof("Wrap %s", obj.GetName())
				mciU, errMCI := client.Resource(mciGVR).Namespace(obj.GetName()).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
				if errMCI != nil {
					klog.Errorf("Error: %v", errMCI)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				mci := &mciv1beta1.ManagedClusterInfo{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(mciU.UnstructuredContent(), &mci)
				if err != nil {
					klog.Errorf("Error: %v", err)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				mcU, errMC := client.Resource(mcGVR).Get(context.TODO(), mci.GetName(), metav1.GetOptions{})
				if errMC != nil {
					klog.Errorf("Error: %v", errMC)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				klog.Infof("mcU: %v", mcU)
				mc := &mcv1.ManagedCluster{}
				err = runtime.DefaultUnstructuredConverter.FromUnstructured(mcU.UnstructuredContent(), &mc)
				if err != nil {
					klog.Errorf("Error: %v", err)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				available := getAvailableStatus(mc)
				// klog.Infof("mc: %v", mc)
				createdVia := getCreateVia(client, mci, mc)
				clusterID := mci.Status.ClusterID
				//Cluster ID is not available on non-OCP thus use the name
				if clusterID == "" &&
					mci.Status.KubeVendor != mciv1beta1.KubeVendorOpenShift {
					clusterID = mci.GetName()
				}

				//ClusterID is not available on OCP 3.x thus use the name
				if clusterID == "" &&
					mci.Status.KubeVendor == mciv1beta1.KubeVendorOpenShift && mci.Status.DistributionInfo.OCP.Version == "3" {
					clusterID = mci.GetName()
				}

				version := getVersion(mci)
				core_worker, socket_worker := getCapacity(mc)

				nodeListLength := len(mci.Status.NodeList)

				if clusterID == "" ||
					mci.Status.KubeVendor == "" ||
					mci.Status.CloudVendor == "" ||
					version == "" ||
					nodeListLength == 0 ||
					((core_worker == 0 || socket_worker == 0) && hasWorker(mci)) {
					klog.Infof("Not enough information available for %s", mci.GetName())
					klog.Infof(`\tClusterID=%s,
KubeVendor=%s,
CloudVendor=%s,
Version=%s,
available=%s,
NodeList length=%d,
core_worker=%d,
socket_worker=%d`,
						clusterID,
						mci.Status.KubeVendor,
						mci.Status.CloudVendor,
						version,
						available,
						nodeListLength,
						core_worker,
						socket_worker)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				labelsValues := []string{hubClusterID,
					clusterID,
					string(mci.Status.KubeVendor),
					string(mci.Status.CloudVendor),
					version,
					available,
					createdVia,
					strconv.FormatInt(core_worker, 10),
					strconv.FormatInt(socket_worker, 10),
				}

				f := metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   descClusterInfoDefaultLabels,
						LabelValues: labelsValues,
						Value:       1,
					},
				}}
				klog.Infof("Returning %v", string(f.ByteSlice()))
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

func hasWorker(mci *mciv1beta1.ManagedClusterInfo) bool {
	for _, n := range mci.Status.NodeList {
		if _, ok := n.Labels[workerLabel]; ok {
			return true
		}
	}
	return false
}

func getCapacity(mc *mcv1.ManagedCluster) (core_worker, socket_worker int64) {
	if q, ok := mc.Status.Capacity[resourceCoreWorker]; ok {
		core_worker = q.Value()
	}
	if q, ok := mc.Status.Capacity[resourceSocketWorker]; ok {
		socket_worker = q.Value()
	}
	return
}

func getAvailableStatus(mc *mcv1.ManagedCluster) string {
	status := metav1.ConditionUnknown
	for _, c := range mc.Status.Conditions {
		if c.Type == mcv1.ManagedClusterConditionAvailable {
			status = c.Status
			break
		}
	}
	if status == metav1.ConditionFalse {
		status = metav1.ConditionUnknown
	}
	return string(status)
}

func wrapManagedClusterInfoFunc(f func(*unstructured.Unstructured) metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		Cluster := obj.(*unstructured.Unstructured)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return &metricFamily
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

func createManagedClusterListWatchWithClient(client dynamic.Interface) cache.ListWatch {
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return client.Resource(mcGVR).List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return client.Resource(mcGVR).Watch(context.TODO(), opts)
		},
	}
}

func getCreateVia(client dynamic.Interface, mci *mciv1beta1.ManagedClusterInfo, mc *mcv1.ManagedCluster) string {
	createdVia := createdViaHive
	cd, errCD := client.Resource(cdGVR).Namespace(mci.GetName()).Get(context.TODO(), mci.GetName(), metav1.GetOptions{})
	if errCD != nil {
		createdVia = createdViaOther
		klog.Infof("Cluster Deployment %s not found, err: %s", mci.GetName(), errCD)
	} else {
		if v, ok := mc.Annotations[createdViaAnnotation]; ok && v == createdViaAnnotationDiscovery {
			createdVia = createdViaDiscovery
		}
		klog.Infof("Cluster Deployment: %v,", cd.Object)
	}
	return createdVia
}
