// Copyright (c) 2020 Red Hat, Inc.

package collectors

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func createManagedClusterInfoListWatch(apiserver string, kubeconfig string, ns string) cache.ListWatch {
	config, err := clientcmd.BuildConfigFromFlags(apiserver, kubeconfig)
	if err != nil {
		klog.Fatalf("cannot create Dynamic client: %v", err)
	}
	client := dynamic.NewForConfigOrDie(config)
	return createManagedClusterInfoListWatchWithClient(client, ns)
}
