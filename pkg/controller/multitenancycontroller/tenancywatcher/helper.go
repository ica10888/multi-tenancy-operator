package tenancywatcher

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"strings"
)

func getRCKubeapi (objs []multitenancycontroller.KubeObject) (res []multitenancycontroller.Kubeapi){
	for _, obj := range objs {
		kind := obj.Kubeapi.Kind
		if kind == "Deployment" || kind == "StatefulSet" || kind == "DaemonSet" {
			res = append(res, obj.Kubeapi)
		}
	}
	return
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
