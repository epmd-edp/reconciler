package stage

import (
	"context"
	edpV1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	stage2 "github.com/epmd-edp/reconciler/v2/pkg/service/stage"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	_   reconcile.Reconciler = &ReconcileStage{}
	log                      = logf.Log.WithName("controller_stage")
)

// Add creates a new Stage Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	clientSet, err := platform.CreateOpenshiftClients()
	if err != nil {
		panic(err)
	}

	return &ReconcileStage{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		service: stage2.StageService{
			DB:        db.Instance,
			ClientSet: *clientSet,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("stage-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*edpV1alpha1.Stage)
			newObject := e.ObjectNew.(*edpV1alpha1.Stage)

			if oldObject.Status.Value != newObject.Status.Value {
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

	// Watch for changes to primary resource Stage
	err = c.Watch(&source.Kind{Type: &edpV1alpha1.Stage{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

const stageReconcilerFinalizerName = "stage.reconciler.finalizer.name"

type ReconcileStage struct {
	client  client.Client
	scheme  *runtime.Scheme
	service stage2.StageService
}

func (r *ReconcileStage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.V(2).Info("Reconciling Stage")
	i := &edpV1alpha1.Stage{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	edpN, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errors.Wrap(err, "cannot get edp name")
	}

	if res, err := r.tryToDeleteCDStage(i, *edpN); err != nil || res != nil {
		return *res, err
	}

	st, err := stage.ConvertToStage(*i, *edpN)
	if err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errors.Wrap(err, "couldn't convert to stage dto")
	}

	if err = r.service.PutStage(*st); err != nil {
		return reconcile.Result{RequeueAfter: 2 * time.Second}, errors.Wrap(err, "couldn't put stage")
	}
	rl.V(2).Info("Reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r ReconcileStage) tryToDeleteCDStage(i *edpV1alpha1.Stage, schema string) (*reconcile.Result, error) {
	if i.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(i.ObjectMeta.Finalizers, stageReconcilerFinalizerName) {
			i.ObjectMeta.Finalizers = append(i.ObjectMeta.Finalizers, stageReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), i); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}

	if err := r.service.DeleteCDStage(i.Spec.CdPipeline, i.Spec.Name, schema); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}

	i.ObjectMeta.Finalizers = helper.RemoveString(i.ObjectMeta.Finalizers, stageReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), i); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}
	return &reconcile.Result{}, nil
}
