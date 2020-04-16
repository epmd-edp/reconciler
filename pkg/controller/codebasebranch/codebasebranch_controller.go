package codebasebranch

import (
	"context"
	"github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebasebranch"
	cbs "github.com/epmd-edp/reconciler/v2/pkg/service/codebasebranch"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	errWrap "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_codebasebranch")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCodebaseBranch{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		cbService: cbs.CodebaseBranchService{
			DB: db.Instance,
		},
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("codebasebranch-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*v1alpha1.CodebaseBranch)
			newObject := e.ObjectNew.(*v1alpha1.CodebaseBranch)

			if oldObject.Status.Value != newObject.Status.Value ||
				oldObject.Status.Action != newObject.Status.Action {
				return true
			}

			if !reflect.DeepEqual(oldObject.Spec, newObject.Spec) {
				return true
			}

			if oldObject.Status.LastSuccessfulBuild != newObject.Status.LastSuccessfulBuild {
				return true
			}

			if oldObject.Status.Build != newObject.Status.Build {
				return true
			}

			if newObject.DeletionTimestamp != nil {
				return true
			}
			return false
		},
	}

	err = c.Watch(&source.Kind{Type: &v1alpha1.CodebaseBranch{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCodebaseBranch{}

const codebaseBranchReconcilerFinalizerName = "codebasebranch.reconciler.finalizer.name"

type ReconcileCodebaseBranch struct {
	client    client.Client
	scheme    *runtime.Scheme
	cbService cbs.CodebaseBranchService
}

func (r *ReconcileCodebaseBranch) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.Info("Reconciling CodebaseBranch")

	i := &v1alpha1.CodebaseBranch{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	edpN, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errWrap.Wrap(err, "couldn't get edp name")
	}

	if res, err := r.tryToDeleteCodebaseBranch(i, *edpN); err != nil || res != nil {
		return *res, err
	}

	app, err := codebasebranch.ConvertToCodebaseBranch(*i, *edpN)
	if err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errWrap.Wrap(err, "cannot convert to codebase branch dto")
	}
	if err := r.cbService.PutCodebaseBranch(*app); err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errWrap.Wrap(err, "couldn't insert codebase branch")
	}
	rl.Info("Reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r *ReconcileCodebaseBranch) tryToDeleteCodebaseBranch(cb *v1alpha1.CodebaseBranch, schema string) (*reconcile.Result, error) {
	if cb.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(cb.ObjectMeta.Finalizers, codebaseBranchReconcilerFinalizerName) {
			cb.ObjectMeta.Finalizers = append(cb.ObjectMeta.Finalizers, codebaseBranchReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), cb); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}

	if err := r.cbService.Delete(cb.Spec.CodebaseName, cb.Spec.BranchName, schema); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}

	cb.ObjectMeta.Finalizers = helper.RemoveString(cb.ObjectMeta.Finalizers, codebaseBranchReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), cb); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}
	return &reconcile.Result{}, nil
}
