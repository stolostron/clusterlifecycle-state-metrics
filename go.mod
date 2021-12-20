module github.com/open-cluster-management/clusterlifecycle-state-metrics

go 1.17

replace (
	github.com/kubevirt/terraform-provider-kubevirt => github.com/nirarg/terraform-provider-kubevirt v0.0.0-20201222125919-101cee051ed3
	github.com/metal3-io/baremetal-operator => github.com/openshift/baremetal-operator v0.0.0-20200715132148-0f91f62a41fe
	github.com/metal3-io/cluster-api-provider-baremetal => github.com/openshift/cluster-api-provider-baremetal v0.0.0-20190821174549-a2a477909c1d
	github.com/terraform-providers/terraform-provider-ignition/v2 => github.com/community-terraform-providers/terraform-provider-ignition/v2 v2.1.0
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	kubevirt.io/client-go => kubevirt.io/client-go v0.29.0
	sigs.k8s.io/cluster-api-provider-aws => github.com/openshift/cluster-api-provider-aws v0.2.1-0.20201022175424-d30c7a274820
	sigs.k8s.io/cluster-api-provider-azure => github.com/openshift/cluster-api-provider-azure v0.1.0-alpha.3.0.20201016155852-4090a6970205
	sigs.k8s.io/cluster-api-provider-openstack => github.com/openshift/cluster-api-provider-openstack v0.0.0-20201116051540-155384b859c5
)

require (
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/open-cluster-management/api v0.0.0-20210527013639-a6845f2ebcb1
	github.com/open-cluster-management/applier v0.0.0-20210422205113-6c10f923726b
	github.com/open-cluster-management/library-go v0.0.0-20211008065140-521095fbdcb7
	github.com/open-cluster-management/multicloud-operators-foundation v0.0.0-20210415053135-434684dc0274
	github.com/openshift/api v3.9.1-0.20191111211345-a27ff30ebf09+incompatible
	github.com/openshift/build-machinery-go v0.0.0-20210423112049-9415d7ebd33e
	github.com/operator-framework/operator-sdk v0.17.0
	github.com/prometheus/client_golang v1.10.0
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	k8s.io/apimachinery v0.21.1
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kube-state-metrics v1.9.8
	k8s.io/kubernetes v1.13.0
	sigs.k8s.io/controller-runtime v0.8.3
)

require (
	cloud.google.com/go v0.74.0 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.18 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.13 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.1.1 // indirect
	github.com/Masterminds/sprig/v3 v3.2.2 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/form3tech-oss/jwt-go v3.2.2+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/googleapis/gnostic v0.5.1 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.20.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 // indirect
	golang.org/x/oauth2 v0.0.0-20210210192628-66670185b0cd // indirect
	golang.org/x/sys v0.0.0-20210925032602-92d5a993a665 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.21.1 // indirect
	k8s.io/apiextensions-apiserver v0.21.1 // indirect
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7 // indirect
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
