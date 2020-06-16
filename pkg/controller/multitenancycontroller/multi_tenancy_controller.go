package multitenancycontroller

import (
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_multi_tenancy")

type TenancyOperator string

const (
	UPDATE TenancyOperator = "update"
	CREATE TenancyOperator = "create"
	DELETE TenancyOperator = "delete"
)

type NamespacedChart struct {
	Namespace string
	ChartName string
}
type NamespacedController struct {
	Namespace string
	ControllerName string
}

type TenancyExample struct {
	TenancyOperator TenancyOperator
	NamespacedChart NamespacedChart
	NamespacedController NamespacedController
	Settings []v1alpha1.Setting
}

var TenancyQueue chan TenancyExample

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMultiTenancyController{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("multi-tenancy-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource
	err = c.Watch(&source.Kind{Type: &v1alpha1.Controller{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMultiTenancyController{}

// ReconcileMultiTenancyController reconciles a MultiTenancyController object
type ReconcileMultiTenancyController struct {
	// TODO: Clarify the split client
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}


// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMultiTenancyController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling multiTenancyController")
	
	// Fetch the multiTenancyController instance
	multiTenancyController,err := checkMultiTenancyController(r.client,reqLogger ,request)
	if err != nil {
		return reconcile.Result{}, err
	}

	ten := flatMapTenancies(multiTenancyController.Spec.Tenancies)

	staTen := flatMapTenancies(multiTenancyController.Status.Updated)

	for namespacedChart, _ := range staTen {
		sets := ten[namespacedChart]
		if sets == nil {
			delete := TenancyExample {
				TenancyOperator: DELETE,
				NamespacedChart: namespacedChart,
				NamespacedController:NamespacedController{request.Namespace,request.Name},
				Settings: sets,
			}
			TenancyQueue <- delete
		}
	}
	for namespacedChart, sets := range ten {
		staSets := staTen[namespacedChart]
		if staSets == nil {
			create := TenancyExample {
				TenancyOperator: CREATE,
				NamespacedChart: namespacedChart,
				NamespacedController:NamespacedController{request.Namespace,request.Name},
				Settings: sets,
			}
			TenancyQueue <- create
		} else {
			if ! equal(sets,staSets) {
				update := TenancyExample {
					TenancyOperator: UPDATE,
					NamespacedChart: namespacedChart,
					NamespacedController:NamespacedController{request.Namespace,request.Name},
					Settings: sets,
				}
				TenancyQueue <- update
			}
		}
	}

	return reconcile.Result{}, nil
}

//FIFO work queue
func LoopSchedule(tenancyManager TenancyManager){
	go func(){
		for {
			tenancyExample := <- TenancyQueue
			switch tenancyExample.TenancyOperator {
			case UPDATE:
				ScheduleProcessor(tenancyManager.UpdateSingleTenancyByConfigure,&tenancyExample)
			case CREATE:
				ScheduleProcessor(tenancyManager.CreateSingleTenancyByConfigure,&tenancyExample)
			case DELETE:
				ScheduleProcessor(tenancyManager.DeleteSingleTenancyByConfigure,&tenancyExample)
			}
		}
	}()

}

func ScheduleProcessor(operatorSingleTenancyByConfigure func (*TenancyExample) error,t *TenancyExample){
	reqLogger := log.WithValues("Namespace", t.NamespacedController.Namespace, "Name", t.NamespacedController.ControllerName)
	defer func(){
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s",err),"recover Err: ")
		}
	}()
	err := operatorSingleTenancyByConfigure(t)
	if err != nil {
		reqLogger.Error(err,"Tenancy %s failed, reason: " ,t.TenancyOperator)
	}
}
