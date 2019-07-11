package model

import (
	"github.com/pkg/errors"
	"reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type Stage struct {
	Id              int
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
	Autotests       []AutotestCreateCommand
}

type AutotestCreateCommand struct {
	AutotestName string
	BranchName   string
}

func ConvertToStage(k8sObject v1alpha1.Stage) (*Stage, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}
	spec := k8sObject.Spec

	actionLog := convertStageActionLog(k8sObject.Status)
	status := getStatus(actionLog.Event)

	stage := Stage{
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
	}

	if stage.QualityGate == "autotests" {
		for _, autotestDto := range spec.Autotests {
			stage.Autotests = appendOrCreateAutotest(stage.Autotests, autotestDto)
		}
	}

	return &stage, nil

}

func appendOrCreateAutotest(target []AutotestCreateCommand, autotestDto v1alpha1.AutotestCreateCommand) []AutotestCreateCommand {
	autotest := AutotestCreateCommand{
		AutotestName: autotestDto.AutotestName,
		BranchName:   autotestDto.BranchName,
	}

	if target == nil {
		return []AutotestCreateCommand{autotest}
	}

	return append(target, autotest)
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
