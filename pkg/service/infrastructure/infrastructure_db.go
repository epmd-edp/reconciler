package infrastructure

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("infrastructure-db-service")

type InfrastructureDbService struct {
	DB *sql.DB
}

//DoesSchemaExist checks if schema exists in DB.
func (s InfrastructureDbService) DoesSchemaExist(schema string) (bool, error) {
	log.Info("Start check schema ...")

	txn, err := s.DB.Begin()
	if err != nil {
		return false, err
	}

	isSchemaExist, err := repository.DoesSchemaExist(*txn, schema)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("an error has occurred while checking existing of %v schema", schema))
	}

	err = txn.Commit()
	if err != nil {
		return false, err
	}

	return isSchemaExist, nil
}
