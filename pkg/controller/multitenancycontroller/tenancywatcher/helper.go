package tenancywatcher

import "github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"

func getRCKubeapi (objs []multitenancycontroller.KubeObject) (res []multitenancycontroller.Kubeapi){
	for _, obj := range objs {
		kind := obj.Kubeapi.Kind
		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			res = append(res, obj.Kubeapi)
		}
	}
	return
}
