package model

import (
	"errors"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type CDPipeline struct {
	Name           string
	Namespace      string
	Tenant         string
	CodebaseBranch []string
	ActionLog      ActionLog
	Status         string
}

func ConvertToCDPipeline(k8sObject edpv1alpha1.CDPipeline) (*CDPipeline, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object CD pipeline should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCDPipelineActionLog(k8sObject.Status)

	cdPipeline := CDPipeline{
		Name:           k8sObject.Spec.Name,
		Namespace:      k8sObject.Namespace,
		Tenant:         strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		CodebaseBranch: spec.CodebaseBranch,
		ActionLog:      *actionLog,
		Status:         getStatus(actionLog.Event),
	}

	return &cdPipeline, nil
}

func convertCDPipelineActionLog(status edpv1alpha1.CDPipelineStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           formatStatus(status.Status),
		DetailedMessage: "",
		Username:        "",
		UpdatedAt:       status.LastTimeUpdated,
	}
}

func getStatus(eventStatus string) string {
	if eventStatus == "created" {
		return "active"
	}
	return "inactive"
}
