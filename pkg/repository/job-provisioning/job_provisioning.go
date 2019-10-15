package job_provisioning

import (
	"database/sql"
	"fmt"
)

const (
	SelectJobProvisioningSql = "select id from \"%v\".job_provisioning where name = $1;"
)

func SelectJobProvisioning(txn sql.Tx, name, tenant string) (*int, error) {
	stmt, err := txn.Prepare(fmt.Sprintf(SelectJobProvisioningSql, tenant))
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
