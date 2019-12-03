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

package cdpipeline

import (
	"errors"
	"fmt"
	edpv1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

type CDPipeline struct {
	Name                  string
	Namespace             string
	Tenant                string
	CodebaseBranch        []string
	InputDockerStreams    []string
	ThirdPartyServices    []string
	ActionLog             model.ActionLog
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
func ConvertToCDPipeline(k8sObject edpv1alpha1.CDPipeline, edpName string) (*CDPipeline, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object CD pipeline should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCDPipelineActionLog(k8sObject.Name, k8sObject.Status)

	cdPipeline := CDPipeline{
		Name:                  k8sObject.Spec.Name,
		Namespace:             k8sObject.Namespace,
		Tenant:                edpName,
		InputDockerStreams:    spec.InputDockerStreams,
		ThirdPartyServices:    spec.ThirdPartyServices,
		ActionLog:             *actionLog,
		Status:                k8sObject.Status.Value,
		ApplicationsToPromote: spec.ApplicationsToPromote,
	}

	return &cdPipeline, nil
}

func convertCDPipelineActionLog(cdPipelineName string, status edpv1alpha1.CDPipelineStatus) *model.ActionLog {
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

	al.ActionMessage = fmt.Sprintf(cdPipelineActionMessageMap[fmt.Sprint(status.Action)], cdPipelineName)
	return al
}
