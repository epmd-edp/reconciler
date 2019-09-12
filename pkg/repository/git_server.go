package repository

import (
	"database/sql"
	"fmt"
)

const (
	InsertGitServerSql = "insert into \"%v\".git_server(name, hostname, available) values ($1, $2, $3) returning id;"
	UpdateGitServerSql = "update \"%v\".git_server set available = $1 where id = $2;"
	SelectGitServerSql = "select id from \"%v\".git_server where name = $1;"
)

func CreateGitServer(txn sql.Tx, name string, hostname string, available bool, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertGitServerSql, tenant))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name, hostname, available).Scan(&id)
	return &id, err
}

func UpdateGitServer(txn sql.Tx, id *int, available bool, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(UpdateGitServerSql, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(available, id)
	return err
}

func SelectGitServer(txn sql.Tx, name, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectGitServerSql, tenant))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, err
}
