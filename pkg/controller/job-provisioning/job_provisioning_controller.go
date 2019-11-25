package job_provisioning

import (
	"context"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	jp "github.com/epmd-edp/reconciler/v2/pkg/service/job-provisioning"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sort"
	"time"

	jenkinsV2Api "github.com/epmd-edp/jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	errWrap "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("job_provisioning_controller")

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
	return &ReconcileJobProvision{
		client: mgr.GetClient(),
		JobProvisionService: jp.JobProvisionService{
			DB: db.Instance,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("job-provisioning-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			old := e.ObjectOld.(*jenkinsV2Api.Jenkins).Status.JobProvisions
			new := e.ObjectNew.(*jenkinsV2Api.Jenkins).Status.JobProvisions

			sort.Slice(old, func(i, j int) bool {
				return old[i].Name < old[j].Name
			})
			sort.Slice(new, func(i, j int) bool {
				return new[i].Name < new[j].Name
			})

			if reflect.DeepEqual(old, new) {
				return false
			}

			return true
		},
	}

	// Watch for changes to primary resource Jenkins
	err = c.Watch(&source.Kind{Type: &jenkinsV2Api.Jenkins{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileJobProvision{}

// ReconcileJobProvisioning reconciles a JenkinsCR object
type ReconcileJobProvision struct {
	client              client.Client
	JobProvisionService jp.JobProvisionService
}

// Reconcile reads that state of the cluster for a Jenkins object and makes changes based on the state read
// and what is in the Jenkins.Spec
//
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJobProvision) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Jenkins CR to handle Job Provisios")

	instance := &jenkinsV2Api.Jenkins{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	jp := instance.Status.JobProvisions
	edpN, err := helper.GetEDPName(r.client, instance.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.JobProvisionService.PutJobProvisions(jp, *edpN)
	if err != nil {
		return reconcile.Result{RequeueAfter: time.Second * 120},
			errWrap.Wrapf(err, "an error has occurred while adding {%v} job provisions into DB", jp)
	}

	return reconcile.Result{}, nil
}
