package tenancydirector

import (
	"context"
	"github.com/ghodss/yaml"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancydirector/helm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)
var log = logf.Log.WithName("tenancy_manager")


type ChartDirector struct {
	ChartHome string
}

func ChartDirectorFor() ChartDirector{
	chartHome := os.Getenv("CHART_HOME")
	if chartHome == ""{
		chartHome = "/root/chart"
	}
	return ChartDirector{
		ChartHome: chartHome,
	}
}

func (a *ChartDirector) CreateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,t.NamespacedChart.Namespace,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		return err
	}
	//TODO create namespace

	createOrUpdate(t,data)
	return nil
}

func (a *ChartDirector) UpdateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,t.NamespacedChart.Namespace,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		return err
	}
	createOrUpdate(t,data)
	return nil
}

func (a *ChartDirector) DeleteSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {

	panic("implement me")
}



func createOrUpdate(t *multitenancycontroller.TenancyExample, data string) (objs []KubeObject, err error) {
	var succObjs []KubeObject
	objs,err = Deserializer(data,t.NamespacedChart.Namespace)
	if err != nil {
		return nil, err
	}
	var errs []error
	for _, obj := range objs {
		u := &unstructured.Unstructured{}
		u.SetNamespace(t.NamespacedChart.Namespace)

		switch t.TenancyOperator {
		case multitenancycontroller.CREATE:
			err = t.Reconcile.Client.Create(context.TODO(),obj.Object)
		case multitenancycontroller.UPDATE:
			err = t.Reconcile.Client.Update(context.TODO(),obj.Object)
		}

		if err != nil {
			errs = append(errs, err)
			log.Error(err,"%s %s %s failed in %s",obj.Kubeapi.Kind,obj.Kubeapi.Name,t.TenancyOperator,t.NamespacedChart.Namespace)
		} else {
			succObjs = append(succObjs, obj)
			log.Info("%s %s %s success in %s",obj.Kubeapi.Kind,obj.Kubeapi.Name,t.TenancyOperator,t.NamespacedChart.Namespace)
		}

	}
	if len(errs) > 0 {
		return succObjs,ErrorsFmt("Failed, reason: ",errs)
	}
	return succObjs,nil
}



func Deserializer(data string,namespace string) (objs []KubeObject,err error) {
	var checkDatas []string
	datas := strings.Split(data, "---")
	for _, s := range datas {
		if !(strings.Trim(s, "\n") == "") {
			checkDatas = append(checkDatas, s)
		}
	}
	for _, s := range checkDatas {
		res, kapi, err :=serializerWithNamespace(s,namespace)
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


func serializerWithNamespace(s string,namespace string)(res []byte ,kapi Kubeapi ,err error){
	json, err :=yaml.YAMLToJSON([]byte(s))
	if err != nil {
		return
	}

	u, _, err := unstructured.UnstructuredJSONScheme.Decode(json,nil, nil)
	if err != nil {
		return
	}

	stru := u.(*unstructured.Unstructured)
	if namespace != "" {
		stru.SetNamespace(namespace)
	}
	kapi.Namespace = namespace
	kapi.Name = stru.GetName()
	kapi.Kind = stru.GetKind()
	kapi.ApiVersion = stru.GetAPIVersion()

	res ,err = u.(*unstructured.Unstructured).MarshalJSON()
	return
}
