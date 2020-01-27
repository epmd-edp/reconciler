package codebase

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/service"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
)

var log = logf.Log.WithName("controller_codebase")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Codebase Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCodebase{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		beService: service.BEService{
			DB: db.Instance,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("codebase-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*edpv1alpha1Codebase.Codebase)
			newObject := e.ObjectNew.(*edpv1alpha1Codebase.Codebase)

			if oldObject.Status.Value != newObject.Status.Value ||
				oldObject.Status.Action != newObject.Status.Action {
				return true
			}

			if !reflect.DeepEqual(oldObject.Spec, newObject.Spec) {
				return true
			}

			if newObject.DeletionTimestamp != nil {
				return true
			}
			return false
		},
	}

	// Watch for changes to primary resource Codebase
	err = c.Watch(&source.Kind{Type: &edpv1alpha1Codebase.Codebase{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCodebase{}

const CodebaseReconcilerFinalizerName = "codebase.reconciler.finalizer.name"

// ReconcileCodebase reconciles a Codebase object
type ReconcileCodebase struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client    client.Client
	scheme    *runtime.Scheme
	beService service.BEService
}

func (r *ReconcileCodebase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Codebase")

	i := &edpv1alpha1Codebase.Codebase{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	reqLogger.Info("Codebase has been retrieved", "codebase", i)

	edpN, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		reqLogger.Error(err, "cannot get edp name")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	result, err := r.tryToDeleteCodebase(i, *edpN)
	if err != nil || result != nil {
		return *result, err
	}

	c, err := codebase.Convert(*i, *edpN)
	if err != nil {
		reqLogger.Error(err, "cannot convert codebase to dto")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	err = r.beService.PutBE(*c)
	if err != nil {
		reqLogger.Error(err, "cannot put codebase branch")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	reqLogger.Info("Reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r *ReconcileCodebase) tryToDeleteCodebase(i *edpv1alpha1Codebase.Codebase, schema string) (*reconcile.Result, error) {
	if i.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(i.ObjectMeta.Finalizers, CodebaseReconcilerFinalizerName) {
			i.ObjectMeta.Finalizers = append(i.ObjectMeta.Finalizers, CodebaseReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), i); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}
	if err := r.beService.Delete(i.Name, schema); err != nil {
		return &reconcile.Result{}, err
	}

	i.ObjectMeta.Finalizers = helper.RemoveString(i.ObjectMeta.Finalizers, CodebaseReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), i); err != nil {
		return &reconcile.Result{}, err
	}
	return &reconcile.Result{}, nil
}
