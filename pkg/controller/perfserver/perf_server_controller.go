package perfserver

import (
	"context"
	"github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	perfServerModel "github.com/epmd-edp/reconciler/v2/pkg/model/perfserver"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfserver"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_perf_server")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePerfServer{
		client:      mgr.GetClient(),
		perfService: perfserver.PerfServerService{DB: db.Instance},
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("perf-server-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*v1alpha1.PerfServer)
			newObject := e.ObjectNew.(*v1alpha1.PerfServer)
			if oldObject.Spec != newObject.Spec {
				return true
			}
			if oldObject.Status.Available != newObject.Status.Available {
				return true
			}
			return false
		},
	}

	if err = c.Watch(&source.Kind{Type: &v1alpha1.PerfServer{}}, &handler.EnqueueRequestForObject{}, p); err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcilePerfServer{}

type ReconcilePerfServer struct {
	client      client.Client
	perfService perfserver.PerfServerService
}

func (r *ReconcilePerfServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.Info("Reconciling PerfServer")

	i := &v1alpha1.PerfServer{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	schema, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.perfService.PutPerfServer(perfServerModel.ConvertPerfServerToDto(*i), *schema); err != nil {
		return reconcile.Result{}, err
	}

	rl.Info("PerfServer reconciling has been finished successfully")
	return reconcile.Result{}, nil
}
