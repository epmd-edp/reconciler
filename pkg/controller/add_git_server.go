package controller

import (
	"github.com/epmd-edp/reconciler/v2/pkg/controller/git_server"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, git_server.Add)
}
