package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"strings"
)

const (
	InsertCodebase = "insert into \"%v\".codebase(name, type, language, framework, build_tool, strategy, repository_url, route_site," +
		" route_path, database_kind, database_version, database_capacity, database_storage, status, test_report_framework, description," +
		" git_server_id, git_project_path, jenkins_slave_id, job_provisioning_id, deployment_script, project_status, versioning_type, start_versioning_from)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24) returning id;"
	SelectCodebase       = "select id from \"%v\".codebase where name=$1;"
	SelectCodebaseType   = "select type from \"%v\".codebase where id=$1;"
	UpdateCodebaseStatus = "update \"%v\".codebase set status = $1 where id = $2;"
	SelectApplication    = "select id from \"%v\".codebase where name=$1 and type='application';"
	DeleteCodebase       = "delete from \"%v\".codebase where name=$1;"
)

const (
	projectCreatedStatus = "created"
	projectPushedStatus  = "pushed"
)

func GetCodebaseId(txn sql.Tx, name string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebase, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateCodebase(txn sql.Tx, cb codebase.Codebase, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebase, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cb.Name, cb.Type, strings.ToLower(cb.Language), cb.Framework,
		strings.ToLower(cb.BuildTool), strings.ToLower(cb.Strategy), cb.RepositoryUrl, cb.RouteSite, cb.RoutePath,
		cb.DatabaseKind, cb.DatabaseVersion, cb.DatabaseCapacity, cb.DatabaseStorage, cb.Status,
		cb.TestReportFramework, cb.Description,
		getIntOrNil(cb.GitServerId), getStringOrNil(cb.GitUrlPath), getIntOrNil(cb.JenkinsSlaveId),
		getIntOrNil(cb.JobProvisioningId), cb.DeploymentScript, getStatus(cb.Strategy), cb.VersioningType, cb.StartVersioningFrom).Scan(&id)

	if err != nil {
		return nil, err
	}

	return &id, nil
}

func getStatus(strategy string) string {
	if strategy == "import" {
		return projectPushedStatus
	}
	return projectCreatedStatus
}

func getIntOrNil(value *int) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func getStringOrNil(value *string) interface{} {
	if value == nil {
		return nil
	}
	return *value
}

func GetCodebaseTypeById(txn sql.Tx, cbId int, schemaName string) (*string, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseType, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var cbType string
	err = stmt.QueryRow(cbId).Scan(&cbType)
	if err != nil {
		return nil, err
	}

	return &cbType, nil
}

func UpdateStatusByCodebaseId(txn sql.Tx, cbId int, status string, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateCodebaseStatus, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, cbId)
	return err
}

func GetApplicationId(txn sql.Tx, name string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectApplication, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func Delete(txn sql.Tx, name, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(DeleteCodebase, schema), name); err != nil {
		return err
	}
	return nil
}
