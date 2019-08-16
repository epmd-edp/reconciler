package model

import (
	"fmt"
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

var cdStageActionMessageMap = map[string]string{
	"accept_cd_stage_registration":      "Accept CD Stage %v registration",
	"fetching_user_settings_config_map": "Fetch User Settings from config map during CD Stage %v provision",
	"openshift_project_creation":        "Creation of Openshift Project %v",
	"jenkins_configuration":             "CI Jenkins pipelines codebase %v provisioning",
	"setup_deployment_templates":        "Setup deployment templates for cd_stage %v",
}

// ConvertToStage returns converted to DTO Stage object from K8S.
// An error occurs if method received nil instead of k8s object
func ConvertToStage(k8sObject v1alpha1.Stage) (*Stage, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}
	spec := k8sObject.Spec

	actionLog := convertStageActionLog(k8sObject.Name, k8sObject.Status)

	stage := Stage{
		Name:           spec.Name,
		Tenant:         strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		CdPipelineName: spec.CdPipeline,
		Description:    spec.Description,
		TriggerType:    strings.ToLower(spec.TriggerType),
		Order:          spec.Order,
		ActionLog:      *actionLog,
		Status:         k8sObject.Status.Value,
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

func convertStageActionLog(cdStageName string, status v1alpha1.StageStatus) *ActionLog {
	if &status == nil {
		return nil
	}

	return &ActionLog{
		Event:           formatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          status.Action,
		Result:          status.Result,
		ActionMessage:   fmt.Sprintf(cdStageActionMessageMap[status.Action], cdStageName),
	}
}
