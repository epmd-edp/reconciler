package controller

import (
	"github.com/epmd-edp/reconciler/v2/pkg/controller/template"
	"github.com/epmd-edp/reconciler/v2/pkg/service/platform"
)

func init() {
	if platform.IsOpenshift() {
		// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
		AddToManagerFuncs = append(AddToManagerFuncs, template.Add)
	}
}
