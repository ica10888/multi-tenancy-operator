package tenancydirector

import (
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubeObject struct{
	Kubeapi Kubeapi
	Object runtime.Object
}

type Kubeapi struct {
	ApiVersion string
	Kind       string
	Name   string
	Namespace string
}

func (k *Kubeapi) CreateUnstructured() *unstructured.Unstructured{
	u := new(unstructured.Unstructured)
	u.SetAPIVersion(k.ApiVersion)
	u.SetKind(k.Kind)
	u.SetName(k.Name)
	if k.Namespace != "" {
		u.SetNamespace(k.Namespace)
	}
	return u
}


func SettingToStringValues(sets []v1alpha1.Setting) (strs []string){
	for _, set := range sets {
		strs = append(strs, set.Key + "=" + set.Value)
	}
	return
}

func ErrorsFmt (errFmt string,errs []error) (err error){
	for _, e := range errs {
		errFmt += e.Error()
	}
	return fmt.Errorf("%s",errFmt)
}