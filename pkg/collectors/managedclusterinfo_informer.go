// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	clusterclient "open-cluster-management.io/api/client/cluster/clientset/versioned"
	"time"

	"k8s.io/client-go/tools/cache"
	metricsstore "k8s.io/kube-state-metrics/pkg/metrics_store"
	mcv1 "open-cluster-management.io/api/cluster/v1"
)

// Create ManagedCluster informer, watch and update metrics
func createManagedClusterInformer(ctx context.Context, clusterClient clusterclient.Interface, store *metricsstore.MetricsStore) {
	lw := cache.NewListWatchFromClient(clusterClient.ClusterV1().RESTClient(), "managedclusters", metav1.NamespaceAll, fields.Everything())
	reflector := cache.NewReflector(lw, &mcv1.ManagedCluster{}, store, 60*time.Minute)

	go reflector.Run(ctx.Done())
}
