// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"testing"

	ocinfrav1 "github.com/openshift/api/config/v1"
	ocpclient "github.com/openshift/client-go/config/clientset/versioned"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

func Test_getHubClusterID(t *testing.T) {
	envTest := setupEnvTest(t)
	_, err := envtest.InstallCRDs(envTest.Config, envtest.CRDInstallOptions{
		Paths: []string{"../../test/unit/resources/crds"},
	})
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

	type args struct {
		c *ocpclient.Clientset
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Get cluster id",
			args: args{
				c: ocpClient,
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
