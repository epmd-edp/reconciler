package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"strings"
)

const (
	InsertBE = "insert into codebase(name, tenant_name, type, language, framework, build_tool, strategy, repository_url, route_site," +
		" route_path, database_kind, database_version, database_capacity, database_storage, status)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15) returning id;"
	SelectBE        = "select id from codebase where type=$1 AND name=$2 AND tenant_name=$3;"
	InsertActionLog = "insert into action_log(event, detailed_message, username, updated_at) " +
						"VALUES($1, $2, $3, $4) returning id;"
	InsertCodebaseStatus = "insert into codebase_action_log(codebase_id, action_log_id) " +
							"values($1, $2);"
	CheckDuplicateActionLog = "select codebase.id" +
		" from codebase" +
		"	left join codebase_action_log cal on codebase.id = cal.codebase_id" +
		" left join action_log al on cal.action_log_id = al.id" +
		" WHERE name = $1" +
		"  AND event = $2" +
		"  AND updated_at = $3" +
		" order by updated_at desc" +
		" limit 1;"
	StatusActive = "active"
)

func GetBEId(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(SelectBE)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(be.Type, be.Name, be.Tenant).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateBE(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(InsertBE)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(be.Name, be.Tenant, be.Type, strings.ToLower(be.Language), strings.ToLower(be.Framework),
		strings.ToLower(be.BuildTool), strings.ToLower(be.Strategy), be.RepositoryUrl, be.RouteSite, be.RoutePath,
		be.DatabaseKind, be.DatabaseVersion, be.DatabaseCapacity, be.DatabaseStorage, StatusActive).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}
