package codebase

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/service"
	"github.com/epmd-edp/reconciler/v2/pkg/service/codebaseperfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfserver"
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

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCodebase{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		service: service.CodebaseService{
			DB: db.Instance,
			DataSourceService: perfdatasource.PerfDataSourceService{
				DB: db.Instance,
			},
			PerfService: perfserver.PerfServerService{
				DB: db.Instance,
			},
			CodebaseDsService: codebaseperfdatasource.CodebasePerfDataSourceService{
				DB: db.Instance,
			},
		},
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
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

	err = c.Watch(&source.Kind{Type: &edpv1alpha1Codebase.Codebase{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var (
	_   reconcile.Reconciler = &ReconcileCodebase{}
	log                      = logf.Log.WithName("controller_codebase")
)

const codebaseReconcilerFinalizerName = "codebase.reconciler.finalizer.name"

type ReconcileCodebase struct {
	client  client.Client
	scheme  *runtime.Scheme
	service service.CodebaseService
}

func (r *ReconcileCodebase) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.Info("Reconciling Codebase")

	i := &edpv1alpha1Codebase.Codebase{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	rl.Info("Codebase has been retrieved", "codebase", i)

	edpN, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		rl.Error(err, "cannot get edp name")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	result, err := r.tryToDeleteCodebase(i, *edpN)
	if err != nil || result != nil {
		return *result, err
	}

	c, err := codebase.Convert(*i, *edpN)
	if err != nil {
		rl.Error(err, "cannot convert codebase to dto")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if err = r.service.PutCodebase(*c); err != nil {
		rl.Error(err, "cannot put codebase", "name", c.Name)
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	rl.Info("Reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r *ReconcileCodebase) tryToDeleteCodebase(i *edpv1alpha1Codebase.Codebase, schema string) (*reconcile.Result, error) {
	if i.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(i.ObjectMeta.Finalizers, codebaseReconcilerFinalizerName) {
			i.ObjectMeta.Finalizers = append(i.ObjectMeta.Finalizers, codebaseReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), i); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}
	if err := r.service.Delete(i.Name, schema); err != nil {
		return &reconcile.Result{}, err
	}

	i.ObjectMeta.Finalizers = helper.RemoveString(i.ObjectMeta.Finalizers, codebaseReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), i); err != nil {
		return &reconcile.Result{}, err
	}
	return &reconcile.Result{}, nil
}
