/*
 * Copyright 2019 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package stage

import (
	"fmt"
	"github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/pkg/errors"
	"strings"
)

type Stage struct {
	Id             int
	Name           string
	Tenant         string
	Namespace      string
	CdPipelineName string
	Description    string
	TriggerType    string
	Order          int
	ActionLog      model.ActionLog
	Status         string
	QualityGates   []QualityGate
	Source         Source
}

type Source struct {
	Type    string
	Library Library
}

type Library struct {
	Id       *int
	BranchId *int
	Name     string
	Branch   string
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
	"platform_project_creation":         "Create Platform Project for Stage %v",
	"jenkins_configuration":             "CI Jenkins pipelines %v provisioning",
	"setup_deployment_templates":        "Setup deployment templates for cd_stage %v",
	"create_jenkins_pipeline":           "Create Jenkins pipeline for CD Stage %v",
}

// ConvertToStage returns converted to DTO Stage object from K8S and provided edp name
// An error occurs if method received nil instead of k8s object
func ConvertToStage(k8sObject v1alpha1.Stage, edpName string) (*Stage, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}
	spec := k8sObject.Spec
	actionLog := convertStageActionLog(k8sObject.Name, k8sObject.Status)
	stage := Stage{
		Name:           spec.Name,
		Tenant:         edpName,
		Namespace:      k8sObject.Namespace,
		CdPipelineName: spec.CdPipeline,
		Description:    spec.Description,
		TriggerType:    strings.ToLower(spec.TriggerType),
		Order:          spec.Order,
		ActionLog:      *actionLog,
		Status:         k8sObject.Status.Value,
		QualityGates:   convertQualityGatesFromRequest(spec.QualityGates),
		Source: Source{
			Type: spec.Source.Type,
			Library: Library{
				Name:   spec.Source.Library.Name,
				Branch: spec.Source.Library.Branch,
			},
		},
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

func convertStageActionLog(cdStageName string, status v1alpha1.StageStatus) *model.ActionLog {
	if &status == nil {
		return nil
	}

	al := &model.ActionLog{
		Event:           model.FormatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          fmt.Sprint(status.Action),
		Result:          fmt.Sprint(status.Result),
	}

	if status.Result == "error" {
		al.ActionMessage = status.DetailedMessage
		return al
	}

	al.ActionMessage = fmt.Sprintf(cdStageActionMessageMap[fmt.Sprint(status.Action)], cdStageName)
	return al
}
