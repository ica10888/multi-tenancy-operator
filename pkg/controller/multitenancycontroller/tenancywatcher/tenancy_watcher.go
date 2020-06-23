package tenancywatcher

import "github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"

type NamespacedReplicationController struct {
	Name string
	Namespace string
}

type NamespacedPod struct {
	Name string
	Namespace string
}

type ReplicationControllerWatcher struct {
}

func ReplicationControllerWatcherFor() ReplicationControllerWatcher {
	return ReplicationControllerWatcher{}
}

func (tw ReplicationControllerWatcher) UpdateTenancyPodStatusAndReplicationControllerStatus(objs []multitenancycontroller.KubeObject, t *multitenancycontroller.TenancyExample) {

}



