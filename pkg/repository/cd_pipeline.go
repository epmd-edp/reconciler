package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/cdpipeline"
)

const (
	InsertCDPipeline                  = "insert into \"%v\".cd_pipeline(name, status) VALUES ($1, $2) returning id, name, status;"
	SelectCDPipeline                  = "select * from \"%v\".cd_pipeline cdp where cdp.name = $1 ;"
	UpdateCDPipelineStatusQuery       = "update \"%v\".cd_pipeline set status = $1 where id = $2 ;"
	InsertCDPipelineThirdPartyService = "insert into \"%v\".cd_pipeline_third_party_service(cd_pipeline_id, third_party_service_id) values ($1, $2) ;"
	InsertCDPipelineDockerStream      = "insert into \"%v\".cd_pipeline_docker_stream(cd_pipeline_id, codebase_docker_stream_id) VALUES ($1, $2);"
	DeleteAllDockerStreams            = "delete from \"%v\".cd_pipeline_docker_stream cpds  where cpds.cd_pipeline_id = $1 ;"
	deleteCDPipeline                  = "delete from \"%v\".cd_pipeline where name = $1 ;"
)

func CreateCDPipeline(txn sql.Tx, cdPipeline cdpipeline.CDPipeline, status string, schemaName string) (*model.CDPipelineDTO, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCDPipeline, schemaName))
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

func GetCDPipeline(txn sql.Tx, cdPipelineName string, schemaName string) (*model.CDPipelineDTO, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCDPipeline, schemaName))
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

func UpdateCDPipelineStatus(txn sql.Tx, pipelineId int, cdPipelineStatus string, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateCDPipelineStatusQuery, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cdPipelineStatus, pipelineId)
	return err
}

func CreateCDPipelineThirdPartyService(txn sql.Tx, pipelineId int, serviceId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCDPipelineThirdPartyService, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pipelineId, serviceId)
	return err
}

func CreateCDPipelineDockerStream(txn sql.Tx, pipelineId int, dockerStreamId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCDPipelineDockerStream, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pipelineId, dockerStreamId)
	return err
}

func DeleteCDPipelineDockerStreams(txn sql.Tx, pipelineId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(DeleteAllDockerStreams, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pipelineId)
	return err
}

func DeleteCDPipeline(txn sql.Tx, pipeName, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCDPipeline, schema), pipeName); err != nil {
		return err
	}
	return nil
}
