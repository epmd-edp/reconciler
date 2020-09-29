package controller

import (
	"github.com/epmd-edp/reconciler/v2/pkg/controller/cdpipeline"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/codebasebranch"
	edpComponent "github.com/epmd-edp/reconciler/v2/pkg/controller/edp-component"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/git_server"
	jenkinsSlave "github.com/epmd-edp/reconciler/v2/pkg/controller/jenkins-slave"
	jj "github.com/epmd-edp/reconciler/v2/pkg/controller/jenkins_job"
	jiraServer "github.com/epmd-edp/reconciler/v2/pkg/controller/jira-server"
	jp "github.com/epmd-edp/reconciler/v2/pkg/controller/job-provisioning"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/stage"
	"github.com/epmd-edp/reconciler/v2/pkg/controller/thirdpartyservice"
)

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs, cdpipeline.Add, codebase.Add, codebasebranch.Add,
		edpComponent.Add, git_server.Add, jj.Add, jenkinsSlave.Add, jiraServer.Add, jp.Add, stage.Add, thirdpartyservice.Add)
}
