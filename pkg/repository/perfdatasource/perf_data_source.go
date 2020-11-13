package perfdatasource

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	perfDataSourceExists         = "select exists(select 1 from \"%v\".perf_data_sources where type=$1);"
	insertPerfDataSource         = "insert into \"%v\".perf_data_sources(type) values ($1) returning id;"
	selectPerfDataSource         = "select id from \"%v\".perf_data_sources where type = $1;"
	deleteCodebasePerfDataSource = "delete from \"%[1]v\".codebase_perf_data_sources cpds where cpds.data_source_id =" +
		" (select pds.id from \"%[1]v\".perf_data_sources pds where pds.type = $1) " +
		"and cpds.codebase_id= (select c.id from \"%[1]v\".codebase c where c.name= $2);"
)

func PerfDataSourceExists(txn sql.Tx, dsType, tenant string) (bool, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(perfDataSourceExists, tenant))
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists bool
	if err = stmt.QueryRow(dsType).Scan(&exists); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists, err
}

func InsertPerfDataSource(txn sql.Tx, dsType, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertPerfDataSource, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(dsType)
	return err
}

func GetDataSourceId(txn sql.Tx, dsType, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectPerfDataSource, tenant))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	if err = stmt.QueryRow(dsType).Scan(&id); err != nil {
		return nil, err
	}
	return &id, err
}

func RemoveCodebaseDataSource(txn sql.Tx, codebase, dataSource, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebasePerfDataSource, schema), strings.ToUpper(dataSource), codebase); err != nil {
		return err
	}
	return nil
}
