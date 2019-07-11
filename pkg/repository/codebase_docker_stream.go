package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	CreateCodebaseDockerStreamQuery = "insert into \"%v\".codebase_docker_stream(codebase_id, oc_image_stream_name)" +
		" values($1, $2) returning id;"
	GetDockerStreamsByPipelineNameQuery = "select cds.id, cds.codebase_id, cb.name " +
		"from \"%[1]v\".codebase_docker_stream as cds " +
		"left join \"%[1]v\".codebase cb on cds.codebase_id = cb.id " +
		"left join \"%[1]v\".codebase_branch cbb on cds.id = cbb.output_codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_pipeline_codebase_branch cpcb on cbb.id = cpcb.codebase_branch_id " +
		"left join \"%[1]v\".cd_pipeline pipe on cpcb.cd_pipeline_id = pipe.id " +
		"where pipe.name = $1;"
	GetDockerStreamsByPipelineNameAndStageOrderQuery = "select cds.id, cds.codebase_id, cb.name " +
		"from \"%[1]v\".codebase_docker_stream as cds " +
		"left join \"%[1]v\".codebase cb on cds.codebase_id = cb.id " +
		"left join \"%[1]v\".stage_codebase_docker_stream scds on cds.id = scds.output_codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_stage cs on scds.cd_stage_id = cs.id " +
		"left join \"%[1]v\".cd_pipeline pipe on cs.cd_pipeline_id = pipe.id " +
		"where pipe.name = $1 and cs.\"order\" = $2;"
	CreateStageCodebaseDockerStreamQuery = "insert into \"%v\".stage_codebase_docker_stream " +
		"values($1, $2, $3);"
	RemoveStageCodebaseDockerStream = "delete " +
		"	from \"%v\".stage_codebase_docker_stream scds " +
		"where scds.cd_stage_id = $1 returning scds.output_codebase_docker_stream_id id;"
	RemoveCodebaseDockerStream = "delete from \"%v\".codebase_docker_stream cds where cds.id = $1 ;"
)

func CreateCodebaseDockerStream(txn sql.Tx, schemaName string, codebaseId int, ocImageStreamName string) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(CreateCodebaseDockerStreamQuery, schemaName))
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(codebaseId, ocImageStreamName).Scan(&id)
	return
}

func GetDockerStreamsByPipelineName(txn sql.Tx, schemaName string, cdPipelineName string) ([]model.CodebaseDockerStreamReadDTO, error) {
	query := fmt.Sprintf(GetDockerStreamsByPipelineNameQuery, schemaName)
	stmt, err := txn.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(cdPipelineName)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return getDockerStreamsFromRows(rows)
}

func GetDockerStreamsByPipelineNameAndStageOrder(txn sql.Tx, schemaName string, cdPipelineName string, order int) ([]model.CodebaseDockerStreamReadDTO, error) {
	query := fmt.Sprintf(GetDockerStreamsByPipelineNameAndStageOrderQuery, schemaName)
	stmt, err := txn.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(cdPipelineName, order)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	return getDockerStreamsFromRows(rows)
}

func CreateStageCodebaseDockerStream(txn sql.Tx, schemaName string, stageId int, inputStreamId int, outputStreamId int) error {
	query := fmt.Sprintf(CreateStageCodebaseDockerStreamQuery, schemaName)
	stmt, err := txn.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(stageId, inputStreamId, outputStreamId)

	return err
}

func getDockerStreamsFromRows(rows *sql.Rows) ([]model.CodebaseDockerStreamReadDTO, error) {
	var result []model.CodebaseDockerStreamReadDTO

	for rows.Next() {
		dto := model.CodebaseDockerStreamReadDTO{}
		err := rows.Scan(&dto.CodebaseDockerStreamId, &dto.CodebaseId, &dto.CodebaseName)
		if err != nil {
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

func DeleteStageCodebaseDockerStream(txn sql.Tx, stageId int, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(RemoveStageCodebaseDockerStream, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(stageId).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func DeleteCodebaseDockerStream(txn sql.Tx, id int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(RemoveCodebaseDockerStream, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}
