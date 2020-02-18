package cdpipeline

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/cdpipeline"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	"github.com/epmd-edp/reconciler/v2/pkg/service/cd-pipeline"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	edpv1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"

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

var log = logf.Log.WithName("controller_cdpipeline")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CDPipeline Controller and adds it to the Manager. The Manager will set fields on the Controller
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
	cdpService := cd_pipeline.CdPipelineService{
		DB:        db.Instance,
		ClientSet: *clientSet,
	}
	return &ReconcileCDPipeline{client: mgr.GetClient(), scheme: mgr.GetScheme(), cdpService: cdpService}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cdpipeline-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*edpv1alpha1.CDPipeline)
			newObject := e.ObjectNew.(*edpv1alpha1.CDPipeline)

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

	// Watch for changes to primary resource CDPipeline
	err = c.Watch(&source.Kind{Type: &edpv1alpha1.CDPipeline{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCDPipeline{}

const cdPipelineReconcilerFinalizerName = "cdpipeline.reconciler.finalizer.name"

// ReconcileCDPipeline reconciles a CDPipeline object
type ReconcileCDPipeline struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	cdpService cd_pipeline.CdPipelineService
}

// Reconcile reads that state of the cluster for a CDPipeline object and makes changes based on the state read
// and what is in the CDPipeline.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCDPipeline) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CDPipeline")

	// Fetch the CDPipeline instance
	instance := &edpv1alpha1.CDPipeline{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	reqLogger.Info("CD pipeline has been retrieved", "cd pipeline", instance)

	edpN, err := helper.GetEDPName(r.client, instance.Namespace)
	if err != nil {
		reqLogger.Error(err, "cannot get edp name")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if res, err := r.tryToDeleteCDPipeline(instance, *edpN); err != nil || res != nil {
		return *res, err
	}

	cdp, err := cdpipeline.ConvertToCDPipeline(*instance, *edpN)
	if err != nil {
		reqLogger.Error(err, "cannot convert to cd pipeline dto")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}
	err = r.cdpService.PutCDPipeline(*cdp)
	if err != nil {
		reqLogger.Error(err, "cannot put cd pipeline")
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}

	reqLogger.Info("Reconciling has been finished successfully")
	return reconcile.Result{}, nil
}

func (r *ReconcileCDPipeline) tryToDeleteCDPipeline(p *edpv1alpha1.CDPipeline, schema string) (*reconcile.Result, error) {
	if p.GetDeletionTimestamp().IsZero() {
		if !helper.ContainsString(p.ObjectMeta.Finalizers, cdPipelineReconcilerFinalizerName) {
			p.ObjectMeta.Finalizers = append(p.ObjectMeta.Finalizers, cdPipelineReconcilerFinalizerName)
			if err := r.client.Update(context.TODO(), p); err != nil {
				return &reconcile.Result{}, err
			}
		}
		return nil, nil
	}

	if err := r.cdpService.DeleteCDPipeline(p.Name, schema); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}

	p.ObjectMeta.Finalizers = helper.RemoveString(p.ObjectMeta.Finalizers, cdPipelineReconcilerFinalizerName)
	if err := r.client.Update(context.TODO(), p); err != nil {
		return &reconcile.Result{RequeueAfter: 2 * time.Second}, err
	}
	return &reconcile.Result{}, nil
}
