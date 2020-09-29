package thirdpartyservice

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/service"
)

const (
	insertService = "insert into \"%v\".third_party_service(name, description, version, url, icon) values ($1, $2, $3, $4, $5);"
	selectService = "select id from \"%v\".third_party_service where name=$1;"
)

func CreateService(txn sql.Tx, service service.ServiceDto) error {
	stmt, err := txn.Prepare(fmt.Sprintf(insertService, service.SchemaName))
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err = stmt.Exec(service.Name, service.Description, service.Version, service.Url, service.Icon); err != nil {
		return err
	}

	return nil
}

func GetService(txn sql.Tx, name, schema string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(selectService, schema))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(name).Scan(&id)
	if err != nil {
		_, err = checkNoRows(err)
		return nil, err
	}

	return &id, nil
}

func checkNoRows(err error) (*int, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}
