package thirdpartyservice

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/epmd-edp/reconciler/v2/pkg/model/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateServiceMethod_ShouldBeExecutedSuccessfully(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
		SchemaName:  "fake-schema",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`insert into "fake-schema".third_party_service`).ExpectExec().
		WithArgs(s.Name, s.Description, s.Version, s.Url, s.Icon).
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	err = CreateService(*tx, s)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)
}

func TestCreateServiceMethod_ShouldReturnErrorDuringPrepareStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`insert into "fake-schema".third_party_service`).ExpectExec().
		WithArgs(s.Name, s.Description, s.Version, s.Url, s.Icon).
		WillReturnResult(sqlmock.NewResult(1, 1))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	assert.Error(t, CreateService(*tx, s))
}

func TestGetService_ShouldBeExecutedSuccessfully(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
		SchemaName:  "fake-schema",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs(s.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	id, err := GetService(*tx, "fake-name", "fake-schema")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}

	assert.NoError(t, err)
	assert.Equal(t, 1, *id)
}

func TestGetService_ShouldReturnErrorDuringPrepareStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs(s.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = GetService(*tx, "fake-name", "")

	assert.Error(t, err)
}

func TestGetService_ShouldReturnErrorDuringScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
		SchemaName:  "fake-schema",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs(s.Name).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(nil))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	_, err = GetService(*tx, "fake-name", "fake-schema")

	assert.Error(t, err)
}

func TestGetService_ShouldReturnNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	s := service.ServiceDto{
		Name:        "fake-name",
		Version:     "fake-version",
		Description: "fake-desc",
		Url:         "fake-url",
		Icon:        "fake-icon",
		SchemaName:  "fake-schema",
	}

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs(s.Name).
		WillReturnError(sql.ErrNoRows)

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	id, err := GetService(*tx, "fake-name", "fake-schema")

	assert.NoError(t, err)
	assert.Nil(t, id)
}
