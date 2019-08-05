package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	InsertEventActionLog = "insert into \"%v\".action_log(event, detailed_message, username, updated_at) " +
		"VALUES($1, $2, $3, $4) returning id;"

	InsertCDPipelineActionLog = "insert into \"%v\".cd_pipeline_action_log(cd_pipeline_id, action_log_id) values ($1, $2);"
)

func CreateCDPipelineActionLog(txn sql.Tx, pipelineId int, actionLogId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCDPipelineActionLog, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pipelineId, actionLogId)
	if err != nil {
		return err
	}
	return nil
}

func CreateEventActionLog(txn sql.Tx, actionLog model.ActionLog, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertEventActionLog, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(actionLog.Event, "", actionLog.Username, actionLog.UpdatedAt).Scan(&id)

	return &id, err
}
