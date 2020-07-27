package multitenancycontroller

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubeObject struct {
	Kubeapi Kubeapi
	Object  runtime.Object
}

type Kubeapi struct {
	ApiVersion string
	Kind       string
	Name       string
	Namespace  string
}

func (k *Kubeapi) CreateUnstructured() *unstructured.Unstructured {
	u := new(unstructured.Unstructured)
	u.SetAPIVersion(k.ApiVersion)
	u.SetKind(k.Kind)
	u.SetName(k.Name)
	if k.Namespace != "" {
		u.SetNamespace(k.Namespace)
	}
	return u
}

type TenancyScheduler interface {
	CreateSingleTenancyByConfigure(t *TenancyExample) (objs []KubeObject, err error)
	UpdateSingleTenancyByConfigure(t *TenancyExample) (objs []KubeObject, err error)
	DeleteSingleTenancyByConfigure(t *TenancyExample) (objs []KubeObject, err error)
}

type TenancyWatcher interface {
	InitTenancyWatcher(t *TenancyExample)
	CreateTenancyPodStatusAndReplicationControllerStatus(objs []KubeObject, t *TenancyExample)
	DeleteTenancyPodStatusAndReplicationControllerStatus(objs []KubeObject, t *TenancyExample)
	CreateTenancyNamespacesIfNeed(t *TenancyExample)
	DeleteTenancyNamespacesIfNeed(t *TenancyExample)
}
