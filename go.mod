module github.com/open-cluster-management/clusterlifecycle-state-metrics

go 1.16

replace k8s.io/client-go => k8s.io/client-go v0.19.0

replace k8s.io/kube-state-metrics => k8s.io/kube-state-metrics v0.0.0-20190129120824-7bfed92869b6

// Determined by go.mod in github.com/openshift/hive
replace (
	github.com/Azure/go-autorest => github.com/tombuildsstuff/go-autorest v14.0.1-0.20200416184303-d4e299a3c04a+incompatible
	github.com/Azure/go-autorest/autorest => github.com/tombuildsstuff/go-autorest/autorest v0.10.1-0.20200416184303-d4e299a3c04a
	github.com/Azure/go-autorest/autorest/azure/auth => github.com/tombuildsstuff/go-autorest/autorest/azure/auth v0.4.3-0.20200416184303-d4e299a3c04a
	github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20200715132148-0f91f62a41fe
	github.com/metal3-io/cluster-api-provider-baremetal => github.com/openshift/cluster-api-provider-baremetal v0.0.0-20190821174549-a2a477909c1d
	github.com/openshift/library-go => github.com/openshift/library-go v0.0.0-20200918101923-1e4c94603efe
	github.com/terraform-providers/terraform-provider-aws => github.com/openshift/terraform-provider-aws v1.60.1-0.20200630224953-76d1fb4e5699
	github.com/terraform-providers/terraform-provider-azurerm => github.com/openshift/terraform-provider-azurerm v1.40.1-0.20200707062554-97ea089cc12a
	github.com/terraform-providers/terraform-provider-ignition/v2 => github.com/community-terraform-providers/terraform-provider-ignition/v2 v2.1.0
	sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20200506073438-9d49428ff837
	sigs.k8s.io/cluster-api-provider-azure => github.com/openshift/cluster-api-provider-azure v0.1.0-alpha.3.0.20200120114645-8a9592f1f87b
	sigs.k8s.io/cluster-api-provider-openstack => github.com/openshift/cluster-api-provider-openstack v0.0.0-20200526112135-319a35b2e38e
)

require (
	github.com/onsi/ginkgo v1.14.0
	github.com/onsi/gomega v1.10.1
	github.com/open-cluster-management/api v0.0.0-20201007180356-41d07eee4294
	github.com/open-cluster-management/library-go v0.0.0-20200828173847-299c21e6c3fc
	github.com/open-cluster-management/multicloud-operators-foundation v0.0.0-20201112041030-60ef45157161
	github.com/openshift/api v3.9.1-0.20191112184635-86def77f6f90+incompatible
	github.com/openshift/build-machinery-go v0.0.0-20200819073603-48aa266c95f7
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/prometheus/client_golang v1.7.1
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.3.0
	k8s.io/kube-state-metrics v1.7.2
	sigs.k8s.io/controller-runtime v0.6.3
)
