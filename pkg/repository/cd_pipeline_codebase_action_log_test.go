package repository

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"testing"
	"time"
)

func TestCreateEventActionLog(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	log := model.ActionLog{
		Id:              1,
		Username:        "fake-username",
		UpdatedAt:       time.Now(),
		DetailedMessage: "fake-detailed-message",
		ActionMessage:   "fake-action-message",
		Result:          "success",
		Action:          "setup_initial_structure",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`insert into "fake-schema".action_log`).ExpectQuery().
		WithArgs(log.DetailedMessage, log.Username, log.UpdatedAt, log.Action, log.ActionMessage, log.Result).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(log.Id))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	id, err := CreateEventActionLog(*tx, log, "fake-schema")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	if *id != log.Id {
		t.Fatal(fmt.Sprintf("id is incorrect %v, but expected %v", *id, log.Id))
	}
}
