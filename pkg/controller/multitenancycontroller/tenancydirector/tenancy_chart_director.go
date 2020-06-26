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
)
var log = logf.Log.WithName("tenancy_director")


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
	releaseName := t.NamespacedChart.Namespace
	if t.NamespacedChart.ReleaseName != "" {
		releaseName = t.NamespacedChart.ReleaseName
	}

	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data,err :=helm.Template(repo,releaseName,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}

	return applyOrUpdate(t,conversionCheckDataList(data))
}

func (a ChartDirector) UpdateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) ([]multitenancycontroller.KubeObject,error) {
	releaseName := t.NamespacedChart.Namespace
	if t.NamespacedChart.ReleaseName != "" {
		releaseName = t.NamespacedChart.ReleaseName
	}

	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data, err :=helm.Template(repo,releaseName,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}
	staData, err :=helm.Template(repo,releaseName,"",false,SettingToStringValues(t.StateSettings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}

	checkDatas := conversionCheckDataList(data)
	checkStateDatas := conversionCheckDataList(staData)
	updateDatas := removeListIfNotChanged(checkDatas,checkStateDatas)

	return applyOrUpdate(t,updateDatas)
}

func (a ChartDirector) DeleteSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) ([]multitenancycontroller.KubeObject,error) {
	releaseName := t.NamespacedChart.Namespace
	if t.NamespacedChart.ReleaseName != "" {
		releaseName = t.NamespacedChart.ReleaseName
	}

	repo := path.Join(a.ChartHome,t.NamespacedChart.ChartName)
	data, err :=helm.Template(repo,releaseName,"",false,SettingToStringValues(t.Settings))
	if err != nil {
		log.Error(err,"Helm Template Error")
		return nil,err
	}
	checkDatas := conversionCheckDataList(data)

	var errs []error
	var succObjs []multitenancycontroller.KubeObject

	for _, checkData := range checkDatas {
		obj,err := Deserializer(t.Reconcile.Client,checkData,t.NamespacedChart.Namespace,false)
		if err != nil {
			errs = append(errs, err)
			break
		}
		err = t.Reconcile.Client.Delete(context.TODO(),obj.Object)
		if err != nil {
			errs = append(errs, err)
		} else {
			succObjs = append(succObjs, obj)
		}
	}
	if len(errs) > 0 {
		return succObjs,ErrorsFmt("Failed, reason: ",errs)
	}
	return succObjs,nil

}



func applyOrUpdate(t *multitenancycontroller.TenancyExample, checkDatas []string) (objs []multitenancycontroller.KubeObject, err error) {

	var errs []error
	var succObjs []multitenancycontroller.KubeObject

	for _, checkData := range checkDatas {
		var obj multitenancycontroller.KubeObject

		switch t.TenancyOperator {
		case multitenancycontroller.CREATE:
			obj,err = Deserializer(t.Reconcile.Client,checkData,t.NamespacedChart.Namespace,false)
			if err != nil {
				errs = append(errs, err)
			} else {
				err = t.Reconcile.Client.Create(context.TODO(), obj.Object)
				// TODO create namespace
				if apierrs.IsUnauthorized(err){
					return succObjs,err
				}
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
			if err != nil {
				errs = append(errs, err)
			} else {
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

	if namespace != "" {
		stru.SetNamespace(namespace)
	}

	kapi.Namespace = stru.GetNamespace()
	kapi.Name = stru.GetName()
	kapi.Kind = stru.GetKind()
	kapi.ApiVersion = stru.GetAPIVersion()

	if needResourceVersion {
		err = AddResourceVersionForUpdate(c,stru)
	}


	res ,err = stru.MarshalJSON()
	return
}

func AddResourceVersionForUpdate(c client.Client,obj *unstructured.Unstructured) (err error){
	rvObj := &unstructured.Unstructured{}

	rvObj.SetAPIVersion(obj.GetAPIVersion())
	rvObj.SetKind(obj.GetKind())
	rvObj.SetName(obj.GetName())
	rvObj.SetNamespace(obj.GetNamespace())

	err = c.Get(context.TODO(),types.NamespacedName{obj.GetNamespace(),obj.GetName()},rvObj)

	if err != nil {
		log.Error(err,"Get before update error")
		return
	}
	buf := new(bytes.Buffer)

	unstructured.UnstructuredJSONScheme.Encode(rvObj,buf)

	u, _, err := unstructured.UnstructuredJSONScheme.Decode(buf.Bytes(),nil, nil)
	if err != nil {
		log.Error(err,"Get before Get decode error")
		return
	}


	rvObjstru := u.(*unstructured.Unstructured)
	obj.SetResourceVersion(rvObjstru.GetResourceVersion())

	immutableFieldSolver(obj,rvObjstru)

	return
}


//If error like this:
//is invalid: spec.clusterIP: Invalid value: "": field is immutable
//Here add case to solve
func immutableFieldSolver(obj,rvObjstru *unstructured.Unstructured){
	switch obj.GetKind() {
	case "Service":
		val, found, err := unstructured.NestedString(rvObjstru.Object,"spec", "clusterIP")
		if !found || err != nil {
			val = ""
		}
		unstructured.SetNestedField(obj.Object,val,"spec", "clusterIP")
	}
}