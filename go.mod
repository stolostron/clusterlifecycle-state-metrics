module github.com/open-cluster-management/ocm-state-metrics

go 1.14

require (
	github.com/kr/pretty v0.2.1 // indirect
	github.com/open-cluster-management/api v0.0.0-20201007180356-41d07eee4294
	github.com/openshift/api v3.9.1-0.20191112184635-86def77f6f90+incompatible // indirect
	github.com/openshift/build-machinery-go v0.0.0-20200819073603-48aa266c95f7
	github.com/openshift/hive v0.0.0-20200318152403-0c1ea8babb4e
	github.com/prometheus/client_golang v1.7.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	k8s.io/klog/v2 v2.3.0
	k8s.io/kube-state-metrics v0.0.0-20190129120824-7bfed92869b6
	sigs.k8s.io/controller-runtime v0.6.3 // indirect

)
