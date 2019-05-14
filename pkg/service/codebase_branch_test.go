package service

import (
	"reconciler/pkg/db"
	"reconciler/pkg/model"
	"testing"
	"time"
)

func TestCodebaseBranchService_PutCodebaseBranchIfApplicationDoesNotExist(t *testing.T) {
	dbConn, _ := db.InitConnection()
	beService := CodebaseBranchService{
		DB: *dbConn,
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
	dbConn, err := db.InitConnection()

	if err != nil {
		t.Fatal(err)
	}
	service := CodebaseBranchService{
		DB: *dbConn,
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
			UpdatedAt:       time.Now().Unix(),
		},
	}
	err = service.PutCodebaseBranch(branch)

	if err != nil {
		t.Fatal(err)
	}
}
