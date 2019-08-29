package repository

import (
	"database/sql"
)

const (
	CheckSchema = "select exists(select 1 from pg_namespace where nspname = $1);"
)

func DoesSchemaExist(txn sql.Tx, schema string) (bool, error) {
	var exists bool
	err := txn.QueryRow(CheckSchema, schema).Scan(&exists)
	if err != nil {
		return false, nil
	}
	return exists, nil
}
