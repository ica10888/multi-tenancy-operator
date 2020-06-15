package tenancymanager

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"os"
)

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

func (a ChartManager) CreateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	panic("implement me")
}

func (a ChartManager) UpdateSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	panic("implement me")
}

func (a ChartManager) DeleteSingleTenancyByConfigure(t *multitenancycontroller.TenancyExample) error {
	panic("implement me")
}