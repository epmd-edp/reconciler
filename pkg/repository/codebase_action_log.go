package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	InsertCodebaseStatus = "insert into \"%v\".codebase_action_log(codebase_id, action_log_id) " +
		"values($1, $2);"
	InsertActionLog = "insert into \"%v\".action_log(event, detailed_message, username, updated_at) " +
		"VALUES($1, $2, $3, $4) returning id;"
	CheckDuplicateActionLog = "select codebase.id" +
		" from \"%v\".codebase" +
		"	left join \"%v\".codebase_action_log cal on codebase.id = cal.codebase_id" +
		" left join \"%v\".action_log al on cal.action_log_id = al.id" +
		" WHERE name = $1" +
		"  AND event = $2" +
		"  AND updated_at = $3" +
		" order by updated_at desc" +
		" limit 1;"
)

func CreateCodebaseAction(txn sql.Tx, codebaseId int, codebaseActionId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseStatus, schemaName))
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
	err = stmt.QueryRow(actionLog.Event, "", actionLog.Username, actionLog.UpdatedAt).Scan(&id)

	return &id, err
}

func GetLastIdActionLog(txn sql.Tx, be model.BusinessEntity, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(CheckDuplicateActionLog, schemaName, schemaName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(be.Name, be.ActionLog.Event, be.ActionLog.UpdatedAt).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}
