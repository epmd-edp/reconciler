package jenkins_job

import (
	"context"
	jenv1alpha1 "github.com/epmd-edp/jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/jenkins_job/service"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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

var (
	_   reconcile.Reconciler = &ReconcileJenkinsJob{}
	log                      = logf.Log.WithName("jenkins-job-controller")
)

// Add creates a new Stage Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	c := mgr.GetClient()
	return &ReconcileJenkinsJob{
		client: c,
		scheme: mgr.GetScheme(),
		JenkinsJobService: service.JenkinsJobService{
			DB:     db.Instance,
			Client: c,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("jenkins-job-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObject := e.ObjectOld.(*jenv1alpha1.JenkinsJob)
			newObject := e.ObjectNew.(*jenv1alpha1.JenkinsJob)
			if oldObject.Status.Action != newObject.Status.Action ||
				oldObject.Status.Value != newObject.Status.Value {
				return true
			}
			return false
		},
	}

	// Watch for changes to primary resource Stage
	err = c.Watch(&source.Kind{Type: &jenv1alpha1.JenkinsJob{}}, &handler.EnqueueRequestForObject{}, pred)
	if err != nil {
		return err
	}

	return nil
}

type ReconcileJenkinsJob struct {
	client            client.Client
	scheme            *runtime.Scheme
	JenkinsJobService service.JenkinsJobService
}

func (r *ReconcileJenkinsJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	rl := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	rl.V(2).Info("Reconciling JenkinsJob")
	i := &jenv1alpha1.JenkinsJob{}
	if err := r.client.Get(context.TODO(), request.NamespacedName, i); err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	if err := r.JenkinsJobService.UpdateActionLog(i); err != nil {
		return reconcile.Result{RequeueAfter: 5 * time.Second}, err
	}

	rl.V(2).Info("Reconciling JenkinsJob has been finished successfully")
	return reconcile.Result{}, nil
}
