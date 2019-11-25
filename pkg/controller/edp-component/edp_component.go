package edp_component

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	ec "github.com/epmd-edp/reconciler/v2/pkg/service/edp-component"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	edpComponentV1Api "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("edp_component_controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new JobProvisioning Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &EDPComponent{
		client:              mgr.GetClient(),
		EDPComponentService: ec.EDPComponentService{DB: db.Instance},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("edp-component-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*edpComponentV1Api.EDPComponent).Spec
			new := e.ObjectNew.(*edpComponentV1Api.EDPComponent).Spec

			if reflect.DeepEqual(old, new) {
				return false
			}
			return true
		},
	}

	// Watch for changes to primary resource Jenkins
	err = c.Watch(&source.Kind{Type: &edpComponentV1Api.EDPComponent{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &EDPComponent{}

// EDPComponent reconciles a EDPComponent object
type EDPComponent struct {
	client              client.Client
	EDPComponentService ec.EDPComponentService
}

// Reconcile reads that state of the cluster for a EDPComponent object and makes changes based on the state read
// and what is in the Jenkins.Spec
//
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *EDPComponent) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling EDPComponent CR")

	i := &edpComponentV1Api.EDPComponent{}
	err := r.client.Get(context.TODO(), request.NamespacedName, i)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	c, err := model.ConvertToEDPComponent(*i)
	if err != nil {
		return reconcile.Result{}, err
	}
	log.Info("start reconciling for component", "type", c.Type, "url", c.Url)
	edpN, err := helper.GetEDPName(r.client, i.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.EDPComponentService.PutEDPComponent(*c, *edpN)
	if err != nil {
		return reconcile.Result{RequeueAfter: time.Second * 120}, err
	}

	return reconcile.Result{}, nil
}
