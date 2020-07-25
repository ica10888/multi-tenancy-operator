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

var localClientSet *kubernetes.Clientset

type NamespacedRC struct {
	Ctx context.Context
	CancelFunc context.CancelFunc
	NamespacedRCMap  map[ApiVersionRC]*NamespacedRCMap
}

type NamespacedRCMap struct {
	Ctx context.Context
	CancelFunc context.CancelFunc
	RCName []string
}

type ApiVersionRC struct {
	ApiVersion string
	Kind string
}


type ReplicationControllerWatcher struct {
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

	localClientSet, err = kubernetes.NewForConfig(t.Reconcile.Config)
	if err != nil {
		panic(err)
	}
	for _, ten := range checkMTC.Status.UpdatedTenancies {
		nsCtx, nsCancel := context.WithCancel(context.Background())
		NamespaceMap[ten.Namespace] = &NamespacedRC{nsCtx,nsCancel,nil}
		if len(ten.ReplicationControllerStatusList) > 0 {
			NamespaceMap[ten.Namespace].NamespacedRCMap = make(map[ApiVersionRC]*NamespacedRCMap)
			for _, s := range ten.ReplicationControllerStatusList {
				if NamespaceMap[ten.Namespace].NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}] == nil {
					ctx ,cancel := context.WithCancel(nsCtx)
					NamespaceMap[ten.Namespace].NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}] = &NamespacedRCMap{ ctx,cancel,[]string{s.ReplicationControllerName}}
					//RC register
					replicationControllerRegister(localClientSet,t.Reconcile.Client,ten.Namespace,ApiVersionRC{s.ApiVersion,s.Kind})
				} else {
					list := append(NamespaceMap[ten.Namespace].NamespacedRCMap[ApiVersionRC{s.ApiVersion, s.Kind}].RCName,s.ReplicationControllerName)
					NamespaceMap[ten.Namespace].NamespacedRCMap[ApiVersionRC{s.ApiVersion,s.Kind}].RCName = list
				}
			}
		}
		//All pod register
		for ns, _ := range NamespaceMap {
			podRegister(localClientSet,t.Reconcile.Client,ns)
		}
	}

}
func (w ReplicationControllerWatcher) CreateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	createOrDeleteRCStatus(localClientSet,objs, t)
}

func (w ReplicationControllerWatcher) DeleteTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	createOrDeleteRCStatus(localClientSet,objs, t)
}

func createOrDeleteRCStatus(clientSet *kubernetes.Clientset, objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {
	apis := getRCKubeapi(objs)
	if NamespaceMap[t.NamespacedChart.Namespace] != nil {
		mtC := &v1alpha1.Controller{}
		err := t.Reconcile.Client.Get(context.TODO(), t.NamespacedController, mtC)

		if err != nil {
			log.Error(err, "Get Controller failed")
			return
		}
		for _, api := range apis {
			switch t.TenancyOperator {
			case multitenancycontroller.CREATE:
				if NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}] == nil {
					ctx ,cancel := context.WithCancel(NamespaceMap[t.NamespacedChart.Namespace].Ctx)
					NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}] = &NamespacedRCMap{ctx,cancel, []string{api.Name}}
					//RC register
					replicationControllerRegister(clientSet,t.Reconcile.Client,t.NamespacedChart.Namespace,ApiVersionRC{api.ApiVersion,api.Kind})
				} else {
					rc := NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}].RCName
					rc = append(rc, api.Name)
					NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}].RCName = rc
				}
				//Append CRD Controller Status
				mtC.Status.AppendNamespacedChartReplicationControllerStatusList(api.Namespace, api.Name, api.ApiVersion, api.Kind)
			case multitenancycontroller.DELETE:
				nRCMap := NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}]
				if nRCMap != nil {
					for i, s := range nRCMap.RCName {
						if api.Name == s {
							list := append(nRCMap.RCName[:i], nRCMap.RCName[i+1:]...)
							if len(list) == 0 {
								nRCMap.CancelFunc()
								NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}] = nil
							} else {
								NamespaceMap[t.NamespacedChart.Namespace].NamespacedRCMap[ApiVersionRC{api.ApiVersion, api.Kind}].RCName = list
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
			ctx ,cancel := context.WithCancel(context.Background())
			NamespaceMap[namespace] = &NamespacedRC{ctx ,cancel,make(map[ApiVersionRC]*NamespacedRCMap)}
			//Pod register
			podRegister(localClientSet,t.Reconcile.Client,namespace)
		}
	}

}

func (w ReplicationControllerWatcher) DeleteTenancyNamespacesIfNeed(t *multitenancycontroller.TenancyExample) {
	needDelete := []string{}
	for n := range NamespaceMap {
		isExist := false
		for _, namespace := range t.Namespaces {
			if n == namespace {
				isExist = true
			}
		}
		if !isExist {
			needDelete = append(needDelete, n)
		}
	}
	for _, s := range needDelete {
		NamespaceMap[s].CancelFunc()
		NamespaceMap[s] = nil
	}

}
