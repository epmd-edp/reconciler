package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
	"strings"
)

const (
	InsertCodebase = "insert into \"%v\".codebase(name, type, language, framework, build_tool, strategy, repository_url, route_site," +
		" route_path, database_kind, database_version, database_capacity, database_storage, status)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14) returning id;"
	SelectCodebase = "select id from \"%v\".codebase where type=$1 AND name=$2;"
	StatusActive   = "active"
)

func GetCodebaseId(txn sql.Tx, beType model.BEType, name string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebase, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(beType, name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateCodebase(txn sql.Tx, cb model.BusinessEntity, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebase, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cb.Name, cb.Type, strings.ToLower(cb.Language), strings.ToLower(cb.Framework),
		strings.ToLower(cb.BuildTool), strings.ToLower(cb.Strategy), cb.RepositoryUrl, cb.RouteSite, cb.RoutePath,
		cb.DatabaseKind, cb.DatabaseVersion, cb.DatabaseCapacity, cb.DatabaseStorage, StatusActive).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}
