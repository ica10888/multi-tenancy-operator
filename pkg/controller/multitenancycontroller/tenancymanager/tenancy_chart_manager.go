package tenancymanager

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancymanager/helm"
	yaml2 "gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)
var log = logf.Log.WithName("tenancy_manager")


type ChartManager struct {
	ChartHome string
}

func ChartManagerFor() ChartManager{
	chartHome := os.Getenv("CHART_HOME")
	if chartHome == ""{
		chartHome = "/root/chart"
	}
	return ChartManager{
		ChartHome: chartHome,
	}
}

func (a *ChartManager) CreateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,t.NamespacedChart.Namespace,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		return err
	}
	Create(t,data,"")
	return nil
}

func (a *ChartManager) UpdateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	panic("implement me")
}

func (a *ChartManager) DeleteSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	panic("implement me")
}



func Create(r *multitenancycontroller.TenancyExample,namespace string,data string) (objs []KubeObject,err error) {
	var succObjs []KubeObject
	objs,err = Deserializer(data)
	if err != nil {
		return nil, err
	}
	var errs []error
	for _, obj := range objs {

		u := &unstructured.Unstructured{}
		u.SetNamespace(namespace)
		err = r.Reconcile.Client.Create(context.TODO(),obj.Object)

		if err != nil {
			errs = append(errs, err)
			log.Error(err,"%s %s %s failed",obj.Kubeapi.Kind,obj.Kubeapi.Name,r.TenancyOperator)
		} else {
			succObjs = append(succObjs, obj)
			log.Info("%s %s %s success",obj.Kubeapi.Kind,obj.Kubeapi.Name,r.TenancyOperator)
		}

	}
	if len(errs) > 0 {
		return succObjs,ErrorsFmt("Failed, reason: ",errs)
	}
	return succObjs,nil
}



func Deserializer(data string) (objs []KubeObject,err error) {
	var checkDatas []string
	datas := strings.Split(data, "---")
	for _, s := range datas {
		if !(strings.Trim(s, "\n") == "") {
			checkDatas = append(checkDatas, s)
		}
	}
	for _, s := range checkDatas {
		kapi := Kubeapi{}
		yaml2.Unmarshal([]byte(s), &kapi)

		res ,err :=serializerWithNamespace(s,"dev")
		if err != nil {
			return objs, err
		}

		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(res, nil, nil)
		if err != nil {
			return objs, err
		}
		objs = append(objs, KubeObject{kapi,obj})
	}
	return
}


func serializerWithNamespace(s string,namespace string)(res []byte ,err error){
	json, err :=yaml.YAMLToJSON([]byte(s))
	if err != nil {
		return
	}

	u, _, err := unstructured.UnstructuredJSONScheme.Decode(json,nil, nil)
	if err != nil {
		return
	}
	if namespace != "" {
		u.(*unstructured.Unstructured).SetNamespace(namespace)
	}

	return u.(*unstructured.Unstructured).MarshalJSON()
}
