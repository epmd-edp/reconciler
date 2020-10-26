package perfserver

import (
	"database/sql"
	"fmt"
)

const (
	selectJiraServer = "select id from \"%v\".perf_server where name = $1;"
	updateGitServer  = "update \"%v\".perf_server set available = $1 where id = $2;"
	insertGitServer  = "insert into \"%v\".perf_server(name, available) values ($1, $2) returning id;"
)

func SelectPerfServer(txn sql.Tx, name, tenant string) (*int, error) {
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

func UpdatePerfServer(txn sql.Tx, id *int, available bool, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(updateGitServer, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(available, id)
	return err
}

func CreatePerfServer(txn sql.Tx, name string, available bool, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertGitServer, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, available)
	return err
}
