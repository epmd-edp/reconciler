package repository

import (
	"business-app-reconciler-controller/pkg/db"
	"business-app-reconciler-controller/pkg/model"
	"fmt"
	"testing"
	"time"
)

func TestAppRepo_AddApplication(t *testing.T) {
	dbConn, err := db.InitConnection()

	if err != nil {
		t.Fatal(err)
	}

	appRepo := AppRepo{
		DB: *dbConn,
	}

	app := model.AppEntity{
		Namespace: "fightclub",
		Name:      "fc-ui",

		Lang:      "javascript",
		Framework: "react",
		BuildTool: "npm",
		Strategy:  "create",

		Status: &model.Status{
			Available:       true,
			LastTimeUpdated: time.Now(),
			Status:          "added",
		},
	}

	err = appRepo.AddApplication(app)

	if err != nil {
		fmt.Println(err)
	}
}
