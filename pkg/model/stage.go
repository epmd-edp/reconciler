package model

import (
	"github.com/pkg/errors"
	"reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type Stage struct {
	Name            string
	Tenant          string
	CdPipelineName  string
	Description     string
	TriggerType     string
	QualityGate     string
	JenkinsStepName string
	Order           int
	ActionLog       ActionLog
	Status          string
	Autotests       []string
}

func ConvertToStage(k8sObject v1alpha1.Stage) (*Stage, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}
	spec := k8sObject.Spec

	actionLog := convertStageActionLog(k8sObject.Status)
	status := getStatus(actionLog.Event)

	return &Stage{
		Name:            spec.Name,
		Tenant:          strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		CdPipelineName:  spec.CdPipeline,
		Description:     spec.Description,
		TriggerType:     strings.ToLower(spec.TriggerType),
		QualityGate:     strings.ToLower(spec.QualityGate),
		JenkinsStepName: spec.JenkinsStep,
		Order:           spec.Order,
		ActionLog:       *actionLog,
		Status:          status,
		Autotests:       spec.Autotests,
	}, nil

}

func convertStageActionLog(status v1alpha1.StageStatus) *ActionLog {
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
