package multitenancycontroller

import (
	"context"
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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

func (t TenancyOperator) ToString() string{
	res := ""
	switch t {
	case UPDATE:
		res = "update"
	case CREATE:
		res = "create"
	case DELETE:
		res = "delete"
	}
	return res
}

const (
	UPDATE TenancyOperator = "update"
	CREATE TenancyOperator = "create"
	DELETE TenancyOperator = "delete"
)

type NamespacedChart struct {
	Namespace string
	ChartName string
	ReleaseName string
}
type NamespacedController struct {
	Namespace string
	ControllerName string
}

type TenancyExample struct {
	Reconcile *ReconcileMultiTenancyController
	TenancyOperator TenancyOperator
	NamespacedChart NamespacedChart
	NamespacedController NamespacedController
	Settings map[string]string
	StateSettings map[string]string
}

var TenancyQueue = make(chan TenancyExample)

var fmtAuthErr = `
Unauthorized Error in %s, please execute cmd to solve:
kubectl create namespace %s
echo '{"apiVersion":"v1","kind":"ServiceAccount","metadata":{"name":"multi-tenancy-operator"}}' |  kubectl create -n %s -f -
kubectl get clusterrolebinding multi-tenancy-operator -n %s -o json | jq '.subjects[ .subjects | length] +=  {"kind":"ServiceAccount","name": "multi-tenancy-operator","namespace": "%s"}'  | kubectl apply -n %s -f -`

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
	return &ReconcileMultiTenancyController{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
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
	Client client.Client
	Scheme *runtime.Scheme
}


// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMultiTenancyController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling multiTenancyController")

	defer func(){
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()

	// Fetch the multiTenancyController instance
	multiTenancyController,err := checkMultiTenancyController(r.Client,reqLogger)
	if err != nil {
		reqLogger.Error(err,"Check Err")
		return reconcile.Result{}, err
	}

	if multiTenancyController.InitCheck() {
		r.Client.Update(context.TODO(),multiTenancyController)
		r.Client.Status().Update(context.TODO(),multiTenancyController)
		reqLogger.Info("Init check failed, init multiTenancyController")
		return reconcile.Result{},nil
	}

	//TODO check spec if not change, do nothing

	ten := flatMapTenancies(multiTenancyController.Spec.Tenancies)

	staTen := flatMapUpdatedTenancies(multiTenancyController.Status.UpdatedTenancies)

	teList := []TenancyExample{}
	for namespacedChart, _ := range staTen {
		sets := ten[namespacedChart]
		if sets == nil {
			delete := TenancyExample {
				Reconcile: r,
				TenancyOperator: DELETE,
				NamespacedChart: namespacedChart,
				NamespacedController:NamespacedController{request.Namespace,request.Name},
				Settings: sets,
			}
			chartName := mergeReleaseChartName(namespacedChart.ChartName,namespacedChart.ReleaseName)
			multiTenancyController.Status.RemoveNamespacedChart(chartName,namespacedChart.Namespace)
			multiTenancyController.Status.UpdateNamespacedChartSettings(namespacedChart.ChartName,namespacedChart.Namespace,sets)
			teList = append(teList, delete)

		}
	}
	for namespacedChart, sets := range ten {
		staSets := staTen[namespacedChart]
		if staSets == nil {
			create := TenancyExample {
				Reconcile: r,
				TenancyOperator: CREATE,
				NamespacedChart: namespacedChart,
				NamespacedController:NamespacedController{request.Namespace,request.Name},
				Settings: sets,
				StateSettings: staSets,
			}
			chartName := mergeReleaseChartName(namespacedChart.ChartName,namespacedChart.ReleaseName)
			multiTenancyController.Status.AppendNamespacedChart(chartName,namespacedChart.Namespace)
			teList = append(teList, create)
		} else {
			if ! equal(sets,staSets) {
				update := TenancyExample {
					Reconcile: r,
					TenancyOperator: UPDATE,
					NamespacedChart: namespacedChart,
					NamespacedController:NamespacedController{request.Namespace,request.Name},
					Settings: sets,
				}
				chartName := mergeReleaseChartName(namespacedChart.ChartName,namespacedChart.ReleaseName)
				multiTenancyController.Status.UpdateNamespacedChartSettings(chartName,namespacedChart.Namespace,sets)
				teList = append(teList, update)
			}
		}
	}
	r.Client.Status().Update(context.TODO(),multiTenancyController)
	for _, example := range teList {
		TenancyQueue <- example
	}
	return reconcile.Result{}, nil
}

//FIFO work queue
func LoopSchedule(tenancyDirector TenancyDirector){
	go func(){
		for {
			tenancyExample := <- TenancyQueue
			switch tenancyExample.TenancyOperator {
			case UPDATE:
				ScheduleProcessor(tenancyDirector.UpdateSingleTenancyByConfigure,&tenancyExample)
			case CREATE:
				ScheduleProcessor(tenancyDirector.CreateSingleTenancyByConfigure,&tenancyExample)
			case DELETE:
				ScheduleProcessor(tenancyDirector.DeleteSingleTenancyByConfigure,&tenancyExample)
			}
		}
	}()

}

func ScheduleProcessor(operatorSingleTenancyByConfigure func (*TenancyExample) ([]KubeObject,error),t *TenancyExample){
	reqLogger := log.WithValues("Namespace", t.NamespacedController.Namespace, "Name", t.NamespacedController.ControllerName)
	defer func(){
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s",err),"recover Err")

			multiTenancyController,err := checkMultiTenancyController(t.Reconcile.Client,reqLogger)
			if err != nil {
				reqLogger.Error(err,"Write ErrorMessage Check Err")
			}
			chartName := mergeReleaseChartName(t.NamespacedChart.ChartName,t.NamespacedChart.ReleaseName)
			multiTenancyController.Status.UpdateNamespacedChartErrorMessage(chartName,t.NamespacedChart.Namespace,fmt.Errorf("%s",err))
			t.Reconcile.Client.Update(context.TODO(),multiTenancyController)
		}
	}()
	reqLogger.Info(fmt.Sprintf("Start to %s",t.TenancyOperator.ToString()))
	_ ,err := operatorSingleTenancyByConfigure(t)

	multiTenancyController,err := checkMultiTenancyController(t.Reconcile.Client,reqLogger)
	if err != nil {
		reqLogger.Error(err,"Write ErrorMessage Check Err")
	}
	if apierrs.IsUnauthorized(err){
		err = fmt.Errorf(fmtAuthErr,t.NamespacedChart.Namespace,t.NamespacedChart.Namespace,t.NamespacedChart.Namespace,multiTenancyController.Namespace,t.NamespacedChart.Namespace,multiTenancyController.Namespace)
	}
	chartName := mergeReleaseChartName(t.NamespacedChart.ChartName,t.NamespacedChart.ReleaseName)
	multiTenancyController.Status.UpdateNamespacedChartErrorMessage(chartName,t.NamespacedChart.Namespace,err)
	t.Reconcile.Client.Update(context.TODO(),multiTenancyController)
}
