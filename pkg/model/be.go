package model

import (
	"errors"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"strings"
	"time"
)

type BusinessEntity struct {
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
}

type ActionLog struct {
	Id              int
	Event           string
	DetailedMessage string
	Username        string
	UpdatedAt       time.Time
}

func Convert(k8sObject edpv1alpha1.Codebase) (*BusinessEntity, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	spec := k8sObject.Spec
	if &spec == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}

	status := convertActionLog(k8sObject.Status)

	app := BusinessEntity{
		Tenant:    strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		Name:      k8sObject.Name,
		Language:  spec.Lang,
		BuildTool: spec.BuildTool,
		Strategy:  string(spec.Strategy),
		ActionLog: *status,
		Type:      spec.Type,
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
	return &app, nil
}

func convertActionLog(status edpv1alpha1.CodebaseStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           formatStatus(status.Status),
		DetailedMessage: "",
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
	}
}

func formatStatus(status string) string {
	return strings.ToLower(strings.Replace(status, " ", "_", -1))
}
