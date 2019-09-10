package service

import (
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"testing"
	"time"
)

func TestCodebaseBranchService_PutCodebaseBranchIfApplicationDoesNotExist(t *testing.T) {
	beService := CodebaseBranchService{
		DB: db.Instance,
	}

	branch := model.CodebaseBranch{
		AppName: "non-exist",
		Name:    "some",
	}

	err := beService.PutCodebaseBranch(branch)

	if err != nil {
		t.Fatal("Error should be occurred if application for name does not exist")
	}
}

func TestCreateBranch(t *testing.T) {
	service := CodebaseBranchService{
		DB: db.Instance,
	}

	branch := model.CodebaseBranch{
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
