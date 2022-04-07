// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

const (
	clusterDeploymentResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
`
	managedClusterResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
`

	managedClusterHiveResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Hive",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
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
				ClusterClaims: []mcv1.ManagedClusterClaim{
					{
						Name:  "kubeversion.open-cluster-management.io",
						Value: "v1.16.2",
					},
				},
			})).Should(BeNil())
		})
		By("Updating status cluster-import", func() {
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
				ClusterClaims: []mcv1.ManagedClusterClaim{
					{
						Name:  "kubeversion.open-cluster-management.io",
						Value: "v1.16.2",
					},
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
				ClusterClaims: []mcv1.ManagedClusterClaim{
					{
						Name:  "kubeversion.open-cluster-management.io",
						Value: "v1.16.2",
					},
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

func updateMCStatus(name string, status mcv1.ManagedClusterStatus) error {
	mcU, err := clientDynamic.Resource(gvrManagedcluster).
		Get(context.TODO(), name, metav1.GetOptions{})
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

	sorted = unique(sorted)

	sorted.Sort()

	//fmt.Println("Sorted: ", sort.StringsAreSorted(sorted))

	return strings.Join(sorted, "\n")
}

func unique(s []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
