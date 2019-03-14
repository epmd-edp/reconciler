package repository

import (
	"business-app-reconciler-controller/pkg/model"
	"database/sql"
)

const (
	InsertBE     = "insert into business_entity(name, tenant, be_type) values($1,$2,$3) returning id"
	InsertProp   = "insert into be_properties(be_id, property, value) values ($1,$2,$3)"
	InsertStatus = "insert into be_status(be_id, status, user_name, message, last_time_update, available) values($1,$2,$3,$4,$5,$6)"
	SelectStatus = "select status_id from statuses_list where lower(status_name) = lower($1)"
	SelectBE     = "select id from business_entity where be_type=$1 AND name=$2 AND tenant=$3 AND delition=$4"
)

func GetBEId(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(SelectBE)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int

	err = stmt.QueryRow(be.Type, be.Name, be.Tenant, 0).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func CreateBE(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	stmt, err := txn.Prepare(InsertBE)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(be.Name, be.Tenant, be.Type).Scan(&id)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func CreateProps(txn sql.Tx, beId int, props map[string]string) error {
	stmt, err := txn.Prepare(InsertProp)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range props {
		_, err = stmt.Exec(beId, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetStatusId(txn sql.Tx, status string) (*int, error) {
	stmt, err := txn.Prepare(SelectStatus)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var id int
	err = stmt.QueryRow(status).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &id, err
}

func CreateStatus(txn sql.Tx, beId int, status model.Status) error {
	stmt, err := txn.Prepare(InsertStatus)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(beId, status.Id, status.Username, status.Message, status.LastTimeUpdated, status.Available)

	return err
}
