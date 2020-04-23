package controller

import (
	jiraserver "github.com/epmd-edp/reconciler/v2/pkg/controller/jira-server"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, jiraserver.Add)
}
