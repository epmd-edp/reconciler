package codebasebranch

import (
	"database/sql"
	"fmt"
)

const (
	SelectCodebaseBranch = "select cb.id as codebase_branch_id from \"%v\".codebase_branch cb" +
		" left join \"%v\".codebase c on cb.codebase_id = c.id where cb.name=$1 and c.name=$2;"
	InsertCodebaseBranch = "insert into \"%v\".codebase_branch(name, codebase_id, from_commit, output_codebase_docker_stream_id, status, version, build_number, last_success_build, release)" +
		" values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id;"
	UpdateCodebaseBranchStatus = "update \"%v\".codebase_branch set status = $1 where id = $2;"
	UpdateCodebaseBranchValues = "update \"%v\".codebase_branch set version = $1, build_number = $2, last_success_build = $3 where id = $4;"
	deleteCodebaseBranch       = "delete from \"%[1]v\".codebase_branch cb " +
		"	using \"%[1]v\".codebase as c " +
		"where c.name = $1 " +
		"  and cb.name = $2 ;"
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

func CreateCodebaseBranch(txn sql.Tx, name string, beId int, fromCommit string,
	schemaName string, streamId *int, status string, version *string, buildNumber *string, lastSuccessBuild *string, release bool) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseBranch, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name, beId, fromCommit, streamId, status, version, buildNumber, lastSuccessBuild, release).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func UpdateStatusByCodebaseBranchId(txn sql.Tx, branchId int, status string, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateCodebaseBranchStatus, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, branchId)
	return err
}

func UpdateCodebaseBranch(txn sql.Tx, branchId int, version *string, build *string, lastSuccess *string, schemaName string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateCodebaseBranchValues, schemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(version, build, lastSuccess, branchId)
	return err
}

func Delete(txn sql.Tx, codebase, branch, schema string) error {
	if _, err := txn.Exec(fmt.Sprintf(deleteCodebaseBranch, schema), codebase, branch); err != nil {
		return err
	}
	return nil
}
