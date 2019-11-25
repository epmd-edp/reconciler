package repository

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/thirdpartyservice"
)

const (
	InsertService = "insert into \"%v\".third_party_service(name, description, version) values ($1, $2, $3) returning id;"
	SelectService = "select id from \"%v\".third_party_service where name=$1;"
)

func CreateThirdPartyService(txn sql.Tx, service thirdpartyservice.ThirdPartyService, schemaName string) (*int, error) {
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

func GetThirdPartyService(txn sql.Tx, serviceName string, schemaName string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectService, schemaName))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(serviceName).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}
