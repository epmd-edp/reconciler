package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"time"
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
