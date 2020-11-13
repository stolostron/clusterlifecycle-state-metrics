package collectors

import (
	"context"

	ocinfrav1 "github.com/openshift/api/config/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"

	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/klog/v2"
)

var (
	descClusterInfoName          = "ocm_managedcluster_info"
	descClusterInfoHelp          = "Managed cluster information"
	descClusterInfoDefaultLabels = []string{"hub_cluster_id", "cluster_id", "cluster", "vendor", "cloud", "version"}

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

func getManagedClusterInfoMetricFamilies(client dynamic.Interface) []metric.FamilyGenerator {
	hubClusterID := getHubClusterID(client)
	return []metric.FamilyGenerator{
		{
			Name: descClusterInfoName,
			Type: metric.MetricTypeGauge,
			Help: descClusterInfoHelp,
			GenerateFunc: wrapManagedClusterInfoFunc(func(mciObj *unstructured.Unstructured) metric.Family {
				mci := &mciv1beta1.ManagedClusterInfo{}
				err := runtime.DefaultUnstructuredConverter.FromUnstructured(mciObj.UnstructuredContent(), &mci)
				if err != nil {
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				if mci.Status.ClusterID == "" ||
					mci.Status.KubeVendor == "" ||
					mci.Status.CloudVendor == "" ||
					mci.Status.Version == "" {
					return metric.Family{Metrics: []*metric.Metric{}}
				}
				labelsValues := []string{hubClusterID,
					mci.Status.ClusterID,
					mci.GetName(),
					string(mci.Status.KubeVendor),
					string(mci.Status.CloudVendor),
					mci.Status.Version}
				return metric.Family{Metrics: []*metric.Metric{
					{
						LabelKeys:   descClusterInfoDefaultLabels,
						LabelValues: labelsValues,
						Value:       1,
					},
				}}
			}),
		},
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

func createManagedClusterInfoListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)
	return createManagedClusterInfoListWatchWithClient(client, ns)
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
