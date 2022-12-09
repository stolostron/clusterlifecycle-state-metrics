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
	"k8s.io/klog"
	addonv1alpha1 "open-cluster-management.io/api/addon/v1alpha1"
)

const (
	addOnStatusResponse = `# HELP acm_managed_cluster_addon_status_condition Managed cluster add-on status condition
# TYPE acm_managed_cluster_addon_status_condition gauge
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="empty-cluster",managed_cluster_name="empty-cluster",condition="Available",status="unknown"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="local_cluster_id",managed_cluster_name="local-cluster",condition="Available",status="true"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="hive_cluster_id",managed_cluster_name="cluster-hive",condition="Available",status="unknown"} 1
acm_managed_cluster_addon_status_condition{addon_name="work-manager",managed_cluster_id="import_cluster_id",managed_cluster_name="cluster-import",condition="Available",status="false"} 1`
)

var _ = Describe("ManagedClusterAddOn Metrics", func() {
	addOnName := "work-manager"
	clusterNames := []string{"empty-cluster", "local-cluster", "cluster-hive", "cluster-import"}

	BeforeEach(func() {
		SetDefaultEventuallyTimeout(60 * time.Second)
		SetDefaultEventuallyPollingInterval(1 * time.Second)

		// create addons for clusters
		err := newManagedClusterAddOn(addOnName, "empty-cluster", nil)
		Expect(err).To(BeNil())

		err = newManagedClusterAddOn(addOnName, "local-cluster", []metav1.Condition{
			{
				Type:               "Available",
				Status:             metav1.ConditionTrue,
				Message:            "Available",
				Reason:             "Available",
				LastTransitionTime: metav1.Now(),
			},
		})
		Expect(err).To(BeNil())

		err = newManagedClusterAddOn(addOnName, "cluster-hive", nil)
		Expect(err).To(BeNil())

		err = newManagedClusterAddOn(addOnName, "cluster-import", []metav1.Condition{
			{
				Type:               "Available",
				Status:             metav1.ConditionFalse,
				Message:            "NotAvailable",
				Reason:             "NotAvailable",
				LastTransitionTime: metav1.Now(),
			},
		})
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		for _, clusterName := range clusterNames {
			err := addOnClient.AddonV1alpha1().ManagedClusterAddOns(clusterName).Delete(context.TODO(), addOnName, metav1.DeleteOptions{})
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

	Context("acm_managed_cluster_addon_status_condition", func() {
		AfterEach(func() {
			By("Query metrics by sum(acm_managed_cluster_addon_status_condition) by (managed_cluster_name)", func() {
				Eventually(func() error {
					resp, b, err := queryMetrics("sum(acm_managed_cluster_addon_status_condition)+by+(managed_cluster_name)")
					klog.Infof("Get Metrics response: %v", b)
					if err != nil || resp.StatusCode != http.StatusOK {
						return fmt.Errorf("Failed to query metrics %v: %v", resp.StatusCode, err)
					}

					queryResult := QueryResult{}
					if err := json.Unmarshal([]byte(b), &queryResult); err != nil {
						return err
					}

					expectResult := map[string]string{"empty-cluster": "1", "cluster-hive": "1",
						"cluster-import": "1", "local-cluster": "1"}
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
			assertGetMetrics("acm_managed_cluster_addon_status_condition", addOnStatusResponse)
		})
	})
})

func newManagedClusterAddOn(addOnName, clusterName string, conditions []metav1.Condition) error {
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

	// create addon if not exists
	addOn, err := addOnClient.AddonV1alpha1().ManagedClusterAddOns(clusterName).Get(context.TODO(), addOnName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		addOn = &addonv1alpha1.ManagedClusterAddOn{
			ObjectMeta: metav1.ObjectMeta{
				Name:      addOnName,
				Namespace: clusterName,
			},
		}
		addOn, err = addOnClient.AddonV1alpha1().ManagedClusterAddOns(clusterName).Create(context.TODO(), addOn, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	// update addon status if necessary
	addOn.Status.Conditions = conditions
	_, err = addOnClient.AddonV1alpha1().ManagedClusterAddOns(clusterName).UpdateStatus(context.TODO(), addOn, metav1.UpdateOptions{})
	return err
}
