package controllers

import (
	"context"
	"encoding/json"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	workv1 "open-cluster-management.io/api/work/v1"

	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/common"
)

// manifestworkReconciler reconciles a Manifestwork object
type manifestworkReconciler struct {
	client.Client

	// StartTime is the start time of the controller
	StartTime time.Time
}

func NewManifestworkReconciler(c client.Client, startTime time.Time) *manifestworkReconciler {
	return &manifestworkReconciler{
		Client:    c,
		StartTime: startTime,
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *manifestworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	omw := &workv1.ManifestWork{}
	err := r.Client.Get(ctx, req.NamespacedName, omw)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	mw := omw.DeepCopy()
	logger.V(4).Info("start to reconcile manifestwork")

	timestamp := common.GetObservedTimestamp(mw)
	if timestamp != nil {
		return ctrl.Result{}, nil
	}

	err = r.record(ctx, mw)
	if err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("observed timestamp recorded")
	return ctrl.Result{}, nil
}

func (c *manifestworkReconciler) record(ctx context.Context, mw *workv1.ManifestWork) error {
	cond := meta.FindStatusCondition(mw.Status.Conditions, workv1.WorkApplied)
	if cond == nil || cond.Status != metav1.ConditionTrue {
		return nil
	}

	observedTimestamp := &common.ObservedTimestamp{
		AppliedTime: cond.LastTransitionTime.Time,
	}

	observedTimestampValue, err := json.Marshal(observedTimestamp)
	if err != nil {
		return err
	}
	return c.updateAnnotations(ctx, mw, common.AnnotationObservedTimestamp, string(observedTimestampValue))
}

func (r *manifestworkReconciler) updateAnnotations(ctx context.Context,
	mw *workv1.ManifestWork, key, value string) error {
	patch := client.MergeFrom(mw.DeepCopy()) // Ensure we only update the annotations

	// Ensure annotations are updated
	if mw.Annotations == nil {
		mw.Annotations = make(map[string]string)
	}
	mw.Annotations[key] = value

	if err := r.Client.Patch(ctx, mw, patch); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *manifestworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workv1.ManifestWork{}).
		WithEventFilter(predicate.NewPredicateFuncs(
			func(object client.Object) bool {
				mw, ok := object.(*workv1.ManifestWork)
				if !ok {
					return false
				}

				// Only handle the manifestworks created after the controller starts
				if mw.CreationTimestamp.Time.Before(r.StartTime) {
					return false
				}

				report, _ := common.FilterTimestampManifestwork(mw)
				return report
			},
		)).
		Named("ManifestWork").
		Complete(r)
}
