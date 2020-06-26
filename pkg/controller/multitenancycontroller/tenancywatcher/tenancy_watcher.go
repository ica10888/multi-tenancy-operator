package tenancywatcher

import (
	"context"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sync"
)




var NamespaceMap sync.Map

type NamespacedRC struct {
	Ctx context.Context
	NamespacedRCMap  map[ApiVersionRC]NamespacedRCMap
}

type NamespacedRCMap struct {
	Ctx context.Context
	ChartRCList []ChartRC
}

type ApiVersionRC struct {
	ApiVersion string
	Kind string
}

type ChartRC struct {
	ChartName string
	RCName string
}

type ReplicationControllerWatcher struct {
	ClientSet *kubernetes.Clientset
}

func ReplicationControllerWatcherFor(config *rest.Config) ReplicationControllerWatcher {
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	//TODO init watcher
	return ReplicationControllerWatcher{clientSet}
}

func (w ReplicationControllerWatcher) UpdateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {

}

func (w ReplicationControllerWatcher) UpdateTenancyNamespaces(t *multitenancycontroller.TenancyExample) {

}



