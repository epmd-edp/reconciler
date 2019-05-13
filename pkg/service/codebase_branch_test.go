package service

import (
	"reconciler/pkg/db"
	"reconciler/pkg/model"
	"testing"
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
