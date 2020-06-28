package tenancywatcher

import (
	"context"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"k8s.io/client-go/kubernetes"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = logf.Log.WithName("controller_multi_watcher")

//Not thread safe
var NamespaceMap = make(map[string]*NamespacedRC)

type NamespacedRC struct {
	Ctx *context.Context
	NamespacedRCMap  map[ApiVersionRC]*NamespacedRCMap
}

type NamespacedRCMap struct {
	Ctx *context.Context
	RCName []string
}

type ApiVersionRC struct {
	ApiVersion string
	Kind string
}


type ReplicationControllerWatcher struct {
	ClientSet *kubernetes.Clientset
}

func ReplicationControllerWatcherFor() ReplicationControllerWatcher {
	return ReplicationControllerWatcher{}
}

func (w ReplicationControllerWatcher) InitTenancyWatcher(t *multitenancycontroller.TenancyExample) {
	log.Info("Init watcher")

	checkMTC,err := multitenancycontroller.CheckMultiTenancyController(t.Reconcile.Client,log)
	if err != nil {
		panic(err)
	}

	for _, ten := range checkMTC.Status.UpdatedTenancies {
		nRC := NamespaceMap[ten.Namespace]
		nsCtx := context.Background()
		nRC = &NamespacedRC{&nsCtx,nil}
		if len(ten.ReplicationControllerStatusList) > 0 {
			nRC.NamespacedRCMap = make(map[ApiVersionRC]*NamespacedRCMap)
			for _, s := range ten.ReplicationControllerStatusList {
				if nRC.NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}] == nil {
					ctx := context.Background()
					nRC.NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}] = &NamespacedRCMap{ &ctx,[]string{s.ReplicationControllerName}}
					//RC register
					replicationControllerRegister(w.ClientSet,t.Reconcile.Client,ten.Namespace,ApiVersionRC{s.ApiVersion,s.Kind})
				} else {
					list := append(nRC.NamespacedRCMap[ApiVersionRC{s.ApiVersion, s.Kind}].RCName,s.ReplicationControllerName)
					nRC.NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}].RCName = list
				}
			}
		}
		//All pod register
		for ns, _ := range NamespaceMap {
			podRegister(w.ClientSet,t.Reconcile.Client,ns)
		}
	}

}
func (w ReplicationControllerWatcher) CreateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	createOrDeleteRCStatus(w.ClientSet,objs, t)
}

func (w ReplicationControllerWatcher) DeleteTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	createOrDeleteRCStatus(w.ClientSet,objs, t)
}

func createOrDeleteRCStatus(clientSet *kubernetes.Clientset, objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	apis := getRCKubeapi(objs)
	if NamespaceMap[t.NamespacedChart.Namespace] != nil {
		m := NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap
		mtC := &v1alpha1.Controller{}
		err := t.Reconcile.Client.Get(context.TODO(), t.NamespacedController, mtC)

		if err != nil {
			log.Error(err, "Get Controller failed")
			return
		}
		for _, api := range apis {
			switch t.TenancyOperator {
			case multitenancycontroller.CREATE:
				nRCMap := m[ApiVersionRC{api.ApiVersion, api.Kind}]
				if nRCMap == nil {
					ctx := context.Background()
					nRCMap = &NamespacedRCMap{&ctx, []string{api.Name}}
					//RC register
					replicationControllerRegister(clientSet,t.Reconcile.Client,t.NamespacedChart.Namespace,ApiVersionRC{api.ApiVersion,api.Kind})
				} else {
					rc := nRCMap.RCName
					rc = append(rc, api.Name)
					nRCMap.RCName = rc
				}
				//Append CRD Controller Status
				mtC.Status.AppendNamespacedChartReplicationControllerStatusList(api.Namespace, api.Name, api.ApiVersion, api.Kind)
			case multitenancycontroller.DELETE:
				nRCMap := m[ApiVersionRC{api.ApiVersion, api.Kind}]
				if nRCMap != nil {
					for i, s := range nRCMap.RCName {
						if api.Name == s {
							list := append(nRCMap.RCName[:i], nRCMap.RCName[i+1:]...)
							if len(list) == 0 {
								if nRCMap.Ctx != nil {
									context.WithCancel(*nRCMap.Ctx)
								}
								m[ApiVersionRC{api.ApiVersion, api.Kind}] = nil
							} else {
								m[ApiVersionRC{api.ApiVersion, api.Kind}].RCName = list
							}
						}
					}
				}
				//Remove CRD Controller Status
				mtC.Status.RemoveNamespacedChartReplicationControllerStatusListIfExist(api.Namespace, api.Name, api.ApiVersion, api.Kind)
			}

		}
		err = t.Reconcile.Client.Status().Update(context.TODO(), mtC)
		if err != nil {
			log.Error(err, "Update Controller failed")
		}
	}
}


func (w ReplicationControllerWatcher) CreateTenancyNamespacesIfNeed(t *multitenancycontroller.TenancyExample) {
	for _, namespace := range t.Namespaces {
		if NamespaceMap[namespace] == nil {
			ctx := context.Background()
			NamespaceMap[namespace] = &NamespacedRC{&ctx,make(map[ApiVersionRC]*NamespacedRCMap)}
			//Pod register
			podRegister(w.ClientSet,t.Reconcile.Client,namespace)
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
			if v.Ctx != nil {
				context.WithCancel(*v.Ctx)
			}
			NamespaceMap[s] = nil
		}
	}
}
