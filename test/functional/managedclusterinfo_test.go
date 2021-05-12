// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

// +build functional

package functional

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mcv1 "github.com/open-cluster-management/api/cluster/v1"
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
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",status="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",status="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",status="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",status="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
`
	managedClusterHiveResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",status="Unknown",created_via="Hive",core_worker="2",socket_worker="1"} 1
`
)

const (
	masterLabel = "node-role.kubernetes.io/master"
	workerLabel = "node-role.kubernetes.io/worker"

	resourceSocket       mcv1.ResourceName = "socket"
	resourceCore         mcv1.ResourceName = "core"
	resourceCoreWorker   mcv1.ResourceName = "core_worker"
	resourceSocketWorker mcv1.ResourceName = "socket_worker"
	resourceCPUWorker    mcv1.ResourceName = "cpu_worker"
)

var _ = Describe("Metrics", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(20 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)
		By("Cleaning status", func() {
			Expect(updateMCIStatus("local-cluster", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
			Expect(updateMCIStatus("cluster-hive", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
			Expect(updateMCIStatus("cluster-import", mciv1beta1.ClusterInfoStatus{})).Should(BeNil())
			Expect(updateMCStatus("local-cluster", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
						Message:            "hello",
					},
				},
			})).Should(BeNil())
			Expect(updateMCStatus("cluster-hive", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
						Message:            "hello",
					},
				},
			})).Should(BeNil())
			Expect(updateMCStatus("cluster-import", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
						Message:            "hello",
					},
				},
			})).Should(BeNil())
		})
	})

	It("Get Empty Metrics", func() {
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				klog.Infof("Get Empty Metrics response: %s", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				return sortLines(b)
			}).Should(Equal(sortLines(clusterDeploymentResponse)))
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
				// Lable worker with CPU
				NodeList: []mciv1beta1.NodeStatus{
					{
						Name: "worker-3",
						Labels: map[string]string{
							workerLabel: "",
						},
					},
				},
			})).Should(BeNil())
			Expect(updateMCStatus("local-cluster", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
						Message:            "hello",
					},
				},
				Capacity: mcv1.ResourceList{
					resourceCoreWorker:   *resource.NewQuantity(2, resource.DecimalSI),
					resourceSocketWorker: *resource.NewQuantity(1, resource.DecimalSI),
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
				// Lable worker with CPU
				NodeList: []mciv1beta1.NodeStatus{
					{
						Name: "worker-3",
						Labels: map[string]string{
							workerLabel: "",
						},
					},
				},
			})).Should(BeNil())
			Expect(updateMCStatus("cluster-import", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
						Message:            "hello",
					},
				},
				Capacity: mcv1.ResourceList{
					resourceCoreWorker:   *resource.NewQuantity(2, resource.DecimalSI),
					resourceSocketWorker: *resource.NewQuantity(1, resource.DecimalSI),
				},
			})).Should(BeNil())
		})
		// Skip("Skip have to fix")
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				klog.Infof("Get Metrics response: %s", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				return sortLines(b)
			}).Should(Equal(sortLines(managedClusterResponse)))
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
					// Label master with vCPU
					{
						Name: "master",
						Labels: map[string]string{
							masterLabel: "",
						},
					},
					// Label worker with vCPU
					{
						Name: "worker-3",
						Labels: map[string]string{
							workerLabel: "",
						},
					},
				},
			})).Should(BeNil())
			Expect(updateMCStatus("cluster-hive", mcv1.ManagedClusterStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "hello",
						Status:             "True",
						LastTransitionTime: metav1.Now(),
						Reason:             "test",
					},
				},
				Capacity: mcv1.ResourceList{
					resourceCoreWorker:   *resource.NewQuantity(2, resource.DecimalSI),
					resourceSocketWorker: *resource.NewQuantity(1, resource.DecimalSI),
				},
			})).Should(BeNil())
		})
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				klog.Infof("Get created cluster-hive Metrics response: %s", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
				return sortLines(b)
			}).Should(Equal(sortLines(managedClusterHiveResponse)))
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
	mciU, err := clientDynamic.Resource(gvrManagedclusterInfo).Namespace(name).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	b, err := json.Marshal(status)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("MCI Status: %s", string(b))
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	mciU.Object["status"] = m
	_, err = clientDynamic.Resource(gvrManagedclusterInfo).
		Namespace(name).
		UpdateStatus(context.TODO(), mciU, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	// klog.Infof("UpdateMCIStatus: %v", mciU)
	return err
}

func updateMCStatus(name string, status mcv1.ManagedClusterStatus) error {
	mcU, err := clientDynamic.Resource(gvrManagedcluster).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	b, err := json.Marshal(status)
	if err != nil {
		klog.Error(err)
		return err
	}
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("Status: %s", string(b))
	mcU.Object["status"] = m
	_, err = clientDynamic.Resource(gvrManagedcluster).
		UpdateStatus(context.TODO(), mcU, metav1.UpdateOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	mcU, err = clientDynamic.Resource(gvrManagedcluster).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.Infof("UpdateMCStatus updated: %v", mcU)
	return err
}

func sortLines(input string) string {
	var sorted sort.StringSlice

	sorted = strings.Split(input, "\n") // convert to slice

	// just for fun
	//fmt.Println("Sorted: ", sort.StringsAreSorted(sorted))

	sorted.Sort()

	//fmt.Println("Sorted: ", sort.StringsAreSorted(sorted))

	return strings.Join(sorted, "\n")
}
