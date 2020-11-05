package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"strings"
)

const (
	insertCodebase = "insert into \"%v\".codebase(name, type, language, framework, build_tool, strategy, repository_url, route_site," +
		" route_path, database_kind, database_version, database_capacity, database_storage, status, test_report_framework, description," +
		" git_server_id, git_project_path, jenkins_slave_id, job_provisioning_id, deployment_script, project_status, versioning_type," +
		" start_versioning_from, jira_server_id, commit_message_pattern, ticket_name_pattern, ci_tool, perf_server_id)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22," +
		" $23, $24, $25, $26, $27, $28, $29) returning id;"
	selectCodebase       = "select id from \"%v\".codebase where name=$1;"
	selectCodebaseType   = "select type from \"%v\".codebase where id=$1;"
	updateCodebaseStatus = "update \"%v\".codebase set status = $1 where id = $2;"
	selectApplication    = "select id from \"%v\".codebase where name=$1 and type='application';"
	deleteCodebase       = "delete from \"%v\".codebase where name=$1;"
	updateCodebase       = "update \"%v\".codebase set commit_message_pattern = $1, ticket_name_pattern = $2 where name = $3;"
)

const (
	projectCreatedStatus = "created"
	projectPushedStatus  = "pushed"
)

func GetCodebaseId(txn sql.Tx, name string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectCodebase, schemaName))
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

func CreateCodebase(txn sql.Tx, c codebase.Codebase, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(insertCodebase, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(c.Name, c.Type, strings.ToLower(c.Language), c.Framework,
		strings.ToLower(c.BuildTool), strings.ToLower(c.Strategy), c.RepositoryUrl, c.RouteSite, c.RoutePath,
		c.DatabaseKind, c.DatabaseVersion, c.DatabaseCapacity, c.DatabaseStorage, c.Status,
		c.TestReportFramework, c.Description,
		getIntOrNil(c.GitServerId), getStringOrNil(c.GitUrlPath), getIntOrNil(c.JenkinsSlaveId),
		getIntOrNil(c.JobProvisioningId), c.DeploymentScript, getStatus(c.Strategy), c.VersioningType,
		c.StartVersioningFrom, getIntOrNil(c.JiraServerId), getStringOrNil(c.CommitMessagePattern),
		getStringOrNil(c.TicketNamePattern), c.CiTool, getPerfIdOrNil(c.Perf)).Scan(&id)
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

func getPerfIdOrNil(perf *codebase.Perf) interface{} {
	if perf == nil {
		return nil
	}
	return getIntOrNil(perf.Id)
}
func GetCodebaseTypeById(txn sql.Tx, cbId int, schemaName string) (*string, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectCodebaseType, schemaName))
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
	stmt, err := txn.Prepare(fmt.Sprintf(updateCodebaseStatus, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, cbId)
	return err
}

func GetApplicationId(txn sql.Tx, name string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectApplication, schemaName))
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
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebase, schema), name); err != nil {
		return err
	}
	return nil
}

func Update(txn sql.Tx, c codebase.Codebase, schema string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(updateCodebase, schema))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(getStringOrNil(c.CommitMessagePattern), getStringOrNil(c.TicketNamePattern), c.Name)
	return err
}
