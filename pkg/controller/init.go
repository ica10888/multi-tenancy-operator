package controller

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancydirector"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a tenancydirector.
	AddToManagerFuncs = append(AddToManagerFuncs, multitenancycontroller.Add)
	// Choose tenancy-director constructor, like Plugins
	multitenancycontroller.LoopSchedule(tenancydirector.ChartDirectorFor())
}