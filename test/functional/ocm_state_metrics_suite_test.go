// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project


// +build functional

package functional

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	libgoclient "github.com/open-cluster-management/library-go/pkg/client"
	"github.com/open-cluster-management/library-go/pkg/templateprocessor"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	libgoapplier "github.com/open-cluster-management/library-go/pkg/applier"
)

const (
	// kubeConfig = "kind_kubeconfig.yaml"
	kubeConfig = ""
)

var (
	kubeClient            kubernetes.Interface
	defaultClient         client.Client
	clientDynamic         dynamic.Interface
	clientApplier         *libgoapplier.Applier
	gvrManagedclusterInfo schema.GroupVersionResource
)

func init() {
	klog.SetOutput(GinkgoWriter)
	klog.InitFlags(nil)
}

var _ = BeforeSuite(func() {
	gvrManagedclusterInfo = schema.GroupVersionResource{Group: "internal.open-cluster-management.io", Version: "v1beta1", Resource: "managedclusterinfos"}

	setupHub()

})

func TestOCMStateMEtrics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ocm-sate-metrics Suite")
}

func setupHub() {
	var err error
	kubeClient, err = libgoclient.NewDefaultKubeClient(kubeConfig)
	Expect(err).To(BeNil())
	defaultClient, err = libgoclient.NewDefaultClient(kubeConfig, client.Options{})
	Expect(err).To(BeNil())
	clientDynamic, err = libgoclient.NewDefaultKubeClientDynamic(kubeConfig)
	Expect(err).To(BeNil())

	yamlReader := templateprocessor.NewYamlFileReader("resources")
	clientApplier, err = libgoapplier.NewApplier(yamlReader,
		&templateprocessor.Options{},
		defaultClient,
		nil,
		nil,
		libgoapplier.DefaultKubernetesMerger,
		nil)
	Expect(err).To(BeNil())
	Expect(clientApplier.CreateOrUpdateInPath("cr", nil, false, nil)).To(BeNil())

}
