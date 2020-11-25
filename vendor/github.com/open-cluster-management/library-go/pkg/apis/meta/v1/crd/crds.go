package crd

import (
	"context"

	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

//HasCRDs returns false if one CRDs is missing and the list of missing CRDs.
//It returns an error if an error occured while retreiving a CRDs.
//client: the client to use
//expectedCRDs: The list of expected CRDS to find
func HasCRDs(client clientset.Interface, expectedCRDs []string) (has bool, missingCRDs []string, err error) {
	missingCRDs = make([]string, 0)
	has = true
	clientAPIExtensionV1beta1 := client.ApiextensionsV1beta1()
	for _, crd := range expectedCRDs {
		klog.V(1).Infof("Check if %s exists", crd)
		_, errGet := clientAPIExtensionV1beta1.CustomResourceDefinitions().Get(context.TODO(), crd, metav1.GetOptions{})
		if errGet != nil {
			if errors.IsNotFound(errGet) {
				missingCRDs = append(missingCRDs, crd)
				has = false
			} else {
				klog.V(1).Infof("Error while retrieving crd %s: %s", crd, errGet.Error())
				return false, missingCRDs, errGet
			}
		}
	}
	return has, missingCRDs, err
}
