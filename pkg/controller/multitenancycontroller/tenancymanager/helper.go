package tenancymanager

import (
	"fmt"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type KubeObject struct{
	Kubeapi Kubeapi
	Object runtime.Object
}

type Kubeapi struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name   Metadata `yaml:"metadata"`
}

type Metadata struct {
	Name string `yaml:"name"`
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