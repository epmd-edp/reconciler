package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"time"
)

const (
	InsertCodebaseStatus = "insert into codebase_action_log(codebase_id, action_log_id) " +
		"values($1, $2);"
	InsertActionLog = "insert into action_log(event, detailed_message, username, updated_at) " +
		"VALUES($1, $2, $3, $4) returning id;"
	CheckDuplicateActionLog = "select codebase.id" +
		" from codebase" +
		"	left join codebase_action_log cal on codebase.id = cal.codebase_id" +
		" left join action_log al on cal.action_log_id = al.id" +
		" WHERE name = $1" +
		"  AND event = $2" +
		"  AND updated_at = $3" +
		" order by updated_at desc" +
		" limit 1;"
)

func CreateCodebaseAction(txn sql.Tx, codebaseId int, codebaseActionId int) error {
	stmt, err := txn.Prepare(InsertCodebaseStatus)
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

func CreateActionLog(txn sql.Tx, actionLog model.ActionLog) (*int, error) {
	stmt, err := txn.Prepare(InsertActionLog)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(actionLog.Event, "", actionLog.Username, time.Unix(actionLog.UpdatedAt, 0).Format("2006-01-02 15:04:05")).Scan(&id)

	return &id, err
}

func GetLastIdActionLog(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(CheckDuplicateActionLog)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(be.Name, be.ActionLog.Event, time.Unix(be.ActionLog.UpdatedAt, 0).Format("2006-01-02 15:04:05")).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}
