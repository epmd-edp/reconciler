package edp_component

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
)

const (
	InsertEDPComponentSql = "insert into \"%v\".edp_component(type, url, icon) values ($1, $2, $3);"
	SelectEDPComponentSql = "select id from \"%v\".edp_component where type = $1;"
)

func CreateEDPComponent(txn sql.Tx, component model.EDPComponent, tenant string) error {
	stmt, err := txn.Prepare(fmt.Sprintf(InsertEDPComponentSql, tenant))
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(component.Type, component.Url, component.Icon)

	return err
}

func SelectEDPComponent(txn sql.Tx, componentType, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectEDPComponentSql, tenant))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(componentType).Scan(&id)
	if err != nil {
		return checkRows(err)
	}
	return &id, err
}

func checkRows(err error) (*int, error) {
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return nil, err
}
