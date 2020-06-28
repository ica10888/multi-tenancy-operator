package tenancywatcher

import (
	"context"
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/apis/apps"
	"strconv"
)




func appsV1DeploymentWatcher(clientSet *kubernetes.Clientset,c client.Client,nsCtx,rcCtx *context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {

	return watcher(c,nsCtx,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1().Deployments(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*apps.Deployment).Name
		},
		func(obj runtime.Object) string{
			return toReadyString(obj.(*apps.Deployment).Spec.Replicas, obj.(*apps.Deployment).Status.AvailableReplicas)
		})
}



func watcher(c client.Client,nsCtx,rcCtx *context.Context,namespace string,apiVersionRC ApiVersionRC, watchFunc func(string)(watch.Interface,error),getRcNameFunc,getReadyFunc func(runtime.Object) string ) (err error) {
	watcher, err := watchFunc(namespace)
	if err != nil {
		log.Error(err,fmt.Sprintf("Watch %s %s failed in %s",apiVersionRC.ApiVersion,apiVersionRC.Kind,namespace))
		return
	}
	log.Info(fmt.Sprintf("Watcher %s %s in %s begin",apiVersionRC.ApiVersion,apiVersionRC.Kind,namespace))
	EXIT:
	for{
		select {
		case res := <- watcher.ResultChan():
			obj := res.Object
			watcherProcess(obj, namespace, apiVersionRC, c, getRcNameFunc, getReadyFunc)
		case <-(*nsCtx).Done():
			break EXIT
		case <-(*rcCtx).Done():
			break EXIT
		}
	}
	log.Info(fmt.Sprintf("Watcher %s %s in %s stop", apiVersionRC.ApiVersion, apiVersionRC.Kind, namespace))
	return
}



func watcherProcess(obj runtime.Object, namespace string, apiVersionRC ApiVersionRC, c client.Client,getRcNameFunc,getReadyFunc func(runtime.Object) string) (err error) {
	defer func(){
		multitenancycontroller.Mutex.Unlock()
		if err := recover(); err != nil {
			log.Error(fmt.Errorf("%s",err),"recover Err")
		}
	}()
	multitenancycontroller.Mutex.Lock()

	rcName := getRcNameFunc(obj)
	if NamespaceMap[namespace] != nil && NamespaceMap[namespace].NamespacedRCMap[apiVersionRC] != nil {
		for _, s := range NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].RCName {
			if rcName == s {
				ready := getReadyFunc(obj)
				checkMTC, err := multitenancycontroller.CheckMultiTenancyController(c, log)
				if err != nil {
					log.Error(err, "Get Controller failed")
					return
				}
				checkMTC.Status.UpdateNamespacedChartReplicationControllerStatusReady(namespace, rcName, apiVersionRC.ApiVersion, apiVersionRC.Kind, ready)
				err = c.Status().Update(context.TODO(), checkMTC)
				if err != nil {
					log.Error(err, "Update Controller failed")
					return
				}
			}
		}
	}
	return
}

func toReadyString (rep,avaRep int32) string {
	return strconv.FormatInt(int64(avaRep),10) + "/" + strconv.FormatInt(int64(rep),10)
}