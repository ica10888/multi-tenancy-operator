package multitenancycontroller

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/apis/multitenancy/v1alpha1"
	"strings"
)

func flatMapUpdatedTenancies(tenancies []v1alpha1.StatusTenancy) map[NamespacedChart](map[string]string) {
	res := make(map[NamespacedChart](map[string]string))
	if tenancies == nil {
		return res
	}
	for _, tenancy := range tenancies {
		namespace := tenancy.Namespace
		for _, chart := range tenancy.ChartMessages {
			chartName, releaseName := separateReleaseChartName(chart.ChartName)
			res[NamespacedChart{namespace, chartName, releaseName}] = chart.SettingMap
		}
	}
	return res
}

func flatMapTenancies(tenancies []v1alpha1.Tenancy) map[NamespacedChart](map[string]string) {
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
			if chart.ReleaseName != nil {
				releaseName = *chart.ReleaseName
			} else {
				releaseName = ""
			}
			res[NamespacedChart{namespace, chart.ChartName, releaseName}] = sets
		}
	}
	return res
}

func equal(s1, s2 map[string]string) bool {
	if !(len(s1) == len(s2)) {
		return false
	}
	for k, _ := range s1 {
		if k != "" {
			if s1[k] != s2[k] {
				return false
			}
		}
	}
	return true
}

func equalTenancies(t1, t2 []v1alpha1.Tenancy) bool {
	if len(t1) != len(t2) {
		return false
	}
	for i := range t1 {
		if t1[i].Namespace != t2[i].Namespace || len(t1[i].Charts) != len(t2[i].Charts) {
			return false
		}
		for j := range t1[i].Charts {
			if t1[i].Charts[j].ChartName != t2[i].Charts[j].ChartName || !equalStringPointer(t1[i].Charts[j].ReleaseName, t2[i].Charts[j].ReleaseName) || len(t1[i].Charts[j].Settings) != len(t2[i].Charts[j].Settings) {
				return false
			}
			for k := range t1[i].Charts[j].Settings {
				if t1[i].Charts[j].Settings[k] != t2[i].Charts[j].Settings[k] {
					return false
				}
			}
		}
	}
	return true
}

func separateReleaseChartName(releaseChartName string) (string, string) {
	strs := strings.Split(releaseChartName, "(")
	if len(strs) == 1 {
		return releaseChartName, ""
	} else {
		return strs[0], strings.ReplaceAll(strs[1], ")", "")
	}
}

func mergeReleaseChartName(chartName, releaseName string) string {
	if releaseName == "" {
		return chartName
	} else {
		return chartName + "(" + releaseName + ")"
	}
}

func equalStringPointer(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a != nil && *a == *b {
		return true
	}
	return false
}
