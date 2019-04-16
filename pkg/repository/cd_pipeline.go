package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
)

const (
	InsertCDPipeline               = "insert into cd_pipeline(name, status) VALUES ($1, $2) returning id, name, status;"
	InsertCDPipelineCodebaseBranch = "insert into cd_pipeline_codebase_branch(cd_pipeline_id, codebase_branch_id) VALUES ($1, $2);"
	SelectCDPipeline               = "select * from cd_pipeline cdp where cdp.name = $1 ;"
	UpdateCDPipelineStatusQuery    = "update cd_pipeline set status = $1 where id = $2 ;"
)

func CreateCDPipeline(txn sql.Tx, cdPipeline model.CDPipeline, status string) (*model.CDPipelineDTO, error) {
	stmt, err := txn.Prepare(InsertCDPipeline)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var cdPipelineDto model.CDPipelineDTO
	err = stmt.QueryRow(cdPipeline.Name, status).Scan(&cdPipelineDto.Id, &cdPipelineDto.Name, &cdPipelineDto.Status)
	if err != nil {
		return nil, err
	}
	return &cdPipelineDto, nil
}

func CreateCDPipelineCodebaseBranch(txn sql.Tx, pipelineId int, codebaseBranchId int) error {
	stmt, err := txn.Prepare(InsertCDPipelineCodebaseBranch)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pipelineId, codebaseBranchId)
	if err != nil {
		return err
	}
	return nil
}

func GetCDPipeline(txn sql.Tx, cdPipelineName string) (*model.CDPipelineDTO, error) {
	stmt, err := txn.Prepare(SelectCDPipeline)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var cdPipeline model.CDPipelineDTO
	err = stmt.QueryRow(cdPipelineName).Scan(&cdPipeline.Id, &cdPipeline.Name, &cdPipeline.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cdPipeline, nil
}

func UpdateCDPipelineStatus(txn sql.Tx, pipelineId int, cdPipelineStatus string) error {
	stmt, err := txn.Prepare(UpdateCDPipelineStatusQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cdPipelineStatus, pipelineId)
	if err != nil {
		return err
	}
	return nil
}
