// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package options

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	koptions "k8s.io/kube-state-metrics/pkg/options"
)

func init() {
	//TODO this is because the CollectorSet struct is validate the collectors from the commandline using
	//"DefaultCollectors". https://github.com/kubernetes/kube-state-metrics/blob/master/pkg/options/types.go#L80
	koptions.DefaultCollectors["managedclusterinfos"] = struct{}{}
}

var (
	DefaultNamespaces = koptions.NamespaceList{metav1.NamespaceAll}
	DefaultCollectors = koptions.CollectorSet{
		"managedclusters":      struct{}{},
		"managedclusteraddons": struct{}{},
		"manifestworks":        struct{}{},
	}
)
