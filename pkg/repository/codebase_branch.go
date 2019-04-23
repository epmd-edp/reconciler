package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
	"fmt"
)

const (
	SelectCodebaseBranch = "select cb.id as codebase_branch_id from \"%v\".codebase_branch cb" +
		" left join \"%v\".codebase c on cb.codebase_id = c.id where cb.name=$1 and c.name=$2 and c.tenant_name=$3;"
	SelectCodebaseTenantName = "select tenant_name from \"%v\".codebase where name=$1;"
	InsertCodebaseBranch     = "insert into \"%v\".codebase_branch(name, codebase_id, from_commit) values ($1, $2, $3) returning id;"
	SelectCodebaseBranchesId = "select cb.id as cb_id " +
		"from \"%v\".codebase_branch cb " +
		"		left join \"%v\".codebase c on cb.codebase_id = c.id " +
		"where (cb.name = $1 and c.name = $2);"
)

func GetCodebaseBranchId(txn sql.Tx, codebaseName string, codebaseBranchName string, tenant string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseBranch, schemaName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(codebaseBranchName, codebaseName, tenant).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func GetCodebaseTenantName(txn sql.Tx, appName string, schemaName string) (*string, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseTenantName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var tenantName string

	err = stmt.QueryRow(appName).Scan(&tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &tenantName, nil
}

func CreateCodebaseBranch(txn sql.Tx, name string, beId int, fromCommit string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseBranch, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name, beId, fromCommit).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func GetCodebaseBranchesId(txn sql.Tx, appBranch model.ApplicationBranchDTO, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseBranchesId, schemaName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var branchId int
	err = stmt.QueryRow(appBranch.BranchName, appBranch.AppName).Scan(&branchId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &branchId, nil
}
