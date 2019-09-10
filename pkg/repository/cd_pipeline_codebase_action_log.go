package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

const (
	insertEventActionLog = "insert into \"%v\".action_log" +
		"(detailed_message, username, updated_at, action, action_message, result) " +
		"VALUES($1, $2, $3, $4, $5, $6) returning id;"
	insertCDPipelineActionLog = "insert into \"%v\".cd_pipeline_action_log(cd_pipeline_id, action_log_id) values ($1, $2);"
)

func CreateCDPipelineActionLog(txn sql.Tx, pipelineId int, actionLogId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertCDPipelineActionLog, schemaName))
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
	stmt, err := txn.Prepare(fmt.Sprintf(insertEventActionLog, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(actionLog.DetailedMessage, actionLog.Username, actionLog.UpdatedAt,
		actionLog.Action, actionLog.ActionMessage, actionLog.Result).Scan(&id)

	return &id, err
}
