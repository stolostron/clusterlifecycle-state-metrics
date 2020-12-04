// Copyright (c) 2020 Red Hat, Inc.

// +build functional

package functional

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mciv1beta1 "github.com/open-cluster-management/multicloud-operators-foundation/pkg/apis/internal.open-cluster-management.io/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterDeploymentResponse = `# HELP clc_managedcluster_info Managed cluster information
# TYPE clc_managedcluster_info gauge
`
	managedClusterResponse = `# HELP clc_managedcluster_info Managed cluster information
# TYPE clc_managedcluster_info gauge
clc_managedcluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",cluster_id="import_cluster_id",vendor="OpenShift",cloud="Amazon",version="v1.16.2",created_via="Other"} 1
clc_managedcluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",cluster_id="local_cluster_id",vendor="OpenShift",cloud="Amazon",version="v1.16.2",created_via="Other"} 1
`
	managedClusterHiveResponse = `# HELP clc_managedcluster_info Managed cluster information
# TYPE clc_managedcluster_info gauge
clc_managedcluster_info{hub_cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",cluster_id="hive_cluster_id",vendor="OpenShift",cloud="Amazon",version="v1.16.2",created_via="Hive"} 1
`
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
			})).Should(BeNil())
		})
		By("Updating status cluster-import", func() {
			Expect(updateMCIStatus("cluster-import", mciv1beta1.ClusterInfoStatus{
				KubeVendor:  mciv1beta1.KubeVendorOpenShift,
				CloudVendor: mciv1beta1.CloudVendorAWS,
				Version:     "v1.16.2",
				ClusterID:   "import_cluster_id",
			})).Should(BeNil())
		})
		// Skip("Skip have to fix")
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
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
			})).Should(BeNil())
		})
		By("Getting metrics", func() {
			Eventually(func() string {
				resp, b, err := getMetrics()
				if err != nil || resp.StatusCode != http.StatusOK {
					return ""
				}
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
			log.Fatal(err)
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
