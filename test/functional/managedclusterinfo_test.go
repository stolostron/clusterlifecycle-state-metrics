// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// +build functional

package functional

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	clusterDeploymentResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
`
	managedClusterResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",created_via="Other",vcpu="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",created_via="Other",vcpu="0"} 1
`
	managedClusterHiveResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",created_via="Hive",vcpu="3"} 1
`
)

const (
	workerLabel = "node-role.kubernetes.io/worker"
)

var _ = Describe("Metrics", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(20 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)
		By("Cleaning status", func() {
			Expect(updateMCIStatus("local-cluster", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
			Expect(updateMCIStatus("cluster-hive", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
			Expect(updateMCIStatus("cluster-import", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
		})
	})

	It("Get Empty Metrics", func() {
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				klog.Infof("Get Empty Metrics response: %s", b)
				return b
			}).Should(Equal(clusterDeploymentResponse))
		})
	})

	It("Get imported cluster Metrics", func() {
		By("Updating status local-cluster", func() {
			Expect(updateMCIStatus("local-cluster", mciv1beta1.ClusterInfoStatus{
				KubeVendor:  mciv1beta1.KubeVendorOpenShift,
				CloudVendor: mciv1beta1.CloudVendorAWS,
				Version:     "v1.16.2",
				ClusterID:   "local_cluster_id",
				DistributionInfo: mciv1beta1.DistributionInfo{
					Type: mciv1beta1.DistributionTypeOCP,
					OCP: mciv1beta1.OCPDistributionInfo{
						Version: "4.3.1",
					},
				},
			})).Should(BeNil())
		})
		By("Updating status cluster-import", func() {
			Expect(updateMCIStatus("cluster-import", mciv1beta1.ClusterInfoStatus{
				KubeVendor:  mciv1beta1.KubeVendorOpenShift,
				CloudVendor: mciv1beta1.CloudVendorAWS,
				Version:     "v1.16.2",
				ClusterID:   "import_cluster_id",
				DistributionInfo: mciv1beta1.DistributionInfo{
					Type: mciv1beta1.DistributionTypeOCP,
					OCP: mciv1beta1.OCPDistributionInfo{
						Version: "4.3.1",
					},
				},
			})).Should(BeNil())
		})
		// Skip("Skip have to fix")
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				klog.Infof("Get Metrics response: %s", b)
				return b
			}).Should(Equal(managedClusterResponse))
		})
	})

	It("Get created cluster-hive Metrics", func() {
		By("Updating status cluster-hive", func() {
			Expect(updateMCIStatus("cluster-hive", mciv1beta1.ClusterInfoStatus{
				KubeVendor:  mciv1beta1.KubeVendorOpenShift,
				CloudVendor: mciv1beta1.CloudVendorAWS,
				Version:     "v1.16.2",
				ClusterID:   "hive_cluster_id",
				DistributionInfo: mciv1beta1.DistributionInfo{
					Type: mciv1beta1.DistributionTypeOCP,
					OCP: mciv1beta1.OCPDistributionInfo{
						Version: "4.3.1",
					},
				},
				NodeList: []mciv1beta1.NodeStatus{
					// Label worker with vCPU
					mciv1beta1.NodeStatus{
						Name: "worker-3",
						Labels: map[string]string{
							workerLabel: "",
						},
						Capacity: mciv1beta1.ResourceList{
							mciv1beta1.ResourceCPU: *resource.NewQuantity(1, resource.DecimalSI),
						},
					},
					// Label worker with vCPU
					mciv1beta1.NodeStatus{
						Name: "worker-3",
						Labels: map[string]string{
							workerLabel: "",
						},
						Capacity: mciv1beta1.ResourceList{
							mciv1beta1.ResourceCPU: *resource.NewQuantity(2, resource.DecimalSI),
						},
					},
				},
			})).Should(BeNil())
		})
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				klog.Infof("Get created cluster-hive Metrics response: %s", b)
				return b
			}).Should(Equal(managedClusterHiveResponse))
		})
	})
})

func getMetrics() (resp *http.Response, bodyString string, err error) {
	resp, err = http.Get("http://localhost/clusterlifecycle-state-metrics/metrics")
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Fatal(err)
		}
		bodyString = string(bodyBytes)
	}
	return
}

func updateMCIStatus(name string, status mciv1beta1.ClusterInfoStatus) error {
	mciLocalCluster, err := clientDynamic.Resource(gvrManagedclusterInfo).
		Namespace(name).
		Get(context.TODO(), name, metav1.GetOptions{})
	Expect(err).To(BeNil())
	mciLocalCluster.Object["status"] = status
	_, err = clientDynamic.Resource(gvrManagedclusterInfo).
		Namespace(name).
		UpdateStatus(context.TODO(), mciLocalCluster, metav1.UpdateOptions{})
	return err
}
