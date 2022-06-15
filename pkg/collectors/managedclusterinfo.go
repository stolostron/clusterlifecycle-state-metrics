// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"context"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kube-state-metrics/pkg/metric"

	mciv1beta1 "github.com/stolostron/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/klog/v2"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

const (
	resourceCoreWorker   mcv1.ResourceName = "core_worker"
	resourceSocketWorker mcv1.ResourceName = "socket_worker"
)

const (
	createdViaAnnotation       = "open-cluster-management/created-via"
	createdViaAnnotationOther  = "Other"
	serviceNameAnnotation      = "open-cluster-management/service-name"
	serviceNameAnnotationOther = "Other"
)

var serviceNameMapping map[string]string = map[string]string{
	"compute": "Compute",
	"other":   "Other",
}

var createdViaMapping map[string]string = map[string]string{
	"discovery":          "Discovery",
	"assisted-installer": "AssistedInstaller",
	"hive":               "Hive",
	"other":              createdViaAnnotationOther,
}

var (
	descClusterInfoName          = "acm_managed_cluster_info"
	descClusterInfoHelp          = "Managed cluster information"
	descClusterInfoDefaultLabels = []string{"hub_cluster_id",
		"managed_cluster_id",
		"vendor",
		"cloud",
		"service_name",
		"version",
		"available",
		"created_via",
		"core_worker",
		"socket_worker"}
)

func getManagedClusterInfoMetricFamilies(hubClusterID string, clusterclient *clusterclient.Clientset) []metric.FamilyGenerator {
	return []metric.FamilyGenerator{
		{
			Name: descClusterInfoName,
			Type: metric.Gauge,
			Help: descClusterInfoHelp,
			GenerateFunc: wrapManagedClusterInfoFunc(func(obj *mcv1.ManagedCluster) metric.Family {
				klog.Infof("Wrap %s", obj.GetName())
				mc, err := clusterclient.ClusterV1().ManagedClusters().Get(context.Background(), obj.GetName(), metav1.GetOptions{})
				if err != nil {
					klog.Errorf("Error: %v", err)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				// klog.Infof("mc: %v", mc)
				kubeVendor := mc.ObjectMeta.Labels[mciv1beta1.LabelKubeVendor]
				cloudVendor := mc.ObjectMeta.Labels[mciv1beta1.LabelCloudVendor]

				clusterID := getClusterID(mc)
				version := getVersion(mc)
				createdVia := getCreatedVia(mc)
				serviceName := getServiceName(mc)
				available := getAvailableStatus(mc)
				core_worker, socket_worker := getCapacity(mc)

				if clusterID == "" ||
					kubeVendor == "" ||
					cloudVendor == "" ||
					version == "" {
					klog.Infof("Not enough information available for %s", obj.GetName())
					klog.Infof(`\tClusterID=%s,
KubeVendor=%s,
CloudVendor=%s,
ServiceName=%s,
Version=%s,
available=%s,
core_worker=%d,
socket_worker=%d`,
						clusterID,
						kubeVendor,
						cloudVendor,
						serviceName,
						version,
						available,
						core_worker,
						socket_worker)
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				labelsValues := []string{hubClusterID,
					clusterID,
					kubeVendor,
					cloudVendor,
					serviceName,
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

func getClusterID(mc *mcv1.ManagedCluster) string {
	kubeVendor := mc.ObjectMeta.Labels[mciv1beta1.LabelKubeVendor]
	clusterID := mc.ObjectMeta.Labels[mciv1beta1.LabelClusterID]

	//Cluster ID is not available on non-OCP thus use the name
	if clusterID == "" && (kubeVendor != string(mciv1beta1.KubeVendorOpenShift)) {
		clusterID = mc.GetName()
	}
	//ClusterID is not available on OCP 3.x thus use the name
	if clusterID == "" && (kubeVendor == string(mciv1beta1.KubeVendorOpenShift)) && mc.ObjectMeta.Labels[mciv1beta1.OCPVersion] == "3" {
		clusterID = mc.GetName()
	}

	return clusterID
}

func getVersion(mc *mcv1.ManagedCluster) string {
	kubeVendor := mc.ObjectMeta.Labels[mciv1beta1.LabelKubeVendor]
	version := ""

	if kubeVendor == "" {
		return version
	}

	switch kubeVendor {
	case string(mciv1beta1.KubeVendorOpenShift):
		version = mc.ObjectMeta.Labels[mciv1beta1.OCPVersion]
	default:
		for _, c := range mc.Status.ClusterClaims {
			if c.Name == "kubeversion.open-cluster-management.io" {
				version = c.Value
			}
		}
	}

	return version
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

func wrapManagedClusterInfoFunc(f func(obj *mcv1.ManagedCluster) metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		Cluster := obj.(*mcv1.ManagedCluster)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return &metricFamily
	}
}

func getCreatedVia(mc *mcv1.ManagedCluster) string {
	if mc.GetAnnotations() == nil {
		return createdViaAnnotationOther
	}
	if a, ok := mc.GetAnnotations()[createdViaAnnotation]; ok {
		return createdViaMapping[a]
	}
	return createdViaAnnotationOther
}

func getServiceName(mc *mcv1.ManagedCluster) string {
	if mc.GetAnnotations() == nil {
		return serviceNameAnnotationOther
	}
	if a, ok := mc.GetAnnotations()[serviceNameAnnotation]; ok {
		return serviceNameMapping[a]
	}
	return serviceNameAnnotationOther
}
