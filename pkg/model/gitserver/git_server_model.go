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

package gitserver

import (
	"errors"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("git-server-model")

type GitServer struct {
	GitHost                  string
	GitUser                  string
	HttpsPort                int32
	SshPort                  int32
	PrivateSshKey            string
	CreateCodeReviewPipeline bool
	ActionLog                model.ActionLog
	Tenant                   string
	Name                     string
}

func ConvertToGitServer(k8sObj edpv1alpha1Codebase.GitServer, edpName string) (*GitServer, error) {
	log.Info("Start converting GitServer", "data", k8sObj.Name)

	if &k8sObj == nil {
		return nil, errors.New("k8s git server object should not be nil")
	}
	spec := k8sObj.Spec

	actionLog := convertGitServerActionLog(k8sObj.Status)

	gitServer := GitServer{
		GitHost:                  spec.GitHost,
		GitUser:                  spec.GitUser,
		HttpsPort:                spec.HttpsPort,
		SshPort:                  spec.SshPort,
		PrivateSshKey:            spec.NameSshKeySecret,
		CreateCodeReviewPipeline: spec.CreateCodeReviewPipeline,
		ActionLog:                *actionLog,
		Tenant:                   edpName,
		Name:                     k8sObj.Name,
	}

	return &gitServer, nil
}

func convertGitServerActionLog(status edpv1alpha1Codebase.GitServerStatus) *model.ActionLog {
	if &status == nil {
		return nil
	}

	return &model.ActionLog{
		Event:           model.FormatStatus(status.Status),
		DetailedMessage: status.DetailedMessage,
		Username:        status.Username,
		UpdatedAt:       status.LastTimeUpdated,
		Action:          status.Action,
		Result:          status.Result,
	}
}
