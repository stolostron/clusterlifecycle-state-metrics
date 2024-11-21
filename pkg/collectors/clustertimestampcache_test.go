// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package collectors

import (
	"reflect"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	operatorv1 "open-cluster-management.io/api/operator/v1"
	workv1 "open-cluster-management.io/api/work/v1"
)

func Test_ClusterTimestampCache(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2021-09-01T00:01:01Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	t2, err := time.Parse(time.RFC3339, "2021-09-01T00:01:02Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	work1 := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster1-hosted-klusterlet",
			Namespace: "local-cluster",
			Labels: map[string]string{
				HostedClusterLabel: "cluster1",
			},
		},
		Status: workv1.ManifestWorkStatus{
			ResourceStatus: workv1.ManifestResourceStatus{
				Manifests: []workv1.ManifestCondition{
					{
						ResourceMeta: workv1.ManifestResourceMeta{
							Group:    operatorv1.GroupName,
							Kind:     "Klusterlet",
							Resource: "klusterlets",
							Name:     "klusterlet-cluster1",
						},
						StatusFeedbacks: workv1.StatusFeedbackResult{
							Values: []workv1.FeedbackValue{
								{
									Name: "ReadyToApply-status",
									Value: workv1.FieldValue{
										String: &[]string{"True"}[0],
									},
								},
								{
									Name: "ReadyToApply-lastTransitionTime",
									Value: workv1.FieldValue{
										String: &[]string{"2021-09-01T00:01:02Z"}[0],
									},
								},
								{
									Name: "ReadyToApply-message",
									Value: workv1.FieldValue{
										String: &[]string{"Klusterlet is ready to apply, the external managed kubeconfig secret was created at: 2021-09-01T00:01:01Z"}[0],
									},
								},
							},
						},
					},
				},
			},
		},
	}

	work1Modified := work1.DeepCopy()
	work1Modified.Status.ResourceStatus.Manifests[0].StatusFeedbacks.Values[1].Value.String = &[]string{"2021-09-01T00:01:01Z"}[0]
	work1ModifiedNull := work1.DeepCopy()
	work1ModifiedNull.Status.ResourceStatus.Manifests[0].StatusFeedbacks.Values[0].Value.String = &[]string{"False"}[0]

	work2 := &workv1.ManifestWork{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster2-hosted-klusterlet",
			Namespace: "local-cluster",
			Labels: map[string]string{
				HostedClusterLabel: "cluster2",
			},
		},
		Status: workv1.ManifestWorkStatus{
			ResourceStatus: workv1.ManifestResourceStatus{
				Manifests: []workv1.ManifestCondition{
					{
						ResourceMeta: workv1.ManifestResourceMeta{
							Group:    operatorv1.GroupName,
							Kind:     "Klusterlet",
							Resource: "klusterlets",
							Name:     "klusterlet-cluster2",
						},
						StatusFeedbacks: workv1.StatusFeedbackResult{
							Values: []workv1.FeedbackValue{
								{
									Name: "ReadyToApply-status",
									Value: workv1.FieldValue{
										String: &[]string{"True"}[0],
									},
								},
								{
									Name: "ReadyToApply-lastTransitionTime",
									Value: workv1.FieldValue{
										String: &[]string{"2021-09-01T00:01:02Z"}[0],
									},
								},
								{
									Name: "ReadyToApply-message",
									Value: workv1.FieldValue{
										String: &[]string{"Klusterlet is ready to apply, the external managed kubeconfig secret was created at: 2021-09-01T00:01:02Z"}[0],
									},
								},
							},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name                     string
		clusterName              string
		existing                 []interface{}
		toAdd                    []interface{}
		toUpdate                 []interface{}
		toDelete                 []interface{}
		want                     map[string]float64
		numberOfTimestampChanged int
	}{
		{
			name:        "empty",
			clusterName: "cluster1",
			want:        nil,
		},
		{
			name:        "existing",
			clusterName: "cluster1",
			existing:    []interface{}{work1},
			want: map[string]float64{
				StatusManagedClusterKubeconfigProvided: float64(t1.Unix()),
				StatusStartToApplyKlusterletResources:  float64(t2.Unix()),
			},
		},
		{
			name:        "add",
			clusterName: "cluster1",
			toAdd:       []interface{}{work1},
			want: map[string]float64{
				StatusManagedClusterKubeconfigProvided: float64(t1.Unix()),
				StatusStartToApplyKlusterletResources:  float64(t2.Unix()),
			},
			numberOfTimestampChanged: 1,
		},
		{
			name:        "update cluster1",
			clusterName: "cluster1",
			existing:    []interface{}{work1},
			toUpdate:    []interface{}{work1Modified},
			want: map[string]float64{
				StatusManagedClusterKubeconfigProvided: float64(t1.Unix()),
				StatusStartToApplyKlusterletResources:  float64(t1.Unix()),
			},
			numberOfTimestampChanged: 1,
		},
		{
			name:                     "update cluster1 to null",
			clusterName:              "cluster1",
			existing:                 []interface{}{work1},
			toUpdate:                 []interface{}{work1ModifiedNull},
			want:                     nil,
			numberOfTimestampChanged: 1,
		},
		{
			name:        "update cluster2",
			clusterName: "cluster2",
			toAdd:       []interface{}{work1},
			toUpdate:    []interface{}{work2},
			want: map[string]float64{
				StatusManagedClusterKubeconfigProvided: float64(t2.Unix()),
				StatusStartToApplyKlusterletResources:  float64(t2.Unix()),
			},
			numberOfTimestampChanged: 2,
		},
		{
			name:        "delete",
			clusterName: "cluster1",
			existing:    []interface{}{work1},
			toDelete:    []interface{}{work1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			numberOfTimestampChanged := 0
			cache := newClusterTimestampCache()
			cache.AddOnTimestampChangeFunc(func(clusterName string) error {
				numberOfTimestampChanged += 1
				return nil
			})

			err := cache.Replace(tt.existing, "")
			if err != nil {
				t.Errorf("caught unexpected err: %c", err)
			}

			for _, obj := range tt.toAdd {
				cache.Add(obj)
			}

			for _, obj := range tt.toUpdate {
				cache.Update(obj)
			}

			for _, obj := range tt.toDelete {
				cache.Delete(obj)
			}

			if actual := cache.GetClusterTimestamps(tt.clusterName); !reflect.DeepEqual(actual, tt.want) {
				t.Errorf("want\n%v\nbut got\n%v", tt.want, actual)
			}

			if numberOfTimestampChanged != tt.numberOfTimestampChanged {
				t.Errorf("want numberOfTimestampChanged %d\nbut got%d",
					tt.numberOfTimestampChanged, numberOfTimestampChanged)
			}
		})
	}
}
