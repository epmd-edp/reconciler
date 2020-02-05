package stage

import (
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	sm "github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	"testing"
	"time"
)

func TestInsertStage(t *testing.T) {
	txn, err := db.Instance.Begin()
	if err != nil {
		t.Fatal(err)
	}

	s := sm.Stage{
		Name:           "sit",
		CdPipelineName: "test",
		Description:    "Description for stage",
		TriggerType:    "manual",
		Order:          1,
		ActionLog: model.ActionLog{
			Event:           "created",
			DetailedMessage: "",
			Username:        "",
			UpdatedAt:       time.Now(),
		},
		Status: "active",
	}

	id, err := CreateStage(*txn, s, 1)
	if err != nil {
		_ = txn.Rollback()
		t.Fatal(err)
	}

	if err := txn.Commit(); err != nil {
		t.Fatal(err)
	}

	fmt.Printf("id of created stage: %v", id)
}

func TestGetStageId(t *testing.T) {
	txn, err := db.Instance.Begin()
	if err != nil {
		t.Fatal(err)
	}

	id, err := GetStageId(*txn, "tarianyk-test", "sit-1", "team-a")

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(id)
}
