package controller

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancymanager"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a tenancymanager.
	AddToManagerFuncs = append(AddToManagerFuncs, multitenancycontroller.Add)
	multitenancycontroller.LoopSchedule(tenancymanager.ChartManagerFor())
}