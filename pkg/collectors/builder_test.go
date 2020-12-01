package collectors

import (
	"reflect"
	"testing"
	"time"

	managedclusterv1 "github.com/open-cluster-management/api/cluster/v1"
	ocinfrav1 "github.com/openshift/api/config/v1"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kube-state-metrics/pkg/collector"
	"k8s.io/kube-state-metrics/pkg/options"
	"k8s.io/kube-state-metrics/pkg/whiteblacklist"
)

var (
	ctx context.Context = context.TODO()
)

func TestBuilder_WithApiserver(t *testing.T) {
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        options.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		apiserver string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Builder
	}{
		{
			name: "apiserver",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				apiserver: "apiserver",
			},
			want: &Builder{
				apiserver:         "apiserver",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithApiserver(tt.args.apiserver); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithApiserver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithKubeConfig(t *testing.T) {
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        options.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		kubeconfig string
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
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				kubeconfig: "kubeconfig",
			},
			want: &Builder{
				apiserver:         "",
				kubeconfig:        "kubeconfig",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.WithKubeConfig(tt.args.kubeconfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Builder.WithKubeConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_WithEnabledCollectors(t *testing.T) {
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        options.NamespaceList
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
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				c: []string{"col1", "col2"},
			},
			want: &Builder{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
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
		namespaces        options.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		n options.NamespaceList
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
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				// whiteBlackList:    whiteBlackLister{func(s) { return false }, func(s) { return false }},
			},
			args: args{
				n: []string{"ns1"},
			},
			want: &Builder{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        []string{"ns1"},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
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
		namespaces        options.NamespaceList
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
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				whiteBlackList:    nil,
			},
			args: args{
				l: w,
			},
			want: &Builder{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"col1", "col2"},
				whiteBlackList:    w,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
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

func TestBuilder_buildClusterDeploymentCollectorWithClient(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.ClusterVersion{})

	version := &ocinfrav1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Spec: ocinfrav1.ClusterVersionSpec{
			ClusterID: "mycluster_id",
		},
	}

	client := fake.NewSimpleDynamicClient(s, version)

	w, _ := whiteblacklist.New(map[string]struct{}{}, map[string]struct{}{})
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        options.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		client dynamic.Interface
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   *collector.Collector
	}{
		{
			name: "buildClusterDeploymentCollector",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"clusterdeployements"},
				whiteBlackList:    w,
			},
			args: args{
				client: client,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.buildClusterDeploymentCollectorWithClient(tt.args.client); got == nil {
				t.Errorf("Builder.buildClusterDeploymentCollector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_buildManagedClusterCollectorWithClient(t *testing.T) {
	s := scheme.Scheme

	s.AddKnownTypes(managedclusterv1.SchemeGroupVersion, &managedclusterv1.ManagedCluster{})
	s.AddKnownTypes(ocinfrav1.SchemeGroupVersion, &ocinfrav1.ClusterVersion{})

	mcImported := &managedclusterv1.ManagedCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "imported-cluster",
			CreationTimestamp: metav1.Time{Time: time.Unix(1500000000, 0)},
			Labels: map[string]string{
				"cloud":  "aws",
				"vendor": "OpneShift",
			},
		},
		Status: managedclusterv1.ManagedClusterStatus{
			Version: managedclusterv1.ManagedClusterVersion{
				Kubernetes: "v1.16.2",
			},
		},
	}

	version := &ocinfrav1.ClusterVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name: "version",
		},
		Spec: ocinfrav1.ClusterVersionSpec{
			ClusterID: "mycluster_id",
		},
	}

	client := fake.NewSimpleDynamicClient(s, mcImported, version)

	w, _ := whiteblacklist.New(map[string]struct{}{}, map[string]struct{}{})
	type fields struct {
		apiserver         string
		kubeconfig        string
		namespaces        options.NamespaceList
		ctx               context.Context
		enabledCollectors []string
		whiteBlackList    whiteBlackLister
	}
	type args struct {
		client dynamic.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *collector.Collector
	}{
		{
			name: "buildManagedClusterCollectorWithClient",
			fields: fields{
				apiserver:         "",
				kubeconfig:        "",
				namespaces:        options.NamespaceList{},
				ctx:               ctx,
				enabledCollectors: []string{"managedclusterinfos"},
				whiteBlackList:    w,
			},
			args: args{
				client: client,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Builder{
				apiserver:         tt.fields.apiserver,
				kubeconfig:        tt.fields.kubeconfig,
				namespaces:        tt.fields.namespaces,
				ctx:               tt.fields.ctx,
				enabledCollectors: tt.fields.enabledCollectors,
				whiteBlackList:    tt.fields.whiteBlackList,
			}
			if got := b.buildManagedClusterInfoCollectorWithClient(tt.args.client); got == nil {
				t.Errorf("Builder.buildManagedClusterCollectorWithClient() = %v, want %v", got, tt.want)
			}
		})
	}
}
