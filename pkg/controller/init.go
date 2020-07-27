package controller

import (
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancyscheduler"
	"github.com/ica10888/multi-tenancy-operator/pkg/controller/multitenancycontroller/tenancywatcher"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a tenancy-manager.
	AddToManagerFuncs = append(AddToManagerFuncs, multitenancycontroller.Add)

	// Choose tenancy-scheduler constructor and tenancy-watcher constructor, like Plugins
	multitenancycontroller.LoopSchedule(tenancyscheduler.ChartSchedulerFor(), tenancywatcher.ReplicationControllerWatcherFor())
}
