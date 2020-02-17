package stage

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
	deleteCDStage = "delete " +
		"	from \"%[1]v\".cd_stage cs using \"%[1]v\".cd_pipeline cp " +
		"where cs.cd_pipeline_id = cp.id " +
		"and cp.name = $1 " +
		"  and cs.name = $2 ;"
	deleteCodebaseDockerStream   = "delete from \"%v\".codebase_docker_stream where id = $1 ;"
	selectCodebaseDockerStreamId = "select cds.id " +
		"	from \"%[1]v\".codebase_docker_stream cds " +
		"left join \"%[1]v\".stage_codebase_docker_stream scds on cds.id = scds.output_codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_stage cs on scds.cd_stage_id = cs.id " +
		"left join \"%[1]v\".cd_pipeline cp on cs.cd_pipeline_id = cp.id " +
		"where cp.name = $1 " +
		"  and cs.name = $2 ;"
	deleteCodebaseDockerStreamIds = "delete " +
		"	from \"%[1]v\".codebase_docker_stream cds " +
		"where cds.id in (select cds.id " +
		"from \"%[1]v\".codebase_docker_stream cds " +
		"left join \"%[1]v\".stage_codebase_docker_stream scds on cds.id = scds.output_codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_stage cs on scds.cd_stage_id = cs.id " +
		"left join \"%[1]v\".cd_pipeline cp on cs.cd_pipeline_id = cp.id " +
		"where cp.name = $1 );"
)

func CreateStage(txn sql.Tx, stage stage.Stage, cdPipelineId int) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertStage, stage.Tenant))
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

func DeleteCDStage(txn sql.Tx, pipeName, stageName, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCDStage, schema), pipeName, stageName); err != nil {
		return err
	}
	return nil
}

func SelectCodebaseDockerStreamId(txn sql.Tx, pipeName, stageName, schema string) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectCodebaseDockerStreamId, schema))
	if err != nil {
		return
	}
	defer stmt.Close()

	if err = stmt.QueryRow(pipeName, stageName).Scan(&id); err != nil {
		return checkNoRows(err)
	}
	return
}

func DeleteCodebaseDockerStream(txn sql.Tx, id int, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebaseDockerStream, schema), id); err != nil {
		return err
	}
	return nil
}

func DeleteCodebaseDockerStreams(txn sql.Tx, pipeName, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebaseDockerStreamIds, schema), pipeName); err != nil {
		return err
	}
	return nil
}
