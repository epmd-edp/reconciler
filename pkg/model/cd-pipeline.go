package model

import (
	edpv1alpha1 "business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	"errors"
	"strings"
)

type CDPipeline struct {
	Name           string
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
		UpdatedAt:       status.LastTimeUpdated.Unix(),
	}
}

func getStatus(eventStatus string) string {
	if eventStatus == "created" {
		return "active"
	}
	return "inactive"
}
