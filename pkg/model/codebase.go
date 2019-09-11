package model

import (
	"errors"
	"fmt"
	edpv1alpha1 "github.com/epmd-edp/reconciler/v2/pkg/apis/edp/v1alpha1"
	"strings"
	"time"
)

const (
	Application CodebaseType = "application"
	Autotests   CodebaseType = "autotests"
	Library     CodebaseType = "library"
)

type CodebaseType string

type Codebase struct {
	Name                string
	Tenant              string
	Type                string
	Language            string
	Framework           *string
	BuildTool           string
	Strategy            string
	RepositoryUrl       string
	RouteSite           string
	RoutePath           string
	DatabaseKind        string
	DatabaseVersion     string
	DatabaseCapacity    string
	DatabaseStorage     string
	ActionLog           ActionLog
	Description         string
	TestReportFramework string
	Status              string
	GitServer           string
	GitUrlPath          *string
	GitServerId         *int
}

type ActionLog struct {
	Id              int
	Event           string
	DetailedMessage string
	Username        string
	UpdatedAt       time.Time
	Action          string
	ActionMessage   string
	Result          string
}

var codebaseActionMessageMap = map[string]string{
	"codebase_registration":          "Codebase %v registration",
	"accept_codebase_registration":   "Accept codebase %v registration",
	"gerrit_repository_provisioning": "Gerrit repository for codebase %v provisioning",
	"jenkins_configuration":          "CI Jenkins pipelines codebase %v provisioning",
	"perf_registration":              "Registration codebase %v in Perf",
	"setup_deployment_templates":     "Setup deployment templates for codebase %v",
}

func Convert(k8sObject edpv1alpha1.Codebase) (*Codebase, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	spec := k8sObject.Spec
	if &spec == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}

	status := convertActionLog(k8sObject.Name, k8sObject.Status)

	app := Codebase{
		Tenant:    strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		Name:      k8sObject.Name,
		Language:  spec.Lang,
		BuildTool: spec.BuildTool,
		Strategy:  string(spec.Strategy),
		ActionLog: *status,
		Type:      spec.Type,
		Status:    k8sObject.Status.Value,
		GitServer: spec.GitServer,
	}

	framework := spec.Framework
	if framework == "" {
		app.Framework = nil
	} else {
		lowerFramework := strings.ToLower(framework)
		app.Framework = &lowerFramework
	}

	if spec.Repository != nil {
		app.RepositoryUrl = spec.Repository.Url
	} else {
		app.RepositoryUrl = ""
	}

	if spec.Route != nil {
		app.RouteSite = spec.Route.Site
		app.RoutePath = spec.Route.Path
	} else {
		app.RouteSite = ""
		app.RoutePath = ""
	}

	if spec.Database != nil {
		app.DatabaseKind = spec.Database.Kind
		app.DatabaseVersion = spec.Database.Version
		app.DatabaseStorage = spec.Database.Storage
		app.DatabaseCapacity = spec.Database.Capacity
	} else {
		app.DatabaseKind = ""
		app.DatabaseVersion = ""
		app.DatabaseStorage = ""
		app.DatabaseCapacity = ""
	}

	if spec.Description != nil {
		app.Description = *spec.Description
	}

	if spec.TestReportFramework != nil {
		app.TestReportFramework = *spec.TestReportFramework
	}

	if spec.Strategy == "import" {
		app.GitUrlPath = spec.GitUrlPath
	}

	return &app, nil
}

func convertActionLog(name string, status edpv1alpha1.CodebaseStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           formatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          status.Action,
		Result:          status.Result,
		ActionMessage:   fmt.Sprintf(codebaseActionMessageMap[status.Action], name),
	}
}

func formatStatus(status string) string {
	return strings.ToLower(strings.Replace(status, " ", "_", -1))
}
