package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

const (
	InsertActionLog = "insert into \"%v\".action_log(detailed_message, username, updated_at, action, action_message, result) " +
		"VALUES($1, $2, $3, $4, $5, $6) returning id;"

	InsertCodebaseActionLog = "insert into \"%v\".codebase_action_log(codebase_id, action_log_id) " +
		"values($1, $2);"
)

func CreateCodebaseAction(txn sql.Tx, codebaseId int, codebaseActionId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseActionLog, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(codebaseId, codebaseActionId)
	if err != nil {
		return err
	}
	return nil
}

func CreateActionLog(txn sql.Tx, actionLog model.ActionLog, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertActionLog, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(actionLog.DetailedMessage, actionLog.Username, actionLog.UpdatedAt,
		actionLog.Action, actionLog.ActionMessage, actionLog.Result).Scan(&id)

	return &id, err
}
