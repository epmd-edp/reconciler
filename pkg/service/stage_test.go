package service

import (
	"github.com/epmd-edp/reconciler/v2/pkg/db"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	stage2 "github.com/epmd-edp/reconciler/v2/pkg/service/stage"
	"testing"
	"time"
)

func TestPutStage(t *testing.T) {
	service := stage2.StageService{
		DB: db.Instance,
	}

	stage := stage.Stage{
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
