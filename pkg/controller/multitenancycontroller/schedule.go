package multitenancycontroller

import (
	"context"
	"fmt"
)

//FIFO work queue
func LoopSchedule(tenancyScheduler TenancyScheduler, tenancyWatcher TenancyWatcher) {
	go func() {
		for {
			tenancyExample := <-TenancyQueue
			switch tenancyExample.TenancyOperator {
			case UPDATE:
				recoverScheduleProcessor(tenancyScheduler.UpdateSingleTenancyByConfigure, &tenancyExample)
			case CREATE:
				objs := recoverScheduleProcessor(tenancyScheduler.CreateSingleTenancyByConfigure, &tenancyExample)
				recoverNamespaceWatcherProcessor(tenancyWatcher.CreateTenancyNamespacesIfNeed, &tenancyExample)
				recoverRCAndPodWatcherProcessor(tenancyWatcher.CreateTenancyPodStatusAndReplicationControllerStatus, objs, &tenancyExample)
			case DELETE:
				objs := recoverScheduleProcessor(tenancyScheduler.DeleteSingleTenancyByConfigure, &tenancyExample)
				recoverNamespaceWatcherProcessor(tenancyWatcher.DeleteTenancyNamespacesIfNeed, &tenancyExample)
				recoverRCAndPodWatcherProcessor(tenancyWatcher.DeleteTenancyPodStatusAndReplicationControllerStatus, objs, &tenancyExample)
			case INIT:
				recoverNamespaceWatcherProcessor(tenancyWatcher.InitTenancyWatcher, &tenancyExample)
			}
		}
	}()

}

func recoverNamespaceWatcherProcessor(tenancyNamespacesFunc func(*TenancyExample), t *TenancyExample) {
	reqLogger := log.WithValues("Namespace", t.NamespacedController.Namespace, "Name", t.NamespacedController.Name)
	defer func() {
		Mutex.Unlock()
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s", err), "recover Err")
		}
	}()
	Mutex.Lock()
	tenancyNamespacesFunc(t)
}

func recoverRCAndPodWatcherProcessor(tenancyRCAndPodFunc func([]KubeObject, *TenancyExample), objs []KubeObject, t *TenancyExample) {
	reqLogger := log.WithValues("Namespace", t.NamespacedController.Namespace, "Name", t.NamespacedController.Name)
	defer func() {
		Mutex.Unlock()
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s", err), "recover Err")
		}
	}()
	Mutex.Lock()
	tenancyRCAndPodFunc(objs, t)
}

func recoverScheduleProcessor(operatorSingleTenancyByConfigure func(*TenancyExample) ([]KubeObject, error), t *TenancyExample) (objs []KubeObject) {
	reqLogger := log.WithValues("Namespace", t.NamespacedController.Namespace, "Name", t.NamespacedController.Name)
	defer func() {
		if err := recover(); err != nil {
			reqLogger.Error(fmt.Errorf("%s", err), "recover Err")

			mTC, err := CheckMultiTenancyController(t.Reconcile.Client, reqLogger)
			if err != nil {
				reqLogger.Error(err, "Write ErrorMessage Check Err")
			}
			chartName := mergeReleaseChartName(t.NamespacedChart.ChartName, t.NamespacedChart.ReleaseName)
			mTC.Status.UpdateNamespacedChartErrorMessage(chartName, t.NamespacedChart.Namespace, fmt.Errorf("%s", err))
			t.Reconcile.Client.Update(context.TODO(), mTC)
		}
	}()
	reqLogger.Info(fmt.Sprintf("Start to %s", t.TenancyOperator.ToString()))
	objs, err := operatorSingleTenancyByConfigure(t)

	mTC, err := CheckMultiTenancyController(t.Reconcile.Client, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Write ErrorMessage Check Err")
	}

	chartName := mergeReleaseChartName(t.NamespacedChart.ChartName, t.NamespacedChart.ReleaseName)
	mTC.Status.UpdateNamespacedChartErrorMessage(chartName, t.NamespacedChart.Namespace, err)
	t.Reconcile.Client.Update(context.TODO(), mTC)

	return
}
