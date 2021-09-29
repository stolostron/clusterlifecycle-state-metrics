// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"golang.org/x/net/context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// Create ManagedClusterInfo informer, watch and update metrics
func createManagedClusterInfoInformer(apiserver string, kubeconfig string, ns string, store *metricsstore.MetricsStore) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)

	sharedInformers := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, ns, nil)
	informer := sharedInformers.ForResource(mciGVR)

	stopCh := make(chan struct{})
	go startWatchingManagedClusterInfo(stopCh, client, informer.Informer(), store)
}

// Create ManagedCluster informer, watch and update metrics
func createManagedClusterInformer(apiserver string, kubeconfig string, store *metricsstore.MetricsStore) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)

	sharedInformers := dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, 0, v1.NamespaceAll, nil)
	informer := sharedInformers.ForResource(mcGVR)

	stopCh := make(chan struct{})
	go startWatchingManagedCluster(stopCh, informer.Informer(), store)
}

func startWatchingManagedClusterInfo(stopCh <-chan struct{}, client dynamic.Interface, s cache.SharedIndexInformer, store *metricsstore.MetricsStore) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			objMC := findManagedCluster(obj, client)
			store.Add(objMC)
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			objMC := findManagedCluster(obj, client)
			store.Update(objMC)
		},
		DeleteFunc: func(obj interface{}) {
			objMC := findManagedCluster(obj, client)
			store.Delete(objMC)
		},
	}
	s.AddEventHandler(handlers)
	s.Run(stopCh)
}

func startWatchingManagedCluster(stopCh <-chan struct{}, s cache.SharedIndexInformer, store *metricsstore.MetricsStore) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			store.Add(obj)
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			store.Update(obj)
		},
		DeleteFunc: func(obj interface{}) {
			store.Delete(obj)
		},
	}
	s.AddEventHandler(handlers)
	s.Run(stopCh)
}

func findManagedCluster(obj interface{}, client dynamic.Interface) interface{} {
	o, err := meta.Accessor(obj)
	if err != nil {
		klog.Warningf("Failed to get ManagedCluster: %s", err)
		return nil
	}

	mc, err := client.Resource(mcGVR).Get(context.TODO(), o.GetName(), metav1.GetOptions{})
	if err != nil {
		klog.Warningf("Failed to get ManagedCluster: %s", err)
		return nil
	}

	return mc
}
