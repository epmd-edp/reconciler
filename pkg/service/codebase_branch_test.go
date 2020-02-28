package service

import (
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebasebranch"
	cbs "github.com/epmd-edp/reconciler/v2/pkg/service/codebasebranch"
	"testing"
	"time"
)

func TestCodebaseBranchService_PutCodebaseBranchIfApplicationDoesNotExist(t *testing.T) {
	beService := cbs.CodebaseBranchService{
		DB: db.Instance,
	}

	branch := codebasebranch.CodebaseBranch{
		AppName: "non-exist",
		Name:    "some",
	}

	err := beService.PutCodebaseBranch(branch)

	if err != nil {
		t.Fatal("Error should be occurred if application for name does not exist")
	}
}

func TestCreateBranch(t *testing.T) {
	service := cbs.CodebaseBranchService{
		DB: db.Instance,
	}

	branch := codebasebranch.CodebaseBranch{
		Name:       "master",
		Tenant:     "py-test",
		AppName:    "petclinic-be",
		FromCommit: "qwe123",
		ActionLog: model.ActionLog{
			Event:           "created",
			DetailedMessage: "",
			Username:        "",
			UpdatedAt:       time.Now(),
		},
	}
	err := service.PutCodebaseBranch(branch)

	if err != nil {
		t.Fatal(err)
	}
}
