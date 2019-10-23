package model

import (
	"errors"
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var log = logf.Log.WithName("git-server-model")

type GitServer struct {
	GitHost                  string
	GitUser                  string
	HttpsPort                string
	SshPort                  string
	PrivateSshKey            string
	CreateCodeReviewPipeline bool
	ActionLog                ActionLog
	Tenant                   string
	Name                     string
}

func ConvertToGitServer(k8sObj edpv1alpha1Codebase.GitServer) (*GitServer, error) {
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
		Tenant:                   strings.TrimSuffix(k8sObj.Namespace, "-edp-cicd"),
		Name:                     k8sObj.Name,
	}

	return &gitServer, nil
}

func convertGitServerActionLog(status edpv1alpha1Codebase.GitServerStatus) *ActionLog {
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
	}
}
