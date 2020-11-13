package perfdatasourcejenkins

import (
	"context"
	"github.com/epmd-edp/perf-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/util/cluster"
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
	"time"
)

var log = logf.Log.WithName("controller_perf_data_source_jenkins")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePerfDataSourceJenkins{
		client: mgr.GetClient(),
		dsService: perfdatasource.PerfDataSourceService{
			DB: db.Instance,
		},
	}
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	c, err := controller.New("perf-data-source-jenkins-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectNew.(*v1alpha1.PerfDataSourceJenkins).DeletionTimestamp != nil
		},
	}

	if err = c.Watch(&source.Kind{Type: &v1alpha1.PerfDataSourceJenkins{}}, &handler.EnqueueRequestForObject{}, p); err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcilePerfDataSourceJenkins{}

const (
	codebaseKind = "Codebase"

	jenkinsDataSourceReconcilerFinalizerName = "jenkins.data.source.reconciler.finalizer.name"
)

type ReconcilePerfDataSourceJenkins struct {
	client    client.Client
	dsService perfdatasource.PerfDataSourceService
}

func (r *ReconcilePerfDataSourceJenkins) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.Info("Reconciling PerfDataSourceJenkins")

	i := &v1alpha1.PerfDataSourceJenkins{}
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

	result, err := r.tryToDeleteCodebasePerfDataSourceJenkins(i, *schema)
	if err != nil || result != nil {
		return *result, err
	}

	rl.Info("PerfDataSourceJenkins reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r *ReconcilePerfDataSourceJenkins) tryToDeleteCodebasePerfDataSourceJenkins(ds *v1alpha1.PerfDataSourceJenkins, schema string) (*reconcile.Result, error) {
	if ds.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(ds.ObjectMeta.Finalizers, jenkinsDataSourceReconcilerFinalizerName) {
			ds.ObjectMeta.Finalizers = append(ds.ObjectMeta.Finalizers, jenkinsDataSourceReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), ds); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}

	ow := cluster.GetOwnerReference(codebaseKind, ds.GetOwnerReferences())
	if ow == nil {
		log.Info("jenkins data source doesn't contain Codebase owner reference", "data source", ds.Name)
		return &reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if err := r.dsService.RemoveCodebaseDataSource(ow.Name, ds.Spec.Type, schema); err != nil {
		return &reconcile.Result{}, err
	}

	ds.ObjectMeta.Finalizers = helper.RemoveString(ds.ObjectMeta.Finalizers, jenkinsDataSourceReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), ds); err != nil {
		return &reconcile.Result{}, err
	}
	return &reconcile.Result{}, nil
}
