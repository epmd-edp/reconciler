package codebaseperfdatasource

import (
	"database/sql"
	"fmt"
)

const (
	insertPerfDataSource         = "insert into \"%v\".codebase_perf_data_sources(codebase_id, data_source_id) values ($1, $2);"
	codebasePerfDataSourceExists = "select exists(select 1 from \"%v\".codebase_perf_data_sources where codebase_id=$1 and data_source_id=$2);"
	deleteCodebasePerfDataSource = "delete from \"%v\".codebase_perf_data_sources where codebase_id=$1;"
)

func InsertCodebasePerfDataSource(txn sql.Tx, codebaseId, dsId int, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertPerfDataSource, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(codebaseId, dsId)
	return err
}

func CodebasePerfDataSourceExists(txn sql.Tx, codebaseId, dsId int, tenant string) (bool, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(codebasePerfDataSourceExists, tenant))
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists bool
	if err = stmt.QueryRow(codebaseId, dsId).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists, err
}

func DeleteCodebasePerfDataSourceRecord(txn sql.Tx, codebaseId int, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebasePerfDataSource, schema), codebaseId); err != nil {
		return err
	}
	return nil
}
