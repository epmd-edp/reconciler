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
	JenkinsSlave        string
	JenkinsSlaveId      *int
	JobProvisioning     string
	JobProvisioningId   *int
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

	c := Codebase{
		Tenant:          strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		Name:            k8sObject.Name,
		Language:        spec.Lang,
		BuildTool:       spec.BuildTool,
		Strategy:        string(spec.Strategy),
		ActionLog:       *status,
		Type:            spec.Type,
		Status:          k8sObject.Status.Value,
		GitServer:       spec.GitServer,
		JenkinsSlave:    spec.JenkinsSlave,
		JobProvisioning: spec.JobProvisioning,
	}

	framework := spec.Framework
	if framework == "" {
		c.Framework = nil
	} else {
		lowerFramework := strings.ToLower(framework)
		c.Framework = &lowerFramework
	}

	if spec.Repository != nil {
		c.RepositoryUrl = spec.Repository.Url
	} else {
		c.RepositoryUrl = ""
	}

	if spec.Route != nil {
		c.RouteSite = spec.Route.Site
		c.RoutePath = spec.Route.Path
	} else {
		c.RouteSite = ""
		c.RoutePath = ""
	}

	if spec.Database != nil {
		c.DatabaseKind = spec.Database.Kind
		c.DatabaseVersion = spec.Database.Version
		c.DatabaseStorage = spec.Database.Storage
		c.DatabaseCapacity = spec.Database.Capacity
	} else {
		c.DatabaseKind = ""
		c.DatabaseVersion = ""
		c.DatabaseStorage = ""
		c.DatabaseCapacity = ""
	}

	if spec.Description != nil {
		c.Description = *spec.Description
	}

	if spec.TestReportFramework != nil {
		c.TestReportFramework = *spec.TestReportFramework
	}

	if spec.Strategy == "import" {
		c.GitUrlPath = spec.GitUrlPath
	}

	return &c, nil
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
