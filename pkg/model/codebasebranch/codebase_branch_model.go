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

package codebasebranch

import (
	"errors"
	"fmt"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

type CodebaseBranch struct {
	Name             string
	Tenant           string
	AppName          string
	FromCommit       string
	Version          *string
	BuildNumber      *string
	LastSuccessBuild *string
	Release          bool
	Status           string
	ActionLog        model.ActionLog
}

var codebaseBranchActionMessageMap = map[string]string{
	"jenkins_configuration":               "CI Jenkins pipelines for codebase branch %v provisioning for codebase %v",
	"codebase_branch_registration":        "Branch %v for codebase %v registration",
	"accept_codebase_branch_registration": "Accept branch %v for codebase %v registration",
}

func ConvertToCodebaseBranch(k8sObject edpv1alpha1Codebase.CodebaseBranch, edpName string) (*CodebaseBranch, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object application branch object should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCodebaseBranchActionLog(spec.BranchName, spec.CodebaseName, k8sObject.Status)

	branch := CodebaseBranch{
		Name:             spec.BranchName,
		Tenant:           edpName,
		AppName:          spec.CodebaseName,
		FromCommit:       spec.FromCommit,
		Version:          spec.Version,
		Release:          spec.Release,
		BuildNumber:      k8sObject.Status.Build,
		LastSuccessBuild: k8sObject.Status.LastSuccessfulBuild,
		Status:           k8sObject.Status.Value,
		ActionLog:        *actionLog,
	}

	return &branch, nil
}

func convertCodebaseBranchActionLog(brName, cbName string, status edpv1alpha1Codebase.CodebaseBranchStatus) *model.ActionLog {
	if &status == nil {
		return nil
	}

	al := &model.ActionLog{
		Event:           model.FormatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          string(status.Action),
		Result:          string(status.Result),
	}

	if status.Result == "error" {
		al.ActionMessage = status.DetailedMessage
		return al
	}

	al.ActionMessage = fmt.Sprintf(codebaseBranchActionMessageMap[string(status.Action)], brName, cbName)
	return al
}
