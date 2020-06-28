package tenancywatcher

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	corev1 "k8s.io/api/core/v1"
	"regexp"
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

func podNameMatchesRC (kind,rcName,podName string) bool {
	switch kind {
	case "Deployment":
		p := rcName + "-[a-z0-9]{10}-[a-z0-9]{5}"
		m, _ := regexp.MatchString(p, podName)
		return m
	case "StatefulSet":
		p := rcName + "-[0-9]+"
		m, _ := regexp.MatchString(p, podName)
		return m
	case "DaemonSet":
		p := rcName + "-[a-z0-9]{5}"
		m, _ := regexp.MatchString(p, podName)
		return m
	}
	return false
}

func getPhase(obj *corev1.Pod) string {
	csList := obj.Status.ContainerStatuses
	if len(csList) > 0 {
		if csList[0].LastTerminationState.Waiting.Reason != ""{
			return csList[0].LastTerminationState.Waiting.Reason
		}
		if csList[0].LastTerminationState.Terminated.Reason != ""{
			return csList[0].LastTerminationState.Terminated.Reason
		}
	}
	res := ""
	switch obj.Status.Phase {
	case corev1.PodPending:
		res = "Pending"
	case corev1.PodRunning:
		res = "Running"
	case corev1.PodSucceeded:
		res = "Succeeded"
	case corev1.PodFailed:
		res = "Failed"
	case corev1.PodUnknown:
		res = "Unknown"
	}
	if res == "Running" {
		if len(csList) > 0 {
			if csList[0].Ready {
				return res
			} else {
				return "Unavailable"
			}
		}
	}
	return res
}