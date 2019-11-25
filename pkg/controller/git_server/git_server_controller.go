package git_server

import (
	"context"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/helper"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model/gitserver"
	"github.com/epmd-edp/reconciler/v2/pkg/service/git"
	"github.com/epmd-edp/reconciler/v2/pkg/service/infrastructure"
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

var log = logf.Log.WithName("controller_git_server")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GitServer Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitServer{
		Client: mgr.GetClient(),
		GitServerService: git.GitServerService{
			DB: db.Instance,
		},
		InfrastructureDbService: infrastructure.InfrastructureDbService{
			DB: db.Instance,
		},
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("git-server-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GitServer
	err = c.Watch(&source.Kind{Type: &edpv1alpha1Codebase.GitServer{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitServer{}

// ReconcileGitServer reconciles a GitServer object
type ReconcileGitServer struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client                  client.Client
	GitServerService        git.GitServerService
	InfrastructureDbService infrastructure.InfrastructureDbService
}

// Reconcile reads that state of the cluster for a ReconcileGitServer object and makes changes based on the state read
// and what is in the ReconcileGitServer.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGitServer) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling GitServer")

	// Fetch the GitServer instance
	instance := &edpv1alpha1Codebase.GitServer{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
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
	log.WithValues("GitServer", instance)
	edpN, err := helper.GetEDPName(r.Client, instance.Namespace)
	if err != nil {
		return reconcile.Result{}, err
	}
	gitServer, err := gitserver.ConvertToGitServer(*instance, *edpN)
	if err != nil {
		return reconcile.Result{}, err
	}

	exists, err := r.InfrastructureDbService.DoesSchemaExist(gitServer.Tenant)
	if err != nil {
		return reconcile.Result{}, errWrap.Wrap(err, "an error has occurred while checking schema in BD")
	}
	reqLogger.Info("Check schema: ", "schema", gitServer.Tenant, "exists", exists)

	if exists {
		err := r.GitServerService.PutGitServer(*gitServer)
		if err != nil {
			return reconcile.Result{}, err
		}

	}

	return reconcile.Result{}, nil
}
