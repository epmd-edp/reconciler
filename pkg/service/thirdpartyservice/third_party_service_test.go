package thirdpartyservice

import (
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/epmd-edp/reconciler/v2/pkg/model/service"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPutService_ServiceShouldExist(t *testing.T) {
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
	mock.ExpectCommit()

	tps := ThirdPartyService{
		DB: db,
	}

	err = tps.PutService(s)
	assert.NoError(t, err)
}

func TestPutService_GetServiceShouldReturnError(t *testing.T) {
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
		WillReturnError(errors.New("fake"))
	mock.ExpectRollback()

	tps := ThirdPartyService{
		DB: db,
	}

	err = tps.PutService(s)
	assert.Error(t, err)
}

func TestPutService_RollbackShouldBeFailed(t *testing.T) {
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
		WillReturnError(errors.New("fake"))

	tps := ThirdPartyService{
		DB: db,
	}

	err = tps.PutService(s)
	assert.Error(t, err)
}

func TestPutService_ShouldCreateServiceSuccessfully(t *testing.T) {
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
	mock.ExpectPrepare(`insert into "fake-schema".third_party_service`).ExpectExec().
		WithArgs(s.Name, s.Description, s.Version, s.Url, s.Icon).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	tps := ThirdPartyService{
		DB: db,
	}

	err = tps.PutService(s)
	assert.NoError(t, err)
}

func TestGetServicesId_ShouldReturnTwoServices(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs("service1").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs("service2").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	tps := ThirdPartyService{
		DB: db,
	}

	ids, err := tps.GetServicesId(tx, []string{"service1", "service2"}, "fake-schema")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ids))
}

func TestGetServicesId_ShouldReturnError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectPrepare(`select id from "fake-schema".third_party_service`).ExpectQuery().
		WithArgs("service1").
		WillReturnError(errors.New("fake"))

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	tps := ThirdPartyService{
		DB: db,
	}

	_, err = tps.GetServicesId(tx, []string{"service1", "service2"}, "fake-schema")
	assert.Error(t, err)
}
