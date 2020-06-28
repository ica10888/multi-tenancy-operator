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
		}
	}

}

