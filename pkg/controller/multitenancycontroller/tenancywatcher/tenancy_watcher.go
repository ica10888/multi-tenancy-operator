package tenancywatcher

import (
	"context"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var NamespaceMap map[string]*NamespacedRC

type NamespacedRC struct {
	Ctx context.Context
	NamespacedRCMap  map[ApiVersionRC]*NamespacedRCMap
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

func (w ReplicationControllerWatcher) CreateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	apis := getRCKubeapi(objs)
	if NamespaceMap[t.NamespacedChart.Namespace] != nil {
		m := NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap
		for _, api := range apis {
			if m[ApiVersionRC{api.ApiVersion, api.Kind}] == nil {

			}
		}
	}
}

func (w ReplicationControllerWatcher) UpdateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {

}

func (w ReplicationControllerWatcher) DeleteTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {

}

func (w ReplicationControllerWatcher) CreateTenancyNamespacesIfNeed(t *multitenancycontroller.TenancyExample) {
	for _, namespace := range t.Namespaces {
		if NamespaceMap[namespace] == nil {
			NamespaceMap[namespace] = &NamespacedRC{context.Background(),make(map[ApiVersionRC]*NamespacedRCMap)}
		}
	}

}

func (w ReplicationControllerWatcher) DeleteTenancyNamespacesIfNeed(t *multitenancycontroller.TenancyExample) {
	needDelete := []string{}
	for n := range NamespaceMap {
		for _, namespace := range t.Namespaces {
			if n == namespace{
				break
			}
			needDelete = append(needDelete, namespace)
		}
		for _, s := range needDelete {
			v := NamespaceMap[s]
			v.Ctx.Done()
			NamespaceMap[s] = nil
		}
	}
}
