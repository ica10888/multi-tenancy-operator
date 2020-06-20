package tenancydirector

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancydirector/helm"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"path"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)
var log = logf.Log.WithName("tenancy_manager")


type ChartDirector struct {
	ChartHome string
}

func ChartDirectorFor() ChartDirector {
	chartHome := os.Getenv("CHART_HOME")
	if chartHome == ""{
		chartHome = "/root/chart"
	}
	return ChartDirector{
		ChartHome: chartHome,
	}
}

func (a ChartDirector) CreateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) ([]multitenancycontroller.KubeObject,error) {
	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,t.NamespacedChart.Namespace,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}
	//TODO create namespace

	return applyOrUpdate(t,data)
}

func (a ChartDirector) UpdateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) ([]multitenancycontroller.KubeObject,error) {
	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,t.NamespacedChart.Namespace,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}
	return applyOrUpdate(t,data)
}

func (a ChartDirector) DeleteSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) ([]multitenancycontroller.KubeObject,error) {

	panic("implement me")
}



func applyOrUpdate(t *multitenancycontroller.TenancyExample, data string) (objs []multitenancycontroller.KubeObject, err error) {

	var checkDatas []string
	datas := strings.Split(data, "---")
	for _, s := range datas {
		if !(strings.Trim(s, "\n") == "") {
			checkDatas = append(checkDatas, s)
		}
	}

	var errs []error
	var succObjs []multitenancycontroller.KubeObject

	for _, checkData := range checkDatas {
		var obj multitenancycontroller.KubeObject

		switch t.TenancyOperator {
		case multitenancycontroller.CREATE:
			obj,err = Deserializer(t.Reconcile.Client,checkData,t.NamespacedChart.Namespace,false)
			if err == nil {
				err = t.Reconcile.Client.Create(context.TODO(), obj.Object)
				if apierrs.IsAlreadyExists(err) {
					log.Info("Is already exists, try to update")

					obj, err = Deserializer(t.Reconcile.Client,checkData,t.NamespacedChart.Namespace,true)
					if err == nil{
						err = t.Reconcile.Client.Update(context.TODO(), obj.Object)
					}
				}
			}
		case multitenancycontroller.UPDATE:
			obj, err = Deserializer(t.Reconcile.Client,checkData,t.NamespacedChart.Namespace,true)
			if err == nil{
				err = t.Reconcile.Client.Update(context.TODO(), obj.Object)
			}
		}

		if err != nil {
			errs = append(errs, err)
			log.Error(err,fmt.Sprintf("%s %s %s failed in %s",obj.Kubeapi.Kind,obj.Kubeapi.Name,t.TenancyOperator.ToString(),t.NamespacedChart.Namespace))
		} else {
			succObjs = append(succObjs, obj)
			log.Info(fmt.Sprintf("%s %s %s success in %s",obj.Kubeapi.Kind,obj.Kubeapi.Name,t.TenancyOperator.ToString(),t.NamespacedChart.Namespace))
		}
	}
	if len(errs) > 0 {
		return succObjs,ErrorsFmt("Failed, reason: ",errs)
	}
	return succObjs,nil
}



func Deserializer(c client.Client,data,namespace string, needResourceVersion bool) (multitenancycontroller.KubeObject, error) {
	res, kapi, err := serializerWithNamespaceAndResourceVersionIfNeed(c,data,namespace,needResourceVersion)
		if err != nil {
			return multitenancycontroller.KubeObject{}, err
		}

		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(res, nil, nil)
		if err != nil {
			return multitenancycontroller.KubeObject{}, err
		}

	return multitenancycontroller.KubeObject{kapi,obj},nil
}


func serializerWithNamespaceAndResourceVersionIfNeed(c client.Client,s,namespace string, needResourceVersion bool)(res []byte ,kapi multitenancycontroller.Kubeapi ,err error){
	json, err :=yaml.YAMLToJSON([]byte(s))
	if err != nil {
		return
	}

	u, _, err := unstructured.UnstructuredJSONScheme.Decode(json,nil, nil)
	if err != nil {
		return
	}

	stru := u.(*unstructured.Unstructured)

	kapi.Namespace = namespace
	kapi.Name = stru.GetName()
	kapi.Kind = stru.GetKind()
	kapi.ApiVersion = stru.GetAPIVersion()

	if needResourceVersion {
		var resourceVersion string
		resourceVersion, err = GetResourceVersionForUpdate(c,stru)
		stru.SetResourceVersion(resourceVersion)
	}

	if namespace != "" {
		stru.SetNamespace(namespace)
	}

	res ,err = u.(*unstructured.Unstructured).MarshalJSON()
	return
}

func GetResourceVersionForUpdate(c client.Client,obj *unstructured.Unstructured) (res string, err error){
	deep := obj.DeepCopyObject()
	err = c.Get(context.TODO(),types.NamespacedName{obj.GetNamespace(),obj.GetName()},deep)
	if err != nil {
		return
	}
	buf := new(bytes.Buffer)
	unstructured.UnstructuredJSONScheme.Encode(deep,buf)
	u, _, err := unstructured.UnstructuredJSONScheme.Decode(buf.Bytes(),nil, nil)
	if err != nil {
		return
	}

	stru := u.(*unstructured.Unstructured)
	res = stru.GetResourceVersion()
	return
}