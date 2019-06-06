package repository

import (
	"database/sql"
	"fmt"
	"reconciler/pkg/model"
)

const (
	InsertService = "insert into \"%v\".service(name, description, version) values ($1, $2, $3) returning id;"
	SelectService = "select id from \"%v\".service where name=$1;"
)

func CreateThirdPartyService(txn sql.Tx, service model.ThirdPartyService, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertService, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(service.Name, service.Description, service.Version).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func GetThirdPartyService(txn sql.Tx, service model.ThirdPartyService, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectService, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(service.Name).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}