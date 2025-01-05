package controllers

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	clusterv1 "open-cluster-management.io/api/cluster/v1"
	workv1 "open-cluster-management.io/api/work/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/common"
)

var (
	testscheme = scheme.Scheme
)

func init() {
	testscheme.AddKnownTypes(clusterv1.SchemeGroupVersion, &workv1.ManifestWork{})
}

func TestReconcile(t *testing.T) {

	before, err := time.Parse(time.RFC3339, "2021-09-01T00:01:01Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	current, err := time.Parse(time.RFC3339, "2021-09-01T00:01:02Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	after, err := time.Parse(time.RFC3339, "2021-09-01T00:01:03Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	lastTransitionTime, err := time.Parse(time.RFC3339, "2021-09-01T00:01:04Z")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	cases := []struct {
		name         string
		works        []client.Object
		request      reconcile.Request
		validateFunc func(t *testing.T, runtimeClient client.Client)
	}{
		{
			name: "manifest work is created before the controller starts",
			works: []client.Object{
				&workv1.ManifestWork{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-work1",
						Namespace: "test",
						Labels: map[string]string{
							common.LabelHostedCluster: "hosting1",
						},
						CreationTimestamp: metav1.NewTime(before),
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-work1",
					Namespace: "test",
				},
			},
			validateFunc: func(t *testing.T, runtimeClient client.Client) {
				mw := &workv1.ManifestWork{}
				if err := runtimeClient.Get(context.TODO(),
					types.NamespacedName{Name: "test-work1", Namespace: "test"}, mw); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if len(mw.Annotations) > 0 {
					t.Errorf("unexpected mw %s/%s annotations: %v", mw.Namespace, mw.Name, mw.Annotations)
				}
			},
		},
		{
			name: "manifest work is created after the controller starts, not applied",
			works: []client.Object{
				&workv1.ManifestWork{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-work1",
						Namespace: "test",
						Labels: map[string]string{
							common.LabelHostedCluster: "hosting1",
						},
						CreationTimestamp: metav1.NewTime(after),
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-work1",
					Namespace: "test",
				},
			},
			validateFunc: func(t *testing.T, runtimeClient client.Client) {
				mw := &workv1.ManifestWork{}
				if err := runtimeClient.Get(context.TODO(),
					types.NamespacedName{Name: "test-work1", Namespace: "test"}, mw); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if len(mw.Annotations) > 0 {
					t.Errorf("unexpected mw %s/%s annotations: %v", mw.Namespace, mw.Name, mw.Annotations)
				}
			},
		},
		{
			name: "manifest work is created after the controller starts, applied",
			works: []client.Object{
				&workv1.ManifestWork{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-work1",
						Namespace: "test",
						Labels: map[string]string{
							common.LabelHostedCluster: "hosting1",
						},
						CreationTimestamp: metav1.NewTime(after),
					},
					Status: workv1.ManifestWorkStatus{
						Conditions: []metav1.Condition{
							{
								Type:               workv1.WorkApplied,
								Status:             metav1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(lastTransitionTime),
							},
						},
					},
				},
			},
			request: reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "test-work1",
					Namespace: "test",
				},
			},
			validateFunc: func(t *testing.T, runtimeClient client.Client) {
				mw := &workv1.ManifestWork{}
				if err := runtimeClient.Get(context.TODO(),
					types.NamespacedName{Name: "test-work1", Namespace: "test"}, mw); err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				timestamp, ok := mw.Annotations[common.AnnotationAppliedTime]
				if !ok {
					t.Errorf("annotations %s not found", common.AnnotationAppliedTime)
				}
				if timestamp != `{"appliedTime":"2021-09-01T08:01:04+08:00"}` {
					t.Errorf("unexpected timestamp annotation value: %s", timestamp)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			ctx := context.TODO()

			runtimeClient := fake.NewClientBuilder().WithScheme(testscheme).
				WithObjects(c.works...).WithStatusSubresource(c.works...).Build()
			r := &ManifestworkReconciler{
				Client:    runtimeClient,
				startTime: current,
				// Scheme: testscheme,
			}

			_, err := r.Reconcile(ctx, c.request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			c.validateFunc(t, r.Client)
		})
	}
}
