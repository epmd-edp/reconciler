package model

import (
	edpv1alpha1 "business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	"errors"
	"strings"
)

type BusinessEntity struct {
	Name   string
	Tenant string
	Type   BEType
	Props  map[string]string
	Status Status
}

type BEType string

const (
	App BEType = "app"
)

type Status struct {
	Id              int
	Available       bool
	LastTimeUpdated int64
	Message         string
	Username        string
}

func Convert(k8sObject edpv1alpha1.BusinessApplication) (*BusinessEntity, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	spec := k8sObject.Spec
	if &spec == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}

	props := make(map[string]string)

	addRouteProps(props, spec.Route)
	addDBProps(props, spec.Database)
	addRepoProps(props, spec.Repository)
	addProp(props, "language", spec.Lang)
	addProp(props, "framework", spec.Framework)
	addProp(props, "build_tool", spec.BuildTool)
	addProp(props, "strategy", string(spec.Strategy))

	status := convertStatus(k8sObject.Status)

	app := BusinessEntity{
		Tenant: strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		Name:   k8sObject.Name,
		Props:  props,
		Status: *status,
	}

	return &app, nil
}

func addDBProps(props map[string]string, database *edpv1alpha1.Database) {
	if database == nil {
		return
	}
	addProp(props, "db_kind", database.Kind)
	addProp(props, "db_version", database.Version)
	addProp(props, "db_capacity", database.Capacity)
	addProp(props, "db_storage", database.Storage)
}

func addRouteProps(props map[string]string, route *edpv1alpha1.Route) {
	if route == nil {
		return
	}
	addProp(props, "route_site", route.Site)
	addProp(props, "route_path", route.Path)
}

func addRepoProps(props map[string]string, repository *edpv1alpha1.Repository) {
	if repository == nil {
		return
	}
	addProp(props, "git_url", repository.Url)
}

func addProp(props map[string]string, key string, value string) {
	if &value != nil {
		props[key] = value
	}
}

func convertStatus(status edpv1alpha1.BusinessApplicationStatus) *Status {
	if &status == nil {
		return nil
	}
	return &Status{
		Available:       status.Available,
		LastTimeUpdated: status.LastTimeUpdated.Unix(),
		Message:         status.Status,
	}
}
