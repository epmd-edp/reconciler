package controller

import (
	jenkins_slave "github.com/epmd-edp/reconciler/v2/pkg/controller/jenkins-slave"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, jenkins_slave.Add)
}
