package model

import (
	"errors"
	"fmt"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
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
	DeploymentScript    string
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

func Convert(k8sObject edpv1alpha1Codebase.Codebase) (*Codebase, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	s := k8sObject.Spec
	if &s == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}

	status := convertActionLog(k8sObject.Name, k8sObject.Status)

	c := Codebase{
		Tenant:           strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		Name:             k8sObject.Name,
		Language:         s.Lang,
		BuildTool:        s.BuildTool,
		Strategy:         string(s.Strategy),
		ActionLog:        *status,
		Type:             s.Type,
		Status:           k8sObject.Status.Value,
		GitServer:        s.GitServer,
		JenkinsSlave:     s.JenkinsSlave,
		JobProvisioning:  s.JobProvisioning,
		DeploymentScript: s.DeploymentScript,
	}

	if s.Framework != nil {
		lowerFramework := strings.ToLower(*s.Framework)
		c.Framework = &lowerFramework
	}

	if s.Repository != nil {
		c.RepositoryUrl = s.Repository.Url
	} else {
		c.RepositoryUrl = ""
	}

	if s.Route != nil {
		c.RouteSite = s.Route.Site
		c.RoutePath = s.Route.Path
	} else {
		c.RouteSite = ""
		c.RoutePath = ""
	}

	if s.Database != nil {
		c.DatabaseKind = s.Database.Kind
		c.DatabaseVersion = s.Database.Version
		c.DatabaseStorage = s.Database.Storage
		c.DatabaseCapacity = s.Database.Capacity
	} else {
		c.DatabaseKind = ""
		c.DatabaseVersion = ""
		c.DatabaseStorage = ""
		c.DatabaseCapacity = ""
	}

	if s.Description != nil {
		c.Description = *s.Description
	}

	if s.TestReportFramework != nil {
		c.TestReportFramework = *s.TestReportFramework
	}

	if s.Strategy == "import" {
		c.GitUrlPath = s.GitUrlPath
	}

	return &c, nil
}

func convertActionLog(name string, status edpv1alpha1Codebase.CodebaseStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           FormatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          string(status.Action),
		Result:          string(status.Result),
		ActionMessage:   fmt.Sprintf(codebaseActionMessageMap[string(status.Action)], name),
	}
}

func FormatStatus(status string) string {
	return strings.ToLower(strings.Replace(status, " ", "_", -1))
}
