package collectors

import (
	"context"

	ocinfrav1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kube-state-metrics/pkg/metric"

	clientset "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	"k8s.io/klog/v2"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
)

var (
	descClusterInfoName          = "ocm_managedcluster_info"
	descClusterInfoHelp          = "Kubernetes labels converted to Prometheus labels."
	descClusterInfoDefaultLabels = []string{"cluster_id", "cluster_domain", "managedcluster_name", "vendor", "cloud", "version"}

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

	infraGVR = schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "infrastructures",
	}
)

func getHubClusterId(c dynamic.Interface) string {

	cvObj, errCv := c.Resource(cvGVR).Get(context.TODO(), "version", metav1.GetOptions{})
	if errCv != nil {
		klog.Fatalf("Error getting cluster version %v \n", errCv)
		panic(errCv.Error())
	}
	cv := &ocinfrav1.ClusterVersion{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(cvObj.UnstructuredContent(), &cv)
	if err != nil {
		klog.Fatalf("Error unmarshal cluster version object%v \n", err)
		panic(errCv.Error())
	}
	return string(cv.Spec.ClusterID)

}

func getHubClusterDomain(c dynamic.Interface) string {

	infraObj, errInfra := c.Resource(infraGVR).Get(context.TODO(), "cluster", metav1.GetOptions{})
	if errInfra != nil {
		klog.Fatalf("Error getting infrastructure cluster %v \n", errInfra)
		panic(errInfra.Error())
	}
	infra := &ocinfrav1.Infrastructure{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(infraObj.UnstructuredContent(), &infra)
	if err != nil {
		klog.Fatalf("Error unmarshal infrastructure cluster object%v \n", err)
		panic(errInfra.Error())
	}
	return infra.Status.EtcdDiscoveryDomain

}

func getManagedClusterMetricFamilies(client dynamic.Interface) []metric.FamilyGenerator {
	hubID := getHubClusterId(client)
	hubDomain := getHubClusterDomain(client)
	return []metric.FamilyGenerator{
		// {
		// 	Name: "ocm_managedcluster_created",
		// 	Type: metric.MetricTypeGauge,
		// 	Help: "Unix creation timestamp",
		// 	GenerateFunc: wrapManagedClusterFunc(func(mc *managedclusterv1.ManagedCluster) metric.Family {
		// 		f := metric.Family{}

		// 		if !mc.CreationTimestamp.IsZero() {
		// 			f.Metrics = append(f.Metrics, &metric.Metric{
		// 				Value: float64(mc.CreationTimestamp.Unix()),
		// 			})
		// 		}

		// 		return f
		// 	}),
		// },
		// Read the clusterdeployment to define if hive or imported
		// {
		// 	Name: descClusterLabelsName,
		// 	Type: metric.MetricTypeGauge,
		// 	Help: descClusterLabelsHelp,
		// 	GenerateFunc: wrapManagedClusterFunc(func(mc *managedclusterv1.ManagedCluster) metric.Family {
		// 		createdVia := "hive"
		// 		_, err := client.Resource(cdGVR).Namespace(mc.GetName()).Get(context.TODO(), mc.GetName(), metav1.GetOptions{})
		// 		if errors.IsNotFound(err) {
		// 			createdVia = "imported"
		// 		}
		// 		labels := mc.GetLabels()
		// 		labelsValues := []string{hubID, hubDomain, mc.Name, labels["vendor"], labels["cloud"], createdVia, mc.Status.Version.Kubernetes}
		// 		return metric.Family{Metrics: []*metric.Metric{
		// 			{
		// 				LabelKeys:   descClusterLabelsDefaultLabels,
		// 				LabelValues: labelsValues,
		// 				Value:       1,
		// 			},
		// 		}}
		// 	}),
		// },
		//Does not read the clusterdeployment
		{
			Name: descClusterInfoName,
			Type: metric.MetricTypeGauge,
			Help: descClusterInfoHelp,
			GenerateFunc: wrapManagedClusterFunc(func(mc *managedclusterv1.ManagedCluster) metric.Family {
				labels := mc.GetLabels()
				labelsValues := []string{hubID, hubDomain, mc.Name, labels["vendor"], labels["cloud"], mc.Status.Version.Kubernetes}
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

func wrapManagedClusterFunc(f func(*managedclusterv1.ManagedCluster) metric.Family) func(interface{}) metric.Family {
	return func(obj interface{}) metric.Family {
		Cluster := obj.(*managedclusterv1.ManagedCluster)

		metricFamily := f(Cluster)

		for _, m := range metricFamily.Metrics {
			m.LabelKeys = append([]string{}, m.LabelKeys...)
			m.LabelValues = append([]string{}, m.LabelValues...)
		}

		return metricFamily
	}
}

func createManagedClusterListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	managedclusterclient, err := createManagedClusterClient(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create ManagedCluster client: %v", err)
	}
	return cache.ListWatch{
		ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
			return managedclusterclient.ClusterV1().ManagedClusters().List(context.TODO(), opts)
		},
		WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
			return managedclusterclient.ClusterV1().ManagedClusters().Watch(context.TODO(), opts)
		},
	}
}

func createManagedClusterClient(apiserver string, kubeconfig string) (*clientset.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	return client, err

}
