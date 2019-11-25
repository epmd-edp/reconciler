package template

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/thirdpartyservice"
	"github.com/epmd-edp/reconciler/v2/pkg/service"
	"github.com/openshift/api/template/v1"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Template Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTemplate{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		service: service.ThirdPartyService{
			DB: db.Instance,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("template-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Template
	err = c.Watch(&source.Kind{Type: &v1.Template{}}, &handler.EnqueueRequestForObject{}, predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTemplate{}

// ReconcileTemplate reconciles a Template object
type ReconcileTemplate struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	service service.ThirdPartyService
}

// Reconcile reads that state of the cluster for a Template object and makes changes based on the state read
// and what is in the Template.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileTemplate) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	log.Println("Reconciling Template")

	// Fetch the Template instance
	instance := &v1.Template{}
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

	log.Printf("Template: %v %v %v %v %v", instance.Name, instance.Namespace, instance.Labels, instance.Annotations,
		instance.APIVersion)

	if strings.Contains(instance.Annotations["tags"], "edp") {
		edpN, err := helper.GetEDPName(r.client, instance.Namespace)
		if err != nil {
			return reconcile.Result{}, err
		}
		s, err := thirdpartyservice.ConvertToService(*instance, *edpN)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = r.service.PutService(*s)
		if err != nil {
			log.Printf("Couldn't save s %v to DB", s.Name)
			return reconcile.Result{RequeueAfter: time.Second * 120}, nil
		}

		log.Printf("Reconciling service template %v/%v has been finished", request.Namespace, request.Name)
		return reconcile.Result{Requeue: false}, nil
	}

	log.Printf("Template %v/%v doesn't contain EDP tag. Skipped. Reconciling service template has been finished",
		request.Namespace, request.Name)
	return reconcile.Result{Requeue: false}, nil
}
