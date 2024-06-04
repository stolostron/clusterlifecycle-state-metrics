// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
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
	clusterInfoDeploymentResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",product="",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1`

	clusterInfoManagedClusterResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",product="AKS",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",product="",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",product="OpenShift",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="2",socket_worker="1",hub_type="mce"} 1`

	clusterInfoManagedClusterHiveResponse = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",product="OpenShift",vendor="OpenShift",cloud="Amazon",service_name="Compute",version="4.3.1",available="Unknown",created_via="Hive",core_worker="2",socket_worker="1",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster",product="",vendor="",cloud="",service_name="Other",version="",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1
acm_managed_cluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",product="",vendor="OpenShift",cloud="Amazon",service_name="Other",version="4.3.1",available="Unknown",created_via="Other",core_worker="0",socket_worker="0",hub_type="mce"} 1`

	clusterLabelsResponse = `# HELP acm_managed_cluster_labels Managed cluster labels
# TYPE acm_managed_cluster_labels gauge
acm_managed_cluster_labels{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id",cloud="Amazon",openshiftVersion="4.3.1",vendor="OpenShift"} 1
acm_managed_cluster_labels{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id",cloud="Amazon",openshiftVersion="4.3.1",vendor="OpenShift"} 1
acm_managed_cluster_labels{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster"} 1
acm_managed_cluster_labels{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id",cloud="Amazon",openshiftVersion="4.3.1",vendor="OpenShift"} 1`

	clusterCountResponse = `# HELP acm_managed_cluster_count Managed cluster count
# TYPE acm_managed_cluster_count gauge
acm_managed_cluster_count 4`

	clusterStatusResponse = `# HELP acm_managed_cluster_status_condition Managed cluster status condition
# TYPE acm_managed_cluster_status_condition gauge
acm_managed_cluster_status_condition{managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="ManagedClusterConditionAvailable",status="true"} 0
acm_managed_cluster_status_condition{managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="ManagedClusterConditionAvailable",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="ManagedClusterConditionAvailable",status="unknown"} 1
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="hello",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="hello",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="hello",status="unknown"} 0
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="ManagedClusterConditionAvailable",status="true"} 0
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="ManagedClusterConditionAvailable",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="ManagedClusterConditionAvailable",status="unknown"} 1
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="hello",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="hello",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="hello",status="unknown"} 0
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="ManagedClusterConditionAvailable",status="true"} 0
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="ManagedClusterConditionAvailable",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="ManagedClusterConditionAvailable",status="unknown"} 1
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="hello",status="true"} 1
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="hello",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="hello",status="unknown"} 0
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="ManagedClusterConditionAvailable",status="true"} 0
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="ManagedClusterConditionAvailable",status="false"} 0
acm_managed_cluster_status_condition{managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="ManagedClusterConditionAvailable",status="unknown"} 1`

	clusterWorkerCoresResponse = `# HELP acm_managed_cluster_worker_cores The number of worker CPU cores of ACM managed clusters
# TYPE acm_managed_cluster_worker_cores gauge
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id"} 0
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id"} 0
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster"} 0
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id"} 0`

	clusterWorkerCoresUpdatedResponse = `# HELP acm_managed_cluster_worker_cores The number of worker CPU cores of ACM managed clusters
# TYPE acm_managed_cluster_worker_cores gauge
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="hive_cluster_id"} 5
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="import_cluster_id"} 0
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="empty-cluster"} 0
acm_managed_cluster_worker_cores{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",managed_cluster_id="local_cluster_id"} 0`
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
				Name               string `json:"__name__"`
				Available          string `json:"available"`
				Cloud              string `json:"cloud"`
				CoreWorker         string `json:"core_worker"`
				CreatedVia         string `json:"created_via"`
				Endpoint           string `json:"endpoint"`
				HubClusterID       string `json:"hub_cluster_id"`
				Instance           string `json:"instance"`
				Job                string `json:"job"`
				ManagedClusterID   string `json:"managed_cluster_id"`
				ManagedClusterName string `json:"managed_cluster_name"`
				Namespace          string `json:"namespace"`
				Pod                string `json:"pod"`
				Service            string `json:"service"`
				ServiceName        string `json:"service_name"`
				SocketWorker       string `json:"socket_worker"`
				Vendor             string `json:"vendor"`
				Version            string `json:"version"`
			} `json:"metric,omitempty"`
			Value []interface{} `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

var _ = Describe("ManagedCluster Metrics", func() {
	BeforeEach(func() {
		SetDefaultEventuallyTimeout(60 * time.Second)
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

	assertGetMetrics := func(metricName, response string) {
		By("Getting metrics", func() {
			Eventually(func() int {
				resp, b, err := getMetric(metricName)
				klog.Infof("Get metrics response: %s", b)
				if err != nil || resp.StatusCode != http.StatusOK {
					return 0
				}
				return len(b)
			}).Should(Equal(len(response)))
		})
	}

	assertRecordingRule := func(record, expr string) {
		var re = regexp.MustCompile(`[0-9]{10}\.[0-9]{3}`)
		By("Querying recording rule ...", func() {
			Eventually(func() bool {
				resp, recordBody, err := queryMetrics(record)
				if err != nil || resp.StatusCode != http.StatusOK {
					return false
				}
				recordBody = strings.Replace(recordBody, fmt.Sprintf("\"__name__\":\"%s\",", record), "", -1)
				recordBody = re.ReplaceAllString(recordBody, `0000000000.000`)
				klog.Infof("Querying record %s response: %s", record, recordBody)

				resp, exprBody, err := queryMetrics(expr)
				if err != nil || resp.StatusCode != http.StatusOK {
					return false
				}
				exprBody = re.ReplaceAllString(exprBody, `0000000000.000`)
				klog.Infof("Querying with expr %s response: %s", expr, exprBody)

				return recordBody == exprBody
			}).Should(BeTrue())
		})
	}

	Context("acm_managed_cluster_info", func() {
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
				}).WithTimeout(30 * time.Second).Should(BeNil())
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

		It("Get Metrics", func() {
			assertGetMetrics("acm_managed_cluster_info", clusterInfoDeploymentResponse)
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
						{
							Name:  "product.open-cluster-management.io",
							Value: "OpenShift",
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
						{
							Name:  "product.open-cluster-management.io",
							Value: "AKS",
						},
					},
				})).Should(BeNil())
			})

			assertGetMetrics("acm_managed_cluster_info", clusterInfoManagedClusterResponse)
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
						{
							Name:  "product.open-cluster-management.io",
							Value: "OpenShift",
						},
					},
				})).Should(BeNil())
			})

			assertGetMetrics("acm_managed_cluster_info", clusterInfoManagedClusterHiveResponse)
		})
	})

	Context("acm_managed_cluster_labels", func() {
		AfterEach(func() {
			By("Query metrics by sum(acm_managed_cluster_labels) by (cloud)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("sum(acm_managed_cluster_labels)+by+(cloud)")
					klog.Infof("Get Empty Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := QueryResult{}
					if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"Amazon": "3", "": "1"}
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
			By("Query metrics by sum(acm_managed_cluster_labels) by (hub_cluster_id)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("sum(acm_managed_cluster_labels)+by+(hub_cluster_id)")
					klog.Infof("Get Empty Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := QueryResult{}
					if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"": "4"}
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

		It("Get Metrics", func() {
			assertGetMetrics("acm_managed_cluster_labels", clusterLabelsResponse)
		})
	})

	Context("acm_managed_cluster_status_condition", func() {
		AfterEach(func() {
			By("Query metrics by sum(acm_managed_cluster_status_condition) by (managed_cluster_name)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("sum(acm_managed_cluster_status_condition)+by+(managed_cluster_name)")
					klog.Infof("Get Empty Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := QueryResult{}
					if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"empty-cluster": "1", "cluster-hive": "2",
						"cluster-import": "2", "local-cluster": "2"}
					actualResult := map[string]string{}
					for _, r := range queryResult.Data.Result {
						key := r.Metric.ManagedClusterName
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

		It("Get Metrics", func() {
			assertGetMetrics("acm_managed_cluster_status_condition", clusterStatusResponse)
		})
	})

	Context("acm_managed_cluster_count", func() {
		AfterEach(func() {
			By("Query metrics of acm_managed_cluster_count", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("acm_managed_cluster_count")
					klog.Infof("Get Empty Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := &QueryResult{}
					if err := json.Unmarshal([]byte(b), queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"acm_managed_cluster_count": "4"}
					actualResult := map[string]string{}
					for _, r := range queryResult.Data.Result {
						key := r.Metric.Name
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

		It("Get Metrics", func() {
			assertGetMetrics("acm_managed_cluster_count", clusterCountResponse)
		})
	})

	Context("acm_managed_cluster_worker_cores", func() {
		ruleExpr := "max(acm_managed_cluster_worker_cores)+by+(hub_cluster_id,managed_cluster_id)"

		AfterEach(func() {
			By("Query metrics by count(acm_managed_cluster_worker_cores) by (hub_cluster_id)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("count(acm_managed_cluster_worker_cores)+by+(hub_cluster_id)")
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
				}).WithTimeout(30 * time.Second).Should(BeNil())
			})
		})

		It("should get metric", func() {
			assertGetMetrics("acm_managed_cluster_worker_cores", clusterWorkerCoresResponse)
			assertRecordingRule("acm_managed_cluster_worker_cores:max", ruleExpr)
		})

		It("should reflect the change on the managed cluster", func() {
			By("Updating status cluster-hive", func() {
				Expect(updateMCStatus("cluster-hive", mcv1.ManagedClusterStatus{
					Capacity: mcv1.ResourceList{
						resourceCoreWorker: *resource.NewQuantity(5, resource.DecimalSI),
					},
				})).Should(BeNil())
			})

			assertGetMetrics("acm_managed_cluster_worker_cores", clusterWorkerCoresUpdatedResponse)
			assertRecordingRule("acm_managed_cluster_worker_cores:max", ruleExpr)
		})
	})
})

func getMetric(metricName string) (resp *http.Response, bodyString string, err error) {
	resp, err = http.Get("http://localhost/clusterlifecycle-state-metrics/metrics")
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			klog.Fatal(err)
		}

		var sb strings.Builder
		scanner := bufio.NewScanner(strings.NewReader(string(bodyBytes)))
		for scanner.Scan() {
			if line := scanner.Text(); strings.Contains(line, metricName) {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
				sb.WriteString(line)
			}
		}
		bodyString = sb.String()
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
