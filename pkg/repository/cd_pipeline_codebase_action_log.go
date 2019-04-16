package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"time"
)

const (
	CheckDuplicateCDPipelineActionLog = "select cdp.id " +
		"	from cd_pipeline as cdp " +
		"left join cd_pipeline_action_log cdpal on cdp.id = cdpal.cd_pipeline_id " +
		"left join action_log al on cdpal.action_log_id = al.id " +
		"where cdp.name = $1 " +
		"  and al.event = $2 " +
		"  and al.updated_at = $3 " +
		"order by al.updated_at desc " +
		"limit 1;"
	InsertEventActionLog = "insert into action_log(event, detailed_message, username, updated_at) " +
		"VALUES($1, $2, $3, $4) returning id;"

	InsertCDPipelineActionLog = "insert into cd_pipeline_action_log(cd_pipeline_id, action_log_id) values ($1, $2);"
)

func CreateCDPipelineActionLog(txn sql.Tx, pipelineId int, actionLogId int) error {
	stmt, err := txn.Prepare(InsertCDPipelineActionLog)
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

func CreateEventActionLog(txn sql.Tx, actionLog model.ActionLog) (*int, error) {
	stmt, err := txn.Prepare(InsertEventActionLog)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(actionLog.Event, "", actionLog.Username, time.Unix(actionLog.UpdatedAt, 0).Format("2006-01-02 15:04:05")).Scan(&id)

	return &id, err
}

func CheckCDPipelineActionLogDuplicate(txn sql.Tx, cdPipeline model.CDPipeline) (bool, error) {
	stmt, err := txn.Prepare(CheckDuplicateCDPipelineActionLog)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cdPipeline.Name, cdPipeline.ActionLog.Event, time.Unix(cdPipeline.ActionLog.UpdatedAt, 0).Format("2006-01-02 15:04:05")).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
