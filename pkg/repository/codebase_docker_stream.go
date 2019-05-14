package repository

import (
	"database/sql"
	"fmt"
)

const (
	CreateCodebaseDockerStreamQuery = "insert into \"%v\".codebase_docker_stream(codebase_id, oc_image_stream_name)" +
		" values($1, $2) returning id;"
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
