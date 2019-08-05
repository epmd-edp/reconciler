package repository

import (
	"database/sql"
	"fmt"
)

const (
	SelectCodebaseBranch = "select cb.id as codebase_branch_id from \"%v\".codebase_branch cb" +
		" left join \"%v\".codebase c on cb.codebase_id = c.id where cb.name=$1 and c.name=$2;"
	InsertCodebaseBranch = "insert into \"%v\".codebase_branch(name, codebase_id, from_commit, output_codebase_docker_stream_id, status)" +
		" values ($1, $2, $3, $4, $5) returning id;"
	UpdateCodebaseBranchStatus = "update \"%v\".codebase_branch set status = $1 where id = $2;"
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
	schemaName string, streamId *int, status string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertCodebaseBranch, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name, beId, fromCommit, streamId, status).Scan(&id)
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
