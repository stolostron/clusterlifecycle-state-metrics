// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	ocinfrav1 "github.com/openshift/api/config/v1"
	ocpclient "github.com/openshift/client-go/config/clientset/versioned"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	koptions "k8s.io/kube-state-metrics/pkg/options"
	"k8s.io/kube-state-metrics/pkg/whiteblacklist"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	ocpclientfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	"k8s.io/client-go/kubernetes/fake"
	kubeclientfake "k8s.io/client-go/kubernetes/fake"
)

var (
	ctx context.Context = context.TODO()
)

func TestBuilder_WithKubeClient(t *testing.T) {
	fakeclient := fake.NewSimpleClientset()
	type fields struct {
		namespaces        koptions.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		kubeclient kubernetes.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{
			name: "kubeconfig",
			fields: fields{
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				kubeclient: fakeclient,
			},
			want: &Builder{
				kubeclient:        fakeclient,
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithKubeclient(tt.args.kubeclient); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithKubeclient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithEnabledCollectors(t *testing.T) {
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        koptions.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		c []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{
			name: "collectors",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				c: []string{"col1", "col2"},
			},
			want: &Builder{
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithEnabledCollectors(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithEnabledCollectors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithNamespaces(t *testing.T) {
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        koptions.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		n koptions.NamespaceList
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{
			name: "namespace",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				n: []string{"ns1"},
			},
			want: &Builder{
				namespaces:        []string{"ns1"},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithNamespaces(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithWhiteBlackList(t *testing.T) {
	w, _ := whiteblacklist.New(map[string]struct{}{}, map[string]struct{}{})
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        koptions.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		l whiteBlackLister
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{
			name: "whiteBlackList",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				whiteBlackList:    nil,
			},
			args: args{
				l: w,
			},
			want: &Builder{
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				whiteBlackList:    w,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithWhiteBlackList(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithWhiteBlackList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	const (
		clusterCollectorHeaders = `# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
# HELP acm_managed_cluster_labels Managed cluster labels
# TYPE acm_managed_cluster_labels gauge
# HELP acm_managed_cluster_status_condition Managed cluster status condition
# TYPE acm_managed_cluster_status_condition gauge
# HELP acm_managed_cluster_worker_cores The number of worker CPU cores of ACM managed clusters
# TYPE acm_managed_cluster_worker_cores gauge
# HELP acm_managed_cluster_count Managed cluster count
# TYPE acm_managed_cluster_count gauge
`
		addOnCollectorHeaders = `# HELP acm_managed_cluster_addon_status_condition Managed cluster add-on status condition
# TYPE acm_managed_cluster_addon_status_condition gauge
`
		workCollectorHeaders = `# HELP acm_manifestwork_status_condition ManifestWork status condition
# TYPE acm_manifestwork_status_condition gauge
# HELP acm_manifestwork_count ManifestWork count
# TYPE acm_manifestwork_count gauge
`
	)

	envTest := setupEnvTest(t)
	_, err := envtest.InstallCRDs(envTest.Config, envtest.CRDInstallOptions{
		Paths: []string{"../../test/unit/resources/crds"},
	})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = envTest.ControlPlane.KubeCtl().Run("create", "ns", "imported-cluster")
	if err != nil {
		t.Fatal(err)
	}

	kubeconfigFile, err := ioutil.TempFile("", "ut-kubeconfig-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(kubeconfigFile.Name())

	err = writeKubeconfigFile(envTest.Config, kubeconfigFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	ocpClient, _ := ocpclient.NewForConfig(envTest.Config)

	version := &ocinfrav1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Spec: ocinfrav1.ClusterVersionSpec{
			ClusterID: "mycluster_id",
		},
	}

	_, err = ocpClient.ConfigV1().
		ClusterVersions().
		Create(context.TODO(), version, metav1.CreateOptions{})
	if err != nil {
		t.Fatal(err)
	}

	w, _ := whiteblacklist.New(map[string]struct{}{}, map[string]struct{}{})
	type fields struct {
		kubeconfig        string
		namespaces        koptions.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			name: "no collector enabled",
			fields: fields{
				kubeconfig:     kubeconfigFile.Name(),
				namespaces:     koptions.NamespaceList{},
				ctx:            ctx,
				whiteBlackList: w,
			},
		},
		{
			name: "managedclusters enabled",
			fields: fields{
				kubeconfig:        kubeconfigFile.Name(),
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"managedclusters"},
				whiteBlackList:    w,
			},
			want: []string{clusterCollectorHeaders},
		},
		{
			name: "all collectors enabled",
			fields: fields{
				kubeconfig:        kubeconfigFile.Name(),
				namespaces:        koptions.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"managedclusters", "managedclusteraddons", "manifestworks"},
				whiteBlackList:    w,
			},
			want: []string{clusterCollectorHeaders, addOnCollectorHeaders, workCollectorHeaders},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder(tt.fields.ctx)
			// b.kubeconfig = tt.fields.kubeconfig
			b.namespaces = tt.fields.namespaces
			b.enabledCollectors = tt.fields.enabledCollectors
			b.whiteBlackList = tt.fields.whiteBlackList

			stores := b.Build()
			if len(stores) != len(tt.want) {
				t.Errorf(
					"number of MetricsStores = %v, want %v",
					len(stores),
					len(tt.want),
				)
			}
			for index, got := range stores {
				buf := new(bytes.Buffer)
				got.WriteAll(buf)
				if buf.String() != tt.want[index] {
					t.Errorf("Expected headers \n%s\ngot\n%s", tt.want[index], buf.String())
				}
			}
		})
	}
}

func writeKubeconfigFile(restConfig *rest.Config, kubeconfigFileName string) error {
	kubeconfig := clientcmdapi.Config{
		Clusters: map[string]*clientcmdapi.Cluster{"default-cluster": {
			Server:                   restConfig.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: restConfig.CAData,
		}},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"default-auth": {
			ClientCertificateData: restConfig.CertData,
			ClientKeyData:         restConfig.KeyData,
		}},
		Contexts: map[string]*clientcmdapi.Context{"default-context": {
			Cluster:   "default-cluster",
			AuthInfo:  "default-auth",
			Namespace: "default",
		}},
		CurrentContext: "default-context",
	}

	return clientcmd.WriteToFile(kubeconfig, kubeconfigFileName)
}

func Test_getHubClusterID(t *testing.T) {

	tests := []struct {
		name           string
		clusterVersion *ocinfrav1.ClusterVersion
		ns             *corev1.Namespace
		want           string
	}{
		{
			name: "Get cluster id",
			clusterVersion: &ocinfrav1.ClusterVersion{
				ObjectMeta: metav1.ObjectMeta{
					Name: "version",
				},
				Spec: ocinfrav1.ClusterVersionSpec{
					ClusterID: "mycluster_id",
				},
			},
			ns:   &corev1.Namespace{},
			want: "mycluster_id",
		},
		{
			name:           "Get cluster id",
			clusterVersion: &ocinfrav1.ClusterVersion{},
			ns: &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{
					Name: "kube-system",
					UID:  "kube-system-uid",
				},
			},
			want: "kube-system-uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeOcpClient := ocpclientfake.NewSimpleClientset(tt.clusterVersion)
			fakeKubeClient := kubeclientfake.NewSimpleClientset(tt.ns)
			if got := getHubClusterID(fakeOcpClient, fakeKubeClient); got != tt.want {
				t.Errorf("getHubClusterID() = %v, want %v", got, tt.want)
			}
		})
	}
}
