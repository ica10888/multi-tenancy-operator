package multitenancycontroller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancydirector"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancywatcher"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sync"
)

var log = logf.Log.WithName("controller_multi_tenancy")

var Mutex sync.Mutex


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


type TenancyExample struct {
	Reconcile *ReconcileMultiTenancyController
	TenancyOperator TenancyOperator
	NamespacedChart NamespacedChart
	NamespacedController types.NamespacedName
	Namespaces []string
	Settings map[string]string
	StateSettings map[string]string
}


var localSpec = []v1alpha1.Tenancy{}

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

	// Choose tenancy-director constructor and tenancy-watcher constructor, like Plugins
	LoopSchedule(tenancydirector.ChartDirectorFor(),tenancywatcher.ReplicationControllerWatcherFor(manager.Manager.GetConfig(mgr.GetConfig())))

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

	defer func(){
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()

	// Fetch the multiTenancyController instance
	checkMultiTenancyController,err := checkMultiTenancyController(r.Client,reqLogger)
	if err != nil {
		reqLogger.Error(err,"Check Err")
		return reconcile.Result{}, err
	}

	if checkMultiTenancyController.InitCheck() {
		r.Client.Update(context.TODO(),checkMultiTenancyController)
		r.Client.Status().Update(context.TODO(),checkMultiTenancyController)
		reqLogger.Info("Init check failed, init multiTenancyController")
		return reconcile.Result{},nil
	}

	if equalTenancies(checkMultiTenancyController.Spec.Tenancies,localSpec) {
		return reconcile.Result{},nil
	}
	localSpec = checkMultiTenancyController.Spec.Tenancies

	teList, result, err := addTenancyExampleList(checkMultiTenancyController, err, r, reqLogger)
	if err != nil {
		return result, err
	}

	for _, example := range teList {
		TenancyQueue <- example
	}
	return reconcile.Result{}, nil
}

func addTenancyExampleList(checkMultiTenancyController *v1alpha1.Controller, err error, r *ReconcileMultiTenancyController, reqLogger logr.Logger) ([]TenancyExample, reconcile.Result, error) {
	defer func(){
		Mutex.Unlock()
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()
	Mutex.Lock()

	multiTenancyController := &v1alpha1.Controller{}

	controllerNamespacedName := types.NamespacedName{checkMultiTenancyController.Namespace, checkMultiTenancyController.Name}

	err = r.Client.Get(context.TODO(), controllerNamespacedName, multiTenancyController)
	if err != nil {
		return nil, reconcile.Result{}, err
	}

	reqLogger.Info("Reconciling multiTenancyController")

	ten := flatMapTenancies(multiTenancyController.Spec.Tenancies)

	staTen := flatMapUpdatedTenancies(multiTenancyController.Status.UpdatedTenancies)

	teList := []TenancyExample{}

	namespaces := []string{}
	for _, tenancy := range multiTenancyController.Spec.Tenancies {
		namespaces = append(namespaces, tenancy.Namespace)
	}

	for namespacedChart, _ := range staTen {
		sets := ten[namespacedChart]
		if sets == nil {
			delete := TenancyExample{
				Reconcile:            r,
				TenancyOperator:      DELETE,
				NamespacedChart:      namespacedChart,
				NamespacedController: controllerNamespacedName,
				Namespaces:           namespaces,
				Settings:             sets,
			}
			chartName := mergeReleaseChartName(namespacedChart.ChartName, namespacedChart.ReleaseName)
			multiTenancyController.Status.RemoveNamespacedChart(chartName, namespacedChart.Namespace)
			multiTenancyController.Status.UpdateNamespacedChartSettings(namespacedChart.ChartName, namespacedChart.Namespace, sets)
			teList = append(teList, delete)

		}
	}
	for namespacedChart, sets := range ten {
		staSets := staTen[namespacedChart]
		if staSets == nil {
			create := TenancyExample{
				Reconcile:            r,
				TenancyOperator:      CREATE,
				NamespacedChart:      namespacedChart,
				NamespacedController: controllerNamespacedName,
				Settings:             sets,
				StateSettings:        staSets,
			}
			chartName := mergeReleaseChartName(namespacedChart.ChartName, namespacedChart.ReleaseName)
			multiTenancyController.Status.AppendNamespacedChart(chartName, namespacedChart.Namespace)
			teList = append(teList, create)
		} else {
			if !equal(sets, staSets) {
				update := TenancyExample{
					Reconcile:            r,
					TenancyOperator:      UPDATE,
					NamespacedChart:      namespacedChart,
					NamespacedController: controllerNamespacedName,
					Namespaces:           namespaces,
					Settings:             sets,
				}
				chartName := mergeReleaseChartName(namespacedChart.ChartName, namespacedChart.ReleaseName)
				multiTenancyController.Status.UpdateNamespacedChartSettings(chartName, namespacedChart.Namespace, sets)
				teList = append(teList, update)
			}
		}
	}
	r.Client.Status().Update(context.TODO(), multiTenancyController)
	return teList, reconcile.Result{},err
}

func checkMultiTenancyController(c client.Client, reqLogger logr.Logger) (*v1alpha1.Controller,error){
	multiTenancyControllerList := &v1alpha1.ControllerList{}
	multiTenancyController := &v1alpha1.Controller{}

	listOpts := []client.ListOption{
		client.InNamespace(metav1.NamespaceAll),
	}
	err := c.List(context.TODO(),multiTenancyControllerList,listOpts...)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("multiTenancyController resource not found. Ignoring since object must be deleted")
			return nil, err
		}
		reqLogger.Error(err, "Failed to get multiTenancyController")
		return nil, err
	}

	//multiTenancyController can not exist more than one at same time
	if len(multiTenancyControllerList.Items) >= 2 {
		oldestController := &v1alpha1.Controller{}
		var unixNano int64 =math.MinInt64
		for _, item := range multiTenancyControllerList.Items {
			if item.ObjectMeta.CreationTimestamp.UnixNano() > unixNano {
				oldestController = &item
			} else {
				c.Delete(context.TODO(),&item)
			}
		}
		err := fmt.Errorf("Controller can not exist more than one at same time, Controller is in %s namespace",oldestController.Namespace)
		reqLogger.Error(err, "Failed to create multiTenancyController")
		return nil, err
	}

	multiTenancyController = &multiTenancyControllerList.Items[0]

	return multiTenancyController,nil
}
