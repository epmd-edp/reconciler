package service

import (
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"testing"
	"time"
)

func TestPutStage(t *testing.T) {
	service := StageService{
		DB: db.Instance,
	}

	stage := model.Stage{
		Name:            "stage",
		CdPipelineName:  "team-a",
		Description:     "Description for stage",
		TriggerType:     "manual",
		QualityGate:     "manual",
		JenkinsStepName: "manual",
		Tenant:          "py-test",
		Order:           3,
		ActionLog: model.ActionLog{
			Event:           "created",
			DetailedMessage: "",
			Username:        "",
			UpdatedAt:       time.Now(),
		},
		Status: "inactive",
	}

	err := service.PutStage(stage)

	if err != nil {
		t.Fatal(err)
	}
}
