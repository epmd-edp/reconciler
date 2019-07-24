package repository

import (
	"database/sql"
	"fmt"
)

const (
	InsertApplicationsToPromote = "insert into \"%v\".applications_to_promote(cd_pipeline_id, codebase_id) values ($1, $2);"
	DeleteApplicationsToPromote = "delete from \"%v\".applications_to_promote where cd_pipeline_id = $1 ;"
)

func CreateApplicationsToPromote(txn sql.Tx, cdPipelineId int, codebaseId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertApplicationsToPromote, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cdPipelineId, codebaseId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveApplicationsToPromote(txn sql.Tx, cdPipelineId int, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(DeleteApplicationsToPromote, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(cdPipelineId)
	if err != nil {
		return err
	}
	return nil
}
