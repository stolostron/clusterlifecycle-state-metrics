// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

const (
	clusterDeploymentResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
`
	managedClusterResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1"} 1
`

	managedClusterHiveResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="2",socket_worker="1"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0"} 1
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

type QueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name             string `json:"__name__"`
				Available        string `json:"available"`
				Cloud            string `json:"cloud"`
				CoreWorker       string `json:"core_worker"`
				CreatedVia       string `json:"created_via"`
				Endpoint         string `json:"endpoint"`
				HubClusterID     string `json:"hub_cluster_id"`
				Instance         string `json:"instance"`
				Job              string `json:"job"`
				ManagedClusterID string `json:"managed_cluster_id"`
				Namespace        string `json:"namespace"`
				Pod              string `json:"pod"`
				Service          string `json:"service"`
				ServiceName      string `json:"service_name"`
				SocketWorker     string `json:"socket_worker"`
				Vendor           string `json:"vendor"`
				Version          string `json:"version"`
			} `json:"metric,omitempty"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

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

	AfterEach(func() {
		By("Query metrics by sum(acm_managed_cluster_info) by (hub_cluster_id)", func() {
			Eventually(func() error {
				resp, b, err := queryMetrics("sum(acm_managed_cluster_info)+by+(hub_cluster_id)")
				klog.Infof("Get Empty Metrics response: %v", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
				}

				queryResult := QueryResult{}
				if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
					return err
				}

				expectResult := map[string]string{"787e5a35-c911-4341-a2e7-65c415147aeb": "4"}
				actualResult := map[string]string{}
				for _, r := range queryResult.Data.Result {
					actualResult[r.Metric.HubClusterID] = r.Value[1].(string)
				}

				klog.Infof("expect result %v", expectResult)
				klog.Infof("actual result %v", actualResult)
				if !reflect.DeepEqual(expectResult, actualResult) {
					return fmt.Errorf("Unexpect queryResult %v", queryResult)
				}

				return nil
			}).Should(BeNil())
		})
		By("Query metrics by sum(acm_managed_cluster_info) by (version)", func() {
			Eventually(func() error {
				resp, b, err := queryMetrics("sum(acm_managed_cluster_info)+by+(version)")
				klog.Infof("Get Empty Metrics response: %v", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
				}

				queryResult := QueryResult{}
				if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
					return err
				}

				expectResult := map[string]string{"4.3.1": "3", "": "1"}
				actualResult := map[string]string{}
				for _, r := range queryResult.Data.Result {
					actualResult[r.Metric.Version] = r.Value[1].(string)
				}

				klog.Infof("expect result %v", expectResult)
				klog.Infof("actual result %v", actualResult)
				if !reflect.DeepEqual(expectResult, actualResult) {
					return fmt.Errorf("Unexpect queryResult %v", queryResult)
				}

				return nil
			}).Should(BeNil())
		})
		By("Query metrics by sum(acm_managed_cluster_info) by (cloud,created_via,vendor)", func() {
			Eventually(func() error {
				resp, b, err := queryMetrics("sum(acm_managed_cluster_info)+by+(cloud,created_via,vendor)")
				klog.Infof("Get Empty Metrics response: %v", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
				}

				queryResult := QueryResult{}
				if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
					return err
				}

				expectResult := map[string]string{"AmazonOtherOpenShift": "2", "AmazonHiveOpenShift": "1", "Other": "1"}
				actualResult := map[string]string{}
				for _, r := range queryResult.Data.Result {
					key := r.Metric.Cloud + r.Metric.CreatedVia + r.Metric.Vendor
					actualResult[key] = r.Value[1].(string)
				}

				klog.Infof("expect result %v", expectResult)
				klog.Infof("actual result %v", actualResult)
				if !reflect.DeepEqual(expectResult, actualResult) {
					return fmt.Errorf("Unexpect queryResult %v", queryResult)
				}

				return nil
			}).Should(BeNil())
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

func queryMetrics(promQL string) (resp *http.Response, bodyString string, err error) {
	promURL := "http://localhost/prometheus-k8s/api/v1/query?query=" + promQL
	klog.Infof("queryMetrics: %s", promURL)
	resp, err = http.Get(promURL)
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
