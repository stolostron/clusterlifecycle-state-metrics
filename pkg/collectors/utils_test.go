// Copyright (c) 2021 Red Hat, Inc.

package collectors

import (
	"testing"

	ocinfrav1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/scheme"
)

func Test_getHubClusterID(t *testing.T) {
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
	type args struct {
		c dynamic.Interface
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Get cluster id",
			args: args{
				c: client,
			},
			want: "mycluster_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHubClusterID(tt.args.c); got != tt.want {
				t.Errorf("getHubClusterID() = %v, want %v", got, tt.want)
			}
		})
	}
}
