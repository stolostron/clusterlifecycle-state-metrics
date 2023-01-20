// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog"
	workv1 "open-cluster-management.io/api/work/v1"
)

const (
	workStatusResponse = `# HELP acm_manifestwork_status_condition ManifestWork status condition
# TYPE acm_manifestwork_status_condition gauge
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Applied",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Applied",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Available",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Available",status="false"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Available",status="unknown"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Applied",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Applied",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Available",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Available",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Available",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Applied",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Applied",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Available",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Available",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Available",status="unknown"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Applied",status="true"} 1
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Applied",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Applied",status="unknown"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Available",status="true"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Available",status="false"} 0
acm_manifestwork_status_condition{manifestwork="test-work1",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Available",status="unknown"} 1`

	workCountResponse = `# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
acm_manifestwork_count 4`
)

var _ = Describe("ManifestWork Metrics", func() {
	workName := "test-work1"
	clusterNames := []string{"empty-cluster", "local-cluster", "cluster-hive", "cluster-import"}

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(60 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)

		// create manifestworks for clusters
		configmap := &corev1.Namespace{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("test-cm-%s", rand.String(5)),
				Namespace: "default",
			},
		}

		err := newManifestWork(workName, "empty-cluster", nil, configmap)
		Expect(err).To(BeNil())

		err = newManifestWork(workName, "local-cluster", []metav1.Condition{
			{
				Type:               "Applied",
				Status:             metav1.ConditionTrue,
				Message:            "Applied",
				Reason:             "Applied",
				LastTransitionTime: metav1.Now(),
			},
		}, configmap)
		Expect(err).To(BeNil())

		err = newManifestWork(workName, "cluster-hive", []metav1.Condition{
			{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				Message:            "NotAvailable",
				Reason:             "NotAvailable",
				LastTransitionTime: metav1.Now(),
			},
		}, configmap)
		Expect(err).To(BeNil())

		err = newManifestWork(workName, "cluster-import", nil, configmap)
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		for _, clusterName := range clusterNames {
			err := workClient.WorkV1().ManifestWorks(clusterName).Delete(context.TODO(), workName, metav1.DeleteOptions{})
			Expect(err).To(BeNil())
		}
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

	Context("acm_manifestwork_status_condition", func() {
		AfterEach(func() {
			By("Query metrics by sum(acm_manifestwork_status_condition) by (managed_cluster_name)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("sum(acm_manifestwork_status_condition)+by+(managed_cluster_name)")
					klog.Infof("Get Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := QueryResult{}
					if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"empty-cluster": "2", "cluster-hive": "2",
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
			assertGetMetrics("acm_manifestwork_status_condition", workStatusResponse)
		})
	})

	Context("acm_manifestwork_count", func() {
		AfterEach(func() {
			By("Query metrics of acm_manifestwork_count", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("acm_manifestwork_count")
					klog.Infof("Get Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := &QueryResult{}
					if err := json.Unmarshal([]byte(b), queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"acm_manifestwork_count": "4"}
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
})

func newManifestWork(workName, clusterName string, conditions []metav1.Condition, objects ...runtime.Object) error {
	// create cluster namespace if necessary
	_, err := kubeClient.CoreV1().Namespaces().Get(context.TODO(), clusterName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: clusterName,
			},
		}
		_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	// create manifestwork if not exists
	work, err := workClient.WorkV1().ManifestWorks(clusterName).Get(context.TODO(), workName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		manifests := []workv1.Manifest{}
		for _, object := range objects {
			manifests = append(manifests, workv1.Manifest{
				RawExtension: runtime.RawExtension{Object: object},
			})
		}

		work = &workv1.ManifestWork{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      workName,
				Namespace: clusterName,
			},
			Spec: workv1.ManifestWorkSpec{
				Workload: workv1.ManifestsTemplate{
					Manifests: manifests,
				},
			},
		}
		work, err = workClient.WorkV1().ManifestWorks(clusterName).Create(context.TODO(), work, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	// update manifestwork status if necessary
	work.Status.Conditions = conditions
	_, err = workClient.WorkV1().ManifestWorks(clusterName).UpdateStatus(context.TODO(), work, metav1.UpdateOptions{})
	return err
}
