package model

import (
	edpv1alpha1 "business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	"errors"
	"strings"
)

type BusinessEntity struct {
	Name             string
	Tenant           string
	Type             BEType
	Language         string
	Framework        string
	BuildTool        string
	Strategy         string
	RepositoryUrl    string
	RouteSite        string
	RoutePath        string
	DatabaseKind     string
	DatabaseVersion  string
	DatabaseCapacity string
	DatabaseStorage  string
	ActionLog        ActionLog
}

type BEType string

const (
	App BEType = "application"
)

type ActionLog struct {
	Id              int
	Event           string
	DetailedMessage string
	Username        string
	UpdatedAt       int64
}

func Convert(k8sObject edpv1alpha1.BusinessApplication) (*BusinessEntity, error) {
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
		Framework: spec.Framework,
		BuildTool: spec.BuildTool,
		Strategy:  string(spec.Strategy),
		ActionLog: *status,
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

	return &app, nil
}

func convertActionLog(status edpv1alpha1.BusinessApplicationStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           formatStatus(status.Status),
		DetailedMessage: "",
		Username:        "",
		UpdatedAt:       status.LastTimeUpdated.Unix(),
	}
}

func formatStatus(status string) string {
	return strings.ToLower(strings.Replace(status, " ", "_", -1))
}
