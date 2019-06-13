package model

import (
	"errors"
	"fmt"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type CodebaseBranch struct {
	Name       string
	Tenant     string
	AppName    string
	FromCommit string
	Status     string
	ActionLog  ActionLog
}

var codebaseBranchActionMessageMap = map[string]string{
	"jenkins_configuration":               "CI Jenkins pipelines for codebase branch %v provisioning for codebase %v",
	"codebase_branch_registration":        "Branch %v for codebase %v registration",
	"accept_codebase_branch_registration": "Accept branch %v for codebase %v registration",
}

func ConvertToCodebaseBranch(k8sObject edpv1alpha1.CodebaseBranch) (*CodebaseBranch, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object application branch object should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCodebaseBranchActionLog(spec.BranchName, spec.CodebaseName, k8sObject.Status)

	branch := CodebaseBranch{
		Name:       spec.BranchName,
		Tenant:     strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		AppName:    spec.CodebaseName,
		FromCommit: spec.FromCommit,
		Status:     k8sObject.Status.Value,
		ActionLog:  *actionLog,
	}

	return &branch, nil
}

func convertCodebaseBranchActionLog(brName, cbName string, status edpv1alpha1.CodebaseBranchStatus) *ActionLog {
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
		ActionMessage:   fmt.Sprintf(codebaseBranchActionMessageMap[status.Action], brName, cbName),
	}
}
