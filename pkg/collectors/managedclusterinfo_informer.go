// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	clusterclient "github.com/open-cluster-management/api/client/cluster/clientset/versioned"
	clusterinformers "github.com/open-cluster-management/api/client/cluster/informers/externalversions"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
)

// Create ManagedCluster informer, watch and update metrics
func createManagedClusterInformer(apiserver string, kubeconfig string, store *metricsstore.MetricsStore) {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create config: %v", err)
	}

	clusterclient, err := clusterclient.NewForConfig(config)
	if err != nil {
		klog.Fatalf("cannot create clusterclient: %v", err)
	}

	clusterinformers := clusterinformers.NewSharedInformerFactory(clusterclient, 0)
	informer := clusterinformers.Cluster().V1().ManagedClusters()

	stopCh := make(chan struct{})
	go startWatchingManagedCluster(stopCh, informer.Informer(), store)
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
