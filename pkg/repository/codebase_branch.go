package repository

import (
	"database/sql"
)

const (
	SelectCodebaseBranch = "select cb.id as codebase_branch_id from codebase_branch cb" +
		" left join codebase c on cb.codebase_id = c.id where cb.name=$1 and c.name=$2 and c.tenant_name=$3;"
	SelectCodebaseTenantName = "select tenant_name from codebase where name=$1;"
	InsertCodebaseBranch = "insert into codebase_branch(name, codebase_id, from_commit) values ($1, $2, $3) returning id;"
)

func GetCodebaseBranchId(txn sql.Tx, codebaseName string, codebaseBranchName string, tenant string) (*int, error) {
	stmt, err := txn.Prepare(SelectCodebaseBranch)
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

func GetCodebaseTenantName(txn sql.Tx, appName string) (*string, error) {
	stmt, err := txn.Prepare(SelectCodebaseTenantName)
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

func CreateCodebaseBranch(txn sql.Tx, name string, beId int, fromCommit string) (*int, error) {
	stmt, err := txn.Prepare(InsertCodebaseBranch)
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