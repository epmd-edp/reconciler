package jira_server

import (
	"database/sql"
	"fmt"
)

const (
	insertGitServer  = "insert into \"%v\".jira_server(name, available) values ($1, $2) returning id;"
	updateGitServer  = "update \"%v\".jira_server set available = $1 where id = $2;"
	selectJiraServer = "select id from \"%v\".jira_server where name = $1;"
)

func CreateJiraServer(txn sql.Tx, name string, available bool, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertGitServer, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, available)
	return err
}

func UpdateJiraServer(txn sql.Tx, id *int, available bool, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(updateGitServer, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(available, id)
	return err
}

func SelectJiraServer(txn sql.Tx, name, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectJiraServer, tenant))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	if err = stmt.QueryRow(name).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, err
}
