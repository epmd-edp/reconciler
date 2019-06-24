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
	UpdateStageStatusQuery                = "update \"%v\".cd_stage set status = $1 where id = $2;"
	CreateStageActionLogQuery             = "insert into \"%v\".cd_stage_action_log(cd_stage_id, action_log_id) values ($1, $2);"
	GetStageIdByPipelineNameAndOrderQuery = "select stage.id from \"%v\".cd_stage stage " +
		"left join \"%v\".cd_pipeline pipe on stage.cd_pipeline_id = pipe.id " +
		"where pipe.name = $1 and stage.\"order\" = $2;"
	InsertCDStageCodebase = "insert into \"%v\".cd_stage_codebase(cd_stage_id, codebase_id) VALUES ($1, $2);"
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
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(stageId, actionLogId)
	return err
}

func GetStageIdByPipelineNameAndOrder(txn sql.Tx, schemaName string, cdPipelineName string, order int) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(GetStageIdByPipelineNameAndOrderQuery, schemaName, schemaName))

	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(cdPipelineName, order).Scan(&id)
	if err != nil {
		return checkNoRows(err)
	}
	return
}

func checkNoRows(err error) (*int, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}

func CreateCDStageCodebase(txn sql.Tx, cdStageId int, autotestId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCDStageCodebase, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cdStageId, autotestId)
	if err != nil {
		return err
	}
	return nil
}
