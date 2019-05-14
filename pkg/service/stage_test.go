package service

import (
	"reconciler/pkg/db"
	"reconciler/pkg/model"
	"testing"
	"time"
)

func TestPutStage(t *testing.T) {
	db, _ := db.InitConnection()
	service := StageService{
		DB: *db,
	}

	stage := model.Stage{
		Name:            "qa",
		CdPipelineName:  "team-a",
		Description:     "Description for stage",
		TriggerType:     "manual",
		QualityGate:     "manual",
		JenkinsStepName: "manual",
		Tenant:          "tarianyk-test",
		Order:           1,
		ActionLog: model.ActionLog{
			Event:           "created",
			DetailedMessage: "",
			Username:        "",
			UpdatedAt:       time.Now().Unix(),
		},
		Status: "inactive",
	}

	err := service.PutStage(stage)

	if err != nil {
		t.Fatal(err)
	}
}
