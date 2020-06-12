package controller

import "github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, multitenancycontroller.Add)
}