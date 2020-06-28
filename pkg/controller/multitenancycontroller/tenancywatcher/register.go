package tenancywatcher

import (
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func replicationControllerRegister(clientSet *kubernetes.Clientset,c client.Client,namespace string,apiVersionRC ApiVersionRC){
	if NamespaceMap[namespace] != nil && NamespaceMap[namespace].NamespacedRCMap[apiVersionRC] != nil {
		switch apiVersionRC {
		case ApiVersionRC{"apps/v1","Deployment"}:
			go appsV1DeploymentWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1","StatefulSet"}:
			go appsV1StatefulSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1","DaemonSet"}:
			go appsV1DaemonSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1beta1","Deployment"}:
			go appsV1beta1DeploymentWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1beta1","StatefulSet"}:
			go appsV1beta1StatefulSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1beta2","Deployment"}:
			go appsV1beta2DeploymentWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1beta2","StatefulSet"}:
			go appsV1beta2StatefulSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"apps/v1beta2","DaemonSet"}:
			go appsV1beta2DaemonSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"extensions/v1beta1","Deployment"}:
			go extensionsV1beta1DeploymentWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		case ApiVersionRC{"extensions/v1beta1","DaemonSet"}:
			go extensionsV1beta1DaemonSetWatcher(clientSet,c,NamespaceMap[namespace].Ctx,NamespaceMap[namespace].NamespacedRCMap[apiVersionRC].Ctx,namespace,apiVersionRC)
		}
	}

}

func podRegister(clientSet *kubernetes.Clientset,c client.Client,namespace string){
	if NamespaceMap[namespace] != nil {
		go podWatcher(clientSet,c,NamespaceMap[namespace].Ctx,namespace)
	}
}
