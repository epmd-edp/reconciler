package jenkins_slave

import (
	"database/sql"
	"fmt"
)

const (
	SelectJenkinsSlaveSql = "select id from \"%v\".jenkins_slave where name = $1;"
	InsertJenkinsSlaveSql = "insert into \"%v\".jenkins_slave(name) values ($1)"
)

func SelectJenkinsSlave(txn sql.Tx, name, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectJenkinsSlaveSql, tenant))
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

func CreateJenkinsSlave(txn sql.Tx, name string, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertJenkinsSlaveSql, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name)

	return err
}
