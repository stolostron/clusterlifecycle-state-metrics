// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

//go:build functional
// +build functional

package functional

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	libgoapplier "github.com/stolostron/applier/pkg/applier"
	libgoclient "github.com/stolostron/library-go/pkg/client"
	libgoconfig "github.com/stolostron/library-go/pkg/config"
	"github.com/stolostron/library-go/pkg/templateprocessor"
	addonclient "open-cluster-management.io/api/client/addon/clientset/versioned"
	workclient "open-cluster-management.io/api/client/work/clientset/versioned"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// kubeConfig = "kind_kubeconfig.yaml"
	kubeConfig = ""
)

var (
	kubeClient        kubernetes.Interface
	defaultClient     client.Client
	clientDynamic     dynamic.Interface
	clientApplier     *libgoapplier.Applier
	gvrManagedcluster schema.GroupVersionResource
	addOnClient       addonclient.Interface
	workClient        workclient.Interface
)

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)
}

var _ = BeforeSuite(func() {
	gvrManagedcluster = schema.GroupVersionResource{Group: "cluster.open-cluster-management.io", Version: "v1", Resource: "managedclusters"}

	setupHub()

})

func TestOCMStateMEtrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ocm-state-metrics Suite")
}

func setupHub() {
	var err error
	kubeClient, err = libgoclient.NewDefaultKubeClient(kubeConfig)
	Expect(err).To(BeNil())
	defaultClient, err = libgoclient.NewDefaultClient(kubeConfig, client.Options{})
	Expect(err).To(BeNil())
	clientDynamic, err = libgoclient.NewDefaultKubeClientDynamic(kubeConfig)
	Expect(err).To(BeNil())

	config, err := libgoconfig.LoadConfig("", kubeConfig, "")
	Expect(err).To(BeNil())
	addOnClient, err = addonclient.NewForConfig(config)
	Expect(err).To(BeNil())

	workClient, err = workclient.NewForConfig(config)
	Expect(err).To(BeNil())

	yamlReader := templateprocessor.NewYamlFileReader("resources")
	clientApplier, err = libgoapplier.NewApplier(yamlReader,
		nil,
		defaultClient,
		nil,
		nil,
		nil)
	Expect(err).To(BeNil())
	Expect(clientApplier.CreateOrUpdateInPath("cr", nil, false, nil)).To(BeNil())

}
