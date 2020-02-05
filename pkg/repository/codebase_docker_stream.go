package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

const (
	CreateCodebaseDockerStreamQuery = "insert into \"%v\".codebase_docker_stream(codebase_branch_id, oc_image_stream_name)" +
		" values($1, $2) returning id;"
	GetDockerStreamsByPipelineNameQuery = "select cds.id, c.id codebase_id, c.name " +
		"from \"%[1]v\".codebase_docker_stream cds " +
		"left join \"%[1]v\".codebase_branch cb on cds.codebase_branch_id = cb.id " +
		"left join \"%[1]v\".codebase c on cb.codebase_id = c.id " +
		"left join \"%[1]v\".cd_pipeline_docker_stream cpds on cds.id = cpds.codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_pipeline cp on cpds.cd_pipeline_id = cp.id " +
		"where cp.name = $1;"
	GetDockerStreamsByPipelineNameAndStageOrderQuery = "select cds.id, c.id codebase_id, c.name " +
		"	from \"%[1]v\".codebase_docker_stream cds " +
		"left join \"%[1]v\".codebase_branch cb on cds.codebase_branch_id = cb.id " +
		"left join \"%[1]v\".codebase c on cb.codebase_id = c.id " +
		"left join \"%[1]v\".stage_codebase_docker_stream scds on cds.id = scds.output_codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_stage cs on scds.cd_stage_id = cs.id " +
		"left join \"%[1]v\".cd_pipeline pipe on cs.cd_pipeline_id = pipe.id " +
		"where pipe.name = $1 and cs.\"order\" = $2;"
	CreateStageCodebaseDockerStreamQuery = "insert into \"%v\".stage_codebase_docker_stream " +
		"values($1, $2, $3);"
	RemoveStageCodebaseDockerStream = "delete " +
		"	from \"%v\".stage_codebase_docker_stream scds " +
		"where scds.cd_stage_id = $1 returning scds.output_codebase_docker_stream_id id;"
	SelectSourceInputStream = "select cds.id " +
		"	from \"%[1]v\".codebase_docker_stream cds " +
		"left join \"%[1]v\".codebase_branch cb on cds.codebase_branch_id = cb.id " +
		"left join \"%[1]v\".codebase c on cb.codebase_id = c.id " +
		"left join \"%[1]v\".cd_pipeline_docker_stream cpds on cds.id = cpds.codebase_docker_stream_id " +
		"left join \"%[1]v\".cd_pipeline cp on cpds.cd_pipeline_id = cp.id " +
		"where cp.name = $1 and c.name = $2 ;"
	SelectCodebaseDockerStreamId       = "select id from \"%[1]v\".codebase_docker_stream cds where cds.oc_image_stream_name=$1 ;"
	UpdateCodebaseDockerStreamBranchId = "update \"%v\".codebase_docker_stream set codebase_branch_id = $1 where id = $2 ;"
	SelectCodebaseDockerStreamBranchId = "select cds.codebase_branch_id from \"%v\".codebase_docker_stream cds where cds.id = $1;"
)

func CreateCodebaseDockerStream(txn sql.Tx, schemaName string, branchId *int, ocImageStreamName string) (id *int, err error) {
	stmt, err := txn.Prepare(fmt.Sprintf(CreateCodebaseDockerStreamQuery, schemaName))
	if err != nil {
		return
	}
	defer stmt.Close()

	err = stmt.QueryRow(branchId, ocImageStreamName).Scan(&id)
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

func DeleteStageCodebaseDockerStream(txn sql.Tx, stageId int, schemaName string) ([]int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(RemoveStageCodebaseDockerStream, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(stageId)
	defer rows.Close()
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return getOutputStreamIds(rows)
}

func getOutputStreamIds(rows *sql.Rows) ([]int, error) {
	var result []int

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	err := rows.Err()
	if err != nil {
		return nil, err
	}
	return result, err
}

func GetSourceInputStream(txn sql.Tx, cdPipelineName, codebaseName, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectSourceInputStream, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cdPipelineName, codebaseName).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}

func GetCodebaseDockerStreamId(txn sql.Tx, dockerStream, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseDockerStreamId, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(dockerStream).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}

func UpdateBranchIdCodebaseDockerStream(txn sql.Tx, dockerStreamId int, branchId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateCodebaseDockerStreamBranchId, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(branchId, dockerStreamId)
	return err
}

func GetCodebaseDockerStreamBranchId(txn sql.Tx, dockerStreamId int, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseDockerStreamBranchId, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(dockerStreamId).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}

func checkNoRows(err error) (*int, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}
