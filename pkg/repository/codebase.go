package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"strings"
)

const (
	InsertCodebase = "insert into codebase(name, tenant_name, type, language, framework, build_tool, strategy, repository_url, route_site," +
		" route_path, database_kind, database_version, database_capacity, database_storage, status)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) returning id;"
	SelectCodebase = "select id from codebase where type=$1 AND name=$2 AND tenant_name=$3;"
	StatusActive   = "active"
)

func GetCodebaseId(txn sql.Tx, beType model.BEType, name string, tenant string) (*int, error) {
	stmt, err := txn.Prepare(SelectCodebase)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(beType, name, tenant).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateCodebase(txn sql.Tx, cb model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(InsertCodebase)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(cb.Name, cb.Tenant, cb.Type, strings.ToLower(cb.Language), strings.ToLower(cb.Framework),
		strings.ToLower(cb.BuildTool), strings.ToLower(cb.Strategy), cb.RepositoryUrl, cb.RouteSite, cb.RoutePath,
		cb.DatabaseKind, cb.DatabaseVersion, cb.DatabaseCapacity, cb.DatabaseStorage, StatusActive).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}
