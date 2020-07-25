package tenancywatcher

import (
	"context"
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/apps/v1beta2"
	appsv1 "k8s.io/api/apps/v1"
    extensionsV1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/kubernetes"
	"strconv"
)

func appsV1DeploymentWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1().Deployments(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*appsv1.Deployment).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*appsv1.Deployment).Spec.Replicas != nil {
				rep = *obj.(*appsv1.Deployment).Spec.Replicas
			}
			return toReadyString(rep, obj.(*appsv1.Deployment).Status.AvailableReplicas)
		})
}

func appsV1StatefulSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1().StatefulSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*appsv1.StatefulSet).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*appsv1.StatefulSet).Spec.Replicas != nil {
				rep = *obj.(*appsv1.StatefulSet).Spec.Replicas
			}
			return toReadyString(rep, obj.(*appsv1.StatefulSet).Status.ReadyReplicas)
		})
}

func appsV1DaemonSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1().DaemonSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*appsv1.DaemonSet).Name
		},
		func(obj runtime.Object) string{
			return toDaemonSetReadyString(obj.(*appsv1.DaemonSet).Status.DesiredNumberScheduled,obj.(*appsv1.DaemonSet).Status.CurrentNumberScheduled,obj.(*appsv1.DaemonSet).Status.NumberReady,obj.(*appsv1.DaemonSet).Status.UpdatedNumberScheduled,obj.(*appsv1.DaemonSet).Status.NumberAvailable)
		})
}

func appsV1beta1DeploymentWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1beta1().Deployments(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*v1beta1.Deployment).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*v1beta1.Deployment).Spec.Replicas != nil {
				rep = *obj.(*v1beta1.Deployment).Spec.Replicas
			}
			return toReadyString(rep, obj.(*v1beta1.Deployment).Status.AvailableReplicas)
		})
}

func appsV1beta1StatefulSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1beta1().StatefulSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*v1beta1.StatefulSet).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*v1beta1.StatefulSet).Spec.Replicas != nil {
				rep = *obj.(*v1beta1.StatefulSet).Spec.Replicas
			}
			return toReadyString(rep, obj.(*v1beta1.StatefulSet).Status.ReadyReplicas)
		})
}

func appsV1beta2DeploymentWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1beta2().Deployments(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*v1beta2.Deployment).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*v1beta2.Deployment).Spec.Replicas != nil {
				rep = *obj.(*v1beta2.Deployment).Spec.Replicas
			}
			return toReadyString(rep, obj.(*v1beta2.Deployment).Status.AvailableReplicas)
		})
}

func appsV1beta2StatefulSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1beta2().StatefulSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*v1beta2.StatefulSet).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*v1beta2.StatefulSet).Spec.Replicas != nil {
				rep = *obj.(*v1beta2.StatefulSet).Spec.Replicas
			}
			return toReadyString(rep, obj.(*v1beta2.StatefulSet).Status.ReadyReplicas)
		})
}

func appsV1beta2DaemonSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.AppsV1beta2().DaemonSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*v1beta2.DaemonSet).Name
		},
		func(obj runtime.Object) string{
			return toDaemonSetReadyString(obj.(*v1beta2.DaemonSet).Status.DesiredNumberScheduled,obj.(*v1beta2.DaemonSet).Status.CurrentNumberScheduled,obj.(*v1beta2.DaemonSet).Status.NumberReady,obj.(*v1beta2.DaemonSet).Status.UpdatedNumberScheduled,obj.(*v1beta2.DaemonSet).Status.NumberAvailable)
		})
}

func extensionsV1beta1DeploymentWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.ExtensionsV1beta1().Deployments(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*extensionsV1beta1.Deployment).Name
		},
		func(obj runtime.Object) string{
			rep := int32(0)
			if obj.(*extensionsV1beta1.Deployment).Spec.Replicas != nil {
				rep = *obj.(*extensionsV1beta1.Deployment).Spec.Replicas
			}
			return toReadyString(rep, obj.(*extensionsV1beta1.Deployment).Status.AvailableReplicas)
		})
}

func extensionsV1beta1DaemonSetWatcher(clientSet *kubernetes.Clientset,c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC) (err error) {
	return rcWatcher(c,rcCtx,namespace,apiVersionRC,
		func(namespace string)(watch.Interface,error){
			return clientSet.ExtensionsV1beta1().DaemonSets(namespace).Watch(v1.ListOptions{})
		},
		func(obj runtime.Object) string{
			return obj.(*extensionsV1beta1.DaemonSet).Name
		},
		func(obj runtime.Object) string{
			return toDaemonSetReadyString(obj.(*extensionsV1beta1.DaemonSet).Status.DesiredNumberScheduled,obj.(*extensionsV1beta1.DaemonSet).Status.CurrentNumberScheduled,obj.(*extensionsV1beta1.DaemonSet).Status.NumberReady,obj.(*extensionsV1beta1.DaemonSet).Status.UpdatedNumberScheduled,obj.(*extensionsV1beta1.DaemonSet).Status.NumberAvailable)
		})
}

func rcWatcher(c client.Client,rcCtx context.Context,namespace string,apiVersionRC ApiVersionRC, watchFunc func(string)(watch.Interface,error),getRcNameFunc,getReadyFunc func(runtime.Object) string ) (err error) {
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
		case <- rcCtx.Done():
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
					return err
				}
				//TODO merge update status operators, update local status object many times , send PUT api-server once
				if checkMTC.Status.UpdateNamespacedChartReplicationControllerStatusReady(namespace, rcName, apiVersionRC.ApiVersion, apiVersionRC.Kind, ready) {
					err = c.Status().Update(context.TODO(), checkMTC)
					if err != nil {
						log.Error(err, "Update Controller failed")
						return err
					}
				}
				break
			}
		}
	}
	return
}

func toDaemonSetReadyString (desired,current,ready,update,ava int32) string {
	res := "desired: " +
		strconv.FormatInt(int64(desired),10) +
		" current: " +
		strconv.FormatInt(int64(current),10) +
		" ready: " +
		strconv.FormatInt(int64(ready),10) +
		" up-to-date: " +
		strconv.FormatInt(int64(update),10) +
		" available: " +
		strconv.FormatInt(int64(ava),10)
	return res
}

func toReadyString (rep,avaRep int32) string {
	return strconv.FormatInt(int64(avaRep),10) + "/" + strconv.FormatInt(int64(rep),10)
}