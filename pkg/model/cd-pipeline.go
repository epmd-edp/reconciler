package model

import (
	"errors"
	"fmt"
	edpv1alpha1 "github.com/epmd-edp/reconciler/v2/pkg/apis/edp/v1alpha1"
	"strings"
)

type CDPipeline struct {
	Name                  string
	Namespace             string
	Tenant                string
	CodebaseBranch        []string
	InputDockerStreams    []string
	ThirdPartyServices    []string
	ActionLog             ActionLog
	Status                string
	ApplicationsToPromote []string
}

var cdPipelineActionMessageMap = map[string]string{
	"accept_cd_pipeline_registration": "Accept CD Pipeline %v registration",
	"jenkins_configuration":           "CI Jenkins pipelines %v provisioning",
	"setup_initial_structure":         "Initial structure for CD Pipeline %v is created",
	"cd_pipeline_registration":        "CD Pipeline %v registration",
	"create_jenkins_directory":        "Create directory in Jenkins for CD Pipeline %v",
}

// ConvertToCDPipeline returns converted to DTO CDPipeline object from K8S.
// An error occurs if method received nil instead of k8s object
func ConvertToCDPipeline(k8sObject edpv1alpha1.CDPipeline) (*CDPipeline, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object CD pipeline should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCDPipelineActionLog(k8sObject.Name, k8sObject.Status)

	cdPipeline := CDPipeline{
		Name:                  k8sObject.Spec.Name,
		Namespace:             k8sObject.Namespace,
		Tenant:                strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		InputDockerStreams:    spec.InputDockerStreams,
		ThirdPartyServices:    spec.ThirdPartyServices,
		ActionLog:             *actionLog,
		Status:                k8sObject.Status.Value,
		ApplicationsToPromote: spec.ApplicationsToPromote,
	}

	return &cdPipeline, nil
}

func convertCDPipelineActionLog(cdPipelineName string, status edpv1alpha1.CDPipelineStatus) *ActionLog {
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
		ActionMessage:   fmt.Sprintf(cdPipelineActionMessageMap[status.Action], cdPipelineName),
	}
}
