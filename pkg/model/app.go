package model

import (
	edpv1alpha1 "business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	"errors"
	"time"
)

type AppEntity struct {
	Namespace string
	Name      string

	Lang       string
	Framework  string
	BuildTool  string
	Strategy   string
	Repository string

	Route    *Route
	Database *Database
	Status   *Status
}

type Route struct {
	Site string
	Path string
}

type Database struct {
	Kind     string
	Version  string
	Capacity string
	Storage  string
}

type Status struct {
	Available       bool
	LastTimeUpdated time.Time
	Status          string
}

func Convert(k8sObject edpv1alpha1.BusinessApplication) (*AppEntity, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object cannot be nil")
	}
	spec := k8sObject.Spec
	if &spec == nil {
		return nil, errors.New("k8s spec cannot be nil")
	}
	route := convertRoute(spec.Route)
	db := convertDB(spec.Database)
	status := convertStatus(k8sObject.Status)
	repository := convertRepo(spec.Repository)

	app := AppEntity{
		Namespace:  k8sObject.Namespace,
		Name:       k8sObject.Name,
		Lang:       spec.Lang,
		Framework:  spec.Framework,
		BuildTool:  spec.BuildTool,
		Strategy:   string(spec.Strategy),
		Repository: repository,
		Route:      route,
		Database:   db,
		Status:     status,
	}

	return &app, nil
}

func convertRepo(repository *edpv1alpha1.Repository) string {
	if repository == nil {
		return ""
	}
	return repository.Url
}

func convertStatus(status edpv1alpha1.BusinessApplicationStatus) *Status {
	if &status == nil {
		return nil
	}
	return &Status{
		Available:       status.Available,
		LastTimeUpdated: status.LastTimeUpdated,
		Status:          status.Status,
	}
}

func convertDB(database *edpv1alpha1.Database) *Database {
	if database == nil {
		return nil
	}
	return &Database{
		Kind:     database.Kind,
		Version:  database.Version,
		Capacity: database.Capacity,
		Storage:  database.Storage,
	}
}

func convertRoute(route *edpv1alpha1.Route) *Route {
	if route == nil {
		return nil
	}
	return &Route{
		Site: route.Site,
		Path: route.Path,
	}
}
