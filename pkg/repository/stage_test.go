package repository

import (
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"testing"
	"time"
)

func rTestInsertStage(t *testing.T) {
	database, err := db.InitConnection()
	if err != nil {
		t.Fatal(err)
	}

	txn, err := database.Begin()
	stage := model.Stage{
		Name:            "sit",
		CdPipelineName:  "test",
		Description:     "Description for stage",
		TriggerType:     "manual",
		QualityGate:     "manual",
		JenkinsStepName: "manual",
		Order:           1,
		ActionLog: model.ActionLog{
			Event:           "created",
			DetailedMessage: "",
			Username:        "",
			UpdatedAt:       time.Now().Unix(),
		},
		Status: "active",
	}

	id, err := CreateStage(*txn, "tarianyk-test", stage, 1)

	if err != nil {
		txn.Rollback()
		t.Fatal(err)
	}

	txn.Commit()

	fmt.Printf("id of created stage: %v", id)
}

func TestGetStageId(t *testing.T) {
	database, err := db.InitConnection()
	if err != nil {
		t.Fatal(err)
	}

	txn, err := database.Begin()

	id, err := GetStageId(*txn, "tarianyk-test", "sit-1", "team-a")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(id)
}
