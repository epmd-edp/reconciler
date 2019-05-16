package model

import (
	"errors"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"strings"
)

type CodebaseBranch struct {
	Name       string
	Tenant     string
	AppName    string
	FromCommit string
	ActionLog  ActionLog
}

func ConvertToCodebaseBranch(k8sObject edpv1alpha1.ApplicationBranch) (*CodebaseBranch, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object application branch object should not be nil")
	}
	spec := k8sObject.Spec

	actionLog := convertCodebaseBranchActionLog(k8sObject.Status)

	branch := CodebaseBranch{
		Name:       k8sObject.Spec.BranchName,
		Tenant:     strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
		AppName:    spec.AppName,
		FromCommit: spec.FromCommit,
		ActionLog:  *actionLog,
	}

	return &branch, nil
}

func convertCodebaseBranchActionLog(status edpv1alpha1.ApplicationBranchStatus) *ActionLog {
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
