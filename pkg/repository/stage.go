package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	"log"
)

const (
	InsertStage = "insert into \"%v\".cd_stage(name, cd_pipeline_id, description, trigger_type," +
		" \"order\", status, codebase_branch_id) VALUES ($1, $2, $3, $4, $5, $6, $7) returning id;"
	SelectStageId = "select st.id as st_id from \"%v\".cd_stage st " +
		"left join \"%v\".cd_pipeline pl on st.cd_pipeline_id = pl.id " +
		"where (st.name = $1 and pl.name = $2);"
	UpdateStageStatusQuery                = "update \"%v\".cd_stage set status = $1 where id = $2;"
	CreateStageActionLogQuery             = "insert into \"%v\".cd_stage_action_log(cd_stage_id, action_log_id) values ($1, $2);"
	GetStageIdByPipelineNameAndOrderQuery = "select stage.id from \"%v\".cd_stage stage " +
		"left join \"%v\".cd_pipeline pipe on stage.cd_pipeline_id = pipe.id " +
		"where pipe.name = $1 and stage.\"order\" = $2;"
	GetStagesIdByCDPipelineName = "select cs.id, cs.name, cs.status, cs.trigger_type, cs.description, cs.\"order\" " +
		"	from \"%v\".cd_pipeline cp " +
		"right join \"%v\".cd_stage cs on cp.id = cs.cd_pipeline_id " +
		"where cp.name = $1 ;"
	InsertQualityGate = "insert into \"%v\".quality_gate_stage(quality_gate, step_name, cd_stage_id, codebase_id, codebase_branch_id) " +
		" values ($1, $2, $3, $4, $5) returning id; "
	SelectCodebaseAndBranchIds = "select c.id codebase_id, cb.id codebase_branch_id " +
		"	from \"%v\".codebase c " +
		"left join \"%v\".codebase_branch cb on c.id = cb.codebase_id " +
		"where c.type = 'autotests' " +
		"  and c.name = $1 " +
		"  and cb.name = $2 ; "
)

func CreateStage(txn sql.Tx, schemaName string, stage stage.Stage, cdPipelineId int) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertStage, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(stage.Name, cdPipelineId, stage.Description,
		stage.TriggerType, stage.Order, stage.Status,
		getLibraryBranchIdOrNil(stage.Source)).Scan(&id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func getLibraryBranchIdOrNil(source stage.Source) *int {
	if source.Type == "default" {
		return nil
	}
	return source.Library.BranchId
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

func GetStages(txn sql.Tx, pipelineName string, schemaName string) ([]stage.Stage, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(GetStagesIdByCDPipelineName, schemaName, schemaName))

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(pipelineName)
	defer rows.Close()
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return getStage(rows)
}

func getStage(rows *sql.Rows) ([]stage.Stage, error) {
	var result []stage.Stage

	for rows.Next() {
		dto := stage.Stage{}
		err := rows.Scan(&dto.Id, &dto.Name, &dto.Status, &dto.TriggerType, &dto.Description, &dto.Order)
		if err != nil {
			log.Printf("Error during parsing: %v", err)
			return nil, err
		}
		result = append(result, dto)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return result, err
}

func CreateQualityGate(txn sql.Tx, qualityGateType string, jenkinsStepName string, cdStageId int, codebaseId *int, codebaseBranchId *int, schemaName string) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertQualityGate, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(qualityGateType, jenkinsStepName, cdStageId, codebaseId, codebaseBranchId).Scan(&id)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func GetCodebaseAndBranchIds(txn sql.Tx, autotestName, branchName, schemaName string) (*model.CodebaseBranchIdDTO, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseAndBranchIds, schemaName, schemaName))

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	dto := model.CodebaseBranchIdDTO{}
	err = stmt.QueryRow(autotestName, branchName).Scan(&dto.CodebaseId, &dto.BranchId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &dto, nil
}
