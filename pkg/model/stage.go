package model

import (
	"github.com/pkg/errors"
	"reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type Stage struct {
	Id             int
	Name           string
	Tenant         string
	CdPipelineName string
	Description    string
	TriggerType    string
	Order          int
	ActionLog      ActionLog
	Status         string
	QualityGates   []QualityGate
}

type QualityGate struct {
	QualityGate     string
	JenkinsStepName string
	AutotestName    *string
	BranchName      *string
}

func ConvertToStage(k8sObject v1alpha1.Stage) (*Stage, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}
	spec := k8sObject.Spec

	actionLog := convertStageActionLog(k8sObject.Status)
	status := getStatus(actionLog.Event)

	stage := Stage{
		Name:           spec.Name,
		Tenant:         strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		CdPipelineName: spec.CdPipeline,
		Description:    spec.Description,
		TriggerType:    strings.ToLower(spec.TriggerType),
		Order:          spec.Order,
		ActionLog:      *actionLog,
		Status:         status,
		QualityGates:   convertQualityGatesFromRequest(spec.QualityGates),
	}

	return &stage, nil

}

func convertQualityGatesFromRequest(gates []v1alpha1.QualityGate) []QualityGate {
	var result []QualityGate

	for _, val := range gates {
		gate := QualityGate{
			QualityGate:     strings.ToLower(val.QualityGateType),
			JenkinsStepName: strings.ToLower(val.StepName),
		}

		if gate.QualityGate == "autotests" {
			gate.AutotestName = val.AutotestName
			gate.BranchName = val.BranchName
		}

		result = append(result, gate)
	}

	return result
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
