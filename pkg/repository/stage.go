package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	InsertStage = "insert into \"%v\".cd_stage(name, cd_pipeline_id, description, trigger_type, quality_gate," +
		" jenkins_step_name, \"order\", status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id;"
	SelectStageId = "select st.id as st_id from \"%v\".cd_stage st " +
		"left join \"%v\".cd_pipeline pl on st.cd_pipeline_id = pl.id " +
		"where (st.name = $1 and pl.name = $2);"
	UpdateStageStatusQuery    = "update \"%v\".cd_stage set status = $1 where id = $2;"
	CreateStageActionLogQuery = "insert into \"%v\".cd_stage_action_log(cd_stage_id, action_log_id) values ($1, $2);"
)

func CreateStage(txn sql.Tx, schemaName string, stage model.Stage, cdPipelineId int) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertStage, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(stage.Name, cdPipelineId, stage.Description, stage.TriggerType, stage.QualityGate,
		stage.JenkinsStepName, stage.Order, stage.Status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func GetStageId(txn sql.Tx, schemaName string, name string, cdPipelineName string) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectStageId, schemaName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(name, cdPipelineName).Scan(&id)
	if err != nil {
		return checkNoRows(err)
	}
	return id, nil
}

func UpdateStageStatus(txn sql.Tx, schemaName string, id int, status string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateStageStatusQuery, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(status, id)
	return err
}

func CreateStageActionLog(txn sql.Tx, schemaName string, stageId int, actionLogId int) error {
	stmt, err := txn.Prepare(fmt.Sprintf(CreateStageActionLogQuery, schemaName))

	if err != nil {
		return nil
	}
	defer stmt.Close()
	_, err = stmt.Exec(stageId, actionLogId)
	return err
}

func checkNoRows(err error) (*int, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}
