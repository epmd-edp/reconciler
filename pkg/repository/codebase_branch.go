package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	SelectCodebaseBranch = "select cb.id as codebase_branch_id from \"%v\".codebase_branch cb" +
		" left join \"%v\".codebase c on cb.codebase_id = c.id where cb.name=$1 and c.name=$2;"
	InsertCodebaseBranch = "insert into \"%v\".codebase_branch(name, codebase_id, from_commit, output_codebase_docker_stream_id)" +
		" values ($1, $2, $3, $4) returning id;"
	SelectCodebaseBranchesId = "select cb.id as cb_id " +
		"from \"%v\".codebase_branch cb " +
		"		left join \"%v\".codebase c on cb.codebase_id = c.id " +
		"where (cb.name = $1 and c.name = $2);"
)

func GetCodebaseBranchId(txn sql.Tx, codebaseName string, codebaseBranchName string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectCodebaseBranch, schemaName, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(codebaseBranchName, codebaseName).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateCodebaseBranch(txn sql.Tx, name string, beId int, fromCommit string, schemaName string, streamId int) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseBranch, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name, beId, fromCommit, streamId).Scan(&id)
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
