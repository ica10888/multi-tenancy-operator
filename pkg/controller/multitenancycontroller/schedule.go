package multitenancycontroller

import (
	"context"
	"fmt"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

//FIFO work queue
func LoopSchedule(tenancyDirector TenancyDirector,tenancyWatcher TenancyWatcher){
	go func(){
		for {
			tenancyExample := <- TenancyQueue
			switch tenancyExample.TenancyOperator {
			case UPDATE:
				objs:= ScheduleProcessor(tenancyDirector.UpdateSingleTenancyByConfigure,&tenancyExample)
				tenancyWatcher.UpdateTenancyPodStatusAndReplicationControllerStatus(objs,&tenancyExample)
			case CREATE:
				objs:= ScheduleProcessor(tenancyDirector.CreateSingleTenancyByConfigure,&tenancyExample)
				tenancyWatcher.UpdateTenancyWatcher(objs,&tenancyExample)
			case DELETE:
				ScheduleProcessor(tenancyDirector.DeleteSingleTenancyByConfigure,&tenancyExample)
			}
		}
	}()

}

func ScheduleProcessor(operatorSingleTenancyByConfigure func (*TenancyExample) ([]KubeObject,error),t *TenancyExample) (objs []KubeObject) {
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
	objs ,err := operatorSingleTenancyByConfigure(t)

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

	return
}
