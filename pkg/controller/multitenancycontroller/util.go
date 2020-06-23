package multitenancycontroller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"math"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)


func checkMultiTenancyController(c client.Client, reqLogger logr.Logger) (*v1alpha1.Controller,error){
	multiTenancyControllerList := &v1alpha1.ControllerList{}
	multiTenancyController := &v1alpha1.Controller{}

	listOpts := []client.ListOption{
		client.InNamespace(metav1.NamespaceAll),
	}
	err := c.List(context.TODO(),multiTenancyControllerList,listOpts...)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("multiTenancyController resource not found. Ignoring since object must be deleted")
			return nil, err
		}
		reqLogger.Error(err, "Failed to get multiTenancyController")
		return nil, err
	}

	//multiTenancyController can not exist more than one at same time
	if len(multiTenancyControllerList.Items) >= 2 {
		oldestController := &v1alpha1.Controller{}
		var unixNano int64 =math.MinInt64
		for _, item := range multiTenancyControllerList.Items {
			if item.ObjectMeta.CreationTimestamp.UnixNano() > unixNano {
				oldestController = &item
			} else {
				c.Delete(context.TODO(),&item)
			}
		}
		err := fmt.Errorf("Controller can not exist more than one at same time, Controller is in %s namespace",oldestController.Namespace)
		reqLogger.Error(err, "Failed to create multiTenancyController")
		return nil, err
	}

	multiTenancyController = &multiTenancyControllerList.Items[0]

	return multiTenancyController,nil
}



func flatMapUpdatedTenancies(tenancies []v1alpha1.StatusTenancy) (map[NamespacedChart](map[string]string)) {
	res := make(map[NamespacedChart](map[string]string))
	if tenancies == nil {
		return res
	}
	for _, tenancy := range tenancies {
		namespace := tenancy.Namespace
		for _, chart := range tenancy.ChartMessages {
			chartName,releaseName := separateReleaseChartName(chart.ChartName)
			res[NamespacedChart{namespace,chartName,releaseName}] = chart.SettingMap
		}
	}
	return res
}






func flatMapTenancies(tenancies []v1alpha1.Tenancy) (map[NamespacedChart](map[string]string)) {
	res := make(map[NamespacedChart](map[string]string))
	if tenancies == nil {
		return res
	}
	for _, tenancy := range tenancies {
		namespace := tenancy.Namespace
		for _, chart := range tenancy.Charts {
			sets := make(map[string]string)
			for _, set := range chart.Settings {
				sets[set.Key] = set.Value
			}
			var releaseName string
			if chart.ReleaseName != nil{
				releaseName = *chart.ReleaseName
			} else {
				releaseName = ""
			}
			res[NamespacedChart{namespace,chart.ChartName,releaseName}] = sets
		}
	}
	return res
}


func equal(s1,s2 map[string]string) bool{
	if !(len(s1) == len(s2)) {
		return false
	}
	for k, _ := range s1 {
		if k != ""{
			if s1[k] != s2[k] {
				return false
			}
		}
	}
	return true
}



func separateReleaseChartName(releaseChartName string) (string,string){
	strs := strings.Split(releaseChartName,"(")
	if len(strs) == 1 {
		return releaseChartName,""
	} else {
		return strs[0], strings.ReplaceAll(strs[1], ")", "")
	}
}


func mergeReleaseChartName(chartName,releaseName string) string{
	if releaseName == "" {
		return  chartName
	} else {
		return chartName + "(" + releaseName + ")"
	}
}
