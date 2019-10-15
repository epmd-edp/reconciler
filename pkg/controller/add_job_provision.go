package controller

import (
	jp "github.com/epmd-edp/reconciler/v2/pkg/controller/job-provisioning"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, jp.Add)
}
