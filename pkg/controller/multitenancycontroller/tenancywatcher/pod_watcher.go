package tenancywatcher

import (
	"context"
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func podWatcher(clientSet *kubernetes.Clientset,c client.Client,nsCtx *context.Context,namespace string) (err error) {
	watcher, err := clientSet.CoreV1().Pods(namespace).Watch(v1.ListOptions{})
	if err != nil {
		log.Error(err,fmt.Sprintf("Watch pod failed in %s",namespace))
		return
	}
	log.Info(fmt.Sprintf("Watcher pod in %s begin",namespace))
	EXIT:
	for{
		select {
		case res := <- watcher.ResultChan():
			obj := res.Object.(*corev1.Pod)
			if res.Type == watch.Deleted {
				watcherPodDeletedProcess(obj, namespace, c)
			} else {
				watcherPodProcess(obj, namespace, c)
			}
		case <-(*nsCtx).Done():
			break EXIT
		}
	}
	log.Info(fmt.Sprintf("Watcher pod in %s stop", namespace))
	return
}

func watcherPodProcess(obj *corev1.Pod, namespace string, c client.Client) (err error){
	defer func(){
		multitenancycontroller.Mutex.Unlock()
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()
	multitenancycontroller.Mutex.Lock()

	podName := obj.Name
	if NamespaceMap[namespace] != nil {
		OUT:
		for rc, rcMap := range NamespaceMap[namespace].NamespacedRCMap {
			for _, rcName := range rcMap.RCName {
				if podNameMatchesRC(rc.Kind,rcName,podName)	{
					checkMTC, err := multitenancycontroller.CheckMultiTenancyController(c, log)
					if err != nil {
						log.Error(err, "Get Controller failed")
						return
					}
					phase := getPhase(obj)
					if checkMTC.Status.ApplyNamespacedChartPodStatus(namespace,podName,phase) {
						err = c.Status().Update(context.TODO(), checkMTC)
						if err != nil {
							log.Error(err, "Update Controller failed")
							return
						}
					}
					break OUT
				}
			}

		}
	}
	return
}

func watcherPodDeletedProcess(obj *corev1.Pod, namespace string, c client.Client) (err error){
	defer func(){
		multitenancycontroller.Mutex.Unlock()
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()
	multitenancycontroller.Mutex.Lock()

	podName := obj.Name
	if NamespaceMap[namespace] != nil {
	OUT:
		for rc, rcMap := range NamespaceMap[namespace].NamespacedRCMap {
			for _, rcName := range rcMap.RCName {
				if podNameMatchesRC(rc.Kind,rcName,podName)	{
					checkMTC, err := multitenancycontroller.CheckMultiTenancyController(c, log)
					if err != nil {
						log.Error(err, "Get Controller failed")
						return
					}
					if checkMTC.Status.RemoveNamespacedChartPodStatus(namespace,podName) {
						err = c.Status().Update(context.TODO(), checkMTC)
						if err != nil {
							log.Error(err, "Update Controller failed")
							return
						}
					}
					break OUT
				}
			}

		}
	}
	return
}