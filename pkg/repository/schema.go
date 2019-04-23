package repository

import (
	"database/sql"
)

const (
	SelectDbSchema = "SELECT schema_name FROM information_schema.schemata WHERE schema_name=$1;"
)

func GetSchema(txn sql.Tx, tenantName string) (*string, error) {
	defaultSchema := "public"

	stmt, err := txn.Prepare(SelectDbSchema)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var schemaName string

	err = stmt.QueryRow(tenantName).Scan(&schemaName)
	if err != nil {
		if err == sql.ErrNoRows {
			return &defaultSchema, nil
		}
		return nil, err
	}

	return &schemaName, nil
}
