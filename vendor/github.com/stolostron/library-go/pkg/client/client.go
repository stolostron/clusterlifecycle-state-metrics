// Copyright Contributors to the Open Cluster Management project

package client

import (
	"fmt"

	"github.com/stolostron/library-go/pkg/config"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	libgocrdv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/crd"
	libgodeploymentv1 "github.com/stolostron/library-go/pkg/apis/meta/v1/deployment"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

//NewDefaultClient returns a client.Client for the current-context in kubeconfig
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
func NewDefaultClient(kubeconfig string, options client.Options) (client.Client, error) {
	return NewClient("", kubeconfig, "", options)
}

//url: The url of the server
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
//context: The context to connect to
func NewClient(url, kubeconfig, context string, options client.Options) (client.Client, error) {
	klog.V(5).Infof("Create kubeclient for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := config.LoadConfig(url, kubeconfig, context)
	if err != nil {
		return nil, err
	}

	client, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	return client, nil
}

//NewDefaultKubeClient returns a kubernetes.Interface for the current-context in kubeconfig
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
func NewDefaultKubeClient(kubeconfig string) (kubernetes.Interface, error) {
	return NewKubeClient("", kubeconfig, "")
}

//NewKubeClient returns a kubernetes.Interface based on the provided url, kubeconfig and context
//url: The url of the server
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
//context: The context to connect to
func NewKubeClient(url, kubeconfig, context string) (kubernetes.Interface, error) {
	klog.V(5).Infof("Create kubeclient for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := config.LoadConfig(url, kubeconfig, context)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

//NewDefaultKubeClientDynamic returns a dynamic.Interface for the current-context in kubeconfig
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
func NewDefaultKubeClientDynamic(kubeconfig string) (dynamic.Interface, error) {
	return NewKubeClientDynamic("", kubeconfig, "")
}

//NewKubeClientDynamic returns a dynamic.Interface based on the provided url, kubeconfig and context
//url: The url of the server
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
//context: The context to connect to
func NewKubeClientDynamic(url, kubeconfig, context string) (dynamic.Interface, error) {
	klog.V(5).Infof("Create kubeclient dynamic for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := config.LoadConfig(url, kubeconfig, context)
	if err != nil {
		return nil, err
	}

	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

//NewDefaultKubeClientAPIExtension returns a clientset.Interface for the current-context in kubeconfig
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
func NewDefaultKubeClientAPIExtension(kubeconfig string) (clientset.Interface, error) {
	return NewKubeClientAPIExtension("", kubeconfig, "")
}

//NewKubeClientAPIExtension returns a clientset.Interface based on the provided url, kubeconfig and context
//url: The url of the server
//kubeconfig: The path of the kubeconfig, see (../config/config.go#LoadConfig) for more information
//context: The context to connect to
func NewKubeClientAPIExtension(url, kubeconfig, context string) (clientset.Interface, error) {
	klog.V(5).Infof("Create kubeclient apiextension for url %s using kubeconfig path %s\n", url, kubeconfig)
	config, err := config.LoadConfig(url, kubeconfig, context)
	if err != nil {
		return nil, err
	}

	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

//HaveServerResources returns an error if all provided APIGroups are not installed
//client: the client to use
//expectedAPIGroups: The list of expected APIGroups
func HaveServerResources(client clientset.Interface, expectedAPIGroups []string) error {
	clientDiscovery := client.Discovery()
	for _, apiGroup := range expectedAPIGroups {
		klog.V(1).Infof("Check if %s exists", apiGroup)
		_, err := clientDiscovery.ServerResourcesForGroupVersion(apiGroup)
		if err != nil {
			klog.V(1).Infof("Error while retrieving server resource %s: %s", apiGroup, err.Error())
			return err
		}
	}
	return nil
}

//Deprecated:
// Use https://github.com/stolostron/library-go/pkg/apis/meta/v1/crd#HasCRDs
//HaveCRDs returns an error if all provided CRDs are not installed
//client: the client to use
//expectedCRDs: The list of expected CRDS to find
func HaveCRDs(client clientset.Interface, expectedCRDs []string) error {
	has, _, err := libgocrdv1.HasCRDs(client, expectedCRDs)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("Some CRDs are missing")
	}
	return nil
}

//Deprecated:
// Use https://github.com/stolostron/library-go/pkg/apis/meta/v1/deployment#HaveDeploymentsInNamespace
//HaveDeploymentsInNamespace returns an error if all provided deployment are not installed in the given namespace
//client: the client to use
//namespace: The namespace to search in
//expectedDeploymentNames: The deployment names to search
func HaveDeploymentsInNamespace(client kubernetes.Interface,
	namespace string,
	expectedDeploymentNames []string,
) error {
	has, _, err := libgodeploymentv1.HasDeploymentsInNamespace(client, namespace, expectedDeploymentNames)
	if err != nil {
		return err
	}
	if !has {
		return fmt.Errorf("Some deployments are missing or not ready")
	}
	return nil
}
