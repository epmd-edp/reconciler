/*
 * Copyright 2019 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package codebase

import (
	"errors"
	"fmt"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"strings"
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
	ActionLog           model.ActionLog
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
	VersioningType      string
	StartVersioningFrom *string
}

var codebaseActionMessageMap = map[string]string{
	"codebase_registration":          "Codebase %v registration",
	"accept_codebase_registration":   "Accept codebase %v registration",
	"gerrit_repository_provisioning": "Gerrit repository for codebase %v provisioning",
	"jenkins_configuration":          "CI Jenkins pipelines codebase %v provisioning",
	"perf_registration":              "Registration codebase %v in Perf",
	"setup_deployment_templates":     "Setup deployment templates for codebase %v",
	"put_s2i":                        "Put s2i for %v codebase",
	"put_jenkins_folder":             "Put JenkinsFolder CR for %v codebase",
	"clean_data":                     "Clean temporary data for %v codebase",
	"import_project":                 "Start importing project %v",
}

func Convert(k8sObject edpv1alpha1Codebase.Codebase, edpName string) (*Codebase, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	s := k8sObject.Spec
	if &s == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}

	status := convertActionLog(k8sObject.Name, k8sObject.Status)

	c := Codebase{
		Tenant:              edpName,
		Name:                k8sObject.Name,
		Language:            s.Lang,
		BuildTool:           s.BuildTool,
		Strategy:            string(s.Strategy),
		ActionLog:           *status,
		Type:                s.Type,
		Status:              k8sObject.Status.Value,
		GitServer:           s.GitServer,
		JenkinsSlave:        s.JenkinsSlave,
		JobProvisioning:     s.JobProvisioning,
		DeploymentScript:    s.DeploymentScript,
		VersioningType:      string(s.Versioning.Type),
		StartVersioningFrom: s.Versioning.StartFrom,
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

func convertActionLog(name string, status edpv1alpha1Codebase.CodebaseStatus) *model.ActionLog {
	if &status == nil {
		return nil
	}

	al := &model.ActionLog{
		Event:           model.FormatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          string(status.Action),
		Result:          string(status.Result),
	}

	if status.Result == "error" {
		al.ActionMessage = status.DetailedMessage
		return al
	}

	al.ActionMessage = fmt.Sprintf(codebaseActionMessageMap[string(status.Action)], name)
	return al
}
