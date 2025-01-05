package controllers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/stolostron/clusterlifecycle-state-metrics/pkg/common"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	workv1 "open-cluster-management.io/api/work/v1"
)

// ManifestworkReconciler reconciles a Manifestwork object
type ManifestworkReconciler struct {
	client.Client

	startTime time.Time
}

func (r *ManifestworkReconciler) setStartTime(logger logr.Logger) {
	if !r.startTime.IsZero() {
		return
	}
	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	r.startTime = time.Now()
	logger.Info("Set the controller start time", "startTime", r.startTime)
}

// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *ManifestworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)
	r.setStartTime(logger)

	omw := &workv1.ManifestWork{}
	err := r.Client.Get(ctx, req.NamespacedName, omw)
	if err != nil {
		return ctrl.Result{}, err
	}

	mw := omw.DeepCopy()
	logger.Info("test manifestwork", "manifestwor,", mw)

	if mw.CreationTimestamp.Time.Before(r.startTime) {
		return ctrl.Result{}, nil
	}

	if r.recorded(mw) {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, r.record(ctx, mw)
}

func (c *ManifestworkReconciler) recorded(mw *workv1.ManifestWork) bool {
	value, ok := mw.Annotations[common.AnnotationAppliedTime]
	if !ok || len(value) == 0 {
		return false
	}

	appliedTimestamp := &common.AppliedTimestamp{}
	err := json.Unmarshal([]byte(value), appliedTimestamp)
	return err == nil
}

func (c *ManifestworkReconciler) record(ctx context.Context, mw *workv1.ManifestWork) error {
	cond := meta.FindStatusCondition(mw.Status.Conditions, workv1.WorkApplied)
	if cond == nil || cond.Status != metav1.ConditionTrue {
		return nil
	}

	appliedTimestamp := &common.AppliedTimestamp{
		AppliedTime: cond.LastTransitionTime.Time,
	}

	appliedTimeValue, err := json.Marshal(appliedTimestamp)
	if err != nil {
		return err
	}
	return c.updateAnnotations(ctx, mw, common.AnnotationAppliedTime, string(appliedTimeValue))
}

func (r *ManifestworkReconciler) updateAnnotations(ctx context.Context,
	mw *workv1.ManifestWork, key, value string) error {
	patch := client.MergeFrom(mw.DeepCopy()) // Ensure we only update the annotations

	// Ensure annotations are updated
	if mw.Annotations == nil {
		mw.Annotations = make(map[string]string)
	}
	mw.Annotations[key] = value

	data, err := patch.Data(mw)
	klog.InfoS("debug patch", "data", data, "error", err, "type", patch.Type())
	if err := r.Client.Patch(ctx, mw, patch); err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManifestworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workv1.ManifestWork{}).
		WithEventFilter(predicate.NewPredicateFuncs(
			func(object client.Object) bool {
				mw, ok := object.(*workv1.ManifestWork)
				if !ok {
					return false
				}

				report, _ := common.FilterTimestampManifestwork(mw)
				return report
			},
		)).
		Named("Manifestwork").
		Complete(r)
}
