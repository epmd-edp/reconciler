package controller

import (
	"business-app-reconciler-controller/pkg/controller/applicationbranch"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, applicationbranch.Add)
}
