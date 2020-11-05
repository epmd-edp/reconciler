package perfserver

import (
	"database/sql"
	"github.com/epmd-edp/reconciler/v2/pkg/model/perfserver"
	perfServerRepo "github.com/epmd-edp/reconciler/v2/pkg/repository/perfserver"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("perf-server-service")

type PerfServerService struct {
	DB *sql.DB
}

func (s PerfServerService) PutPerfServer(server perfserver.PerfServer, tenant string) error {
	log.Info("start creating PerfServer record in DB", "name", server.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	id, err := perfServerRepo.SelectPerfServer(*txn, server.Name, tenant)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while fetching PerfServer %v", server.Name)
	}

	if err := tryToPutPerfServer(txn, id, server, tenant); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while putting PerfServer %v", server.Name)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("PerfServer has been created/updated", "name", server.Name)
	return nil
}

func tryToPutPerfServer(txn *sql.Tx, id *int, server perfserver.PerfServer, schema string) error {
	if id != nil {
		log.Info("start updating PerfServer", "name", server.Name)
		return perfServerRepo.UpdatePerfServer(*txn, id, server.Available, schema)
	}
	log.Info("start creating PerfServer", "name", server.Name)
	return perfServerRepo.CreatePerfServer(*txn, server.Name, server.Available, schema)
}

func (s PerfServerService) GetPerfServerId(name, tenant string) (*int, error) {
	log.Info("getting perf server id", "name", name)
	txn, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}

	id, err := perfServerRepo.SelectPerfServer(*txn, name, tenant)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has occurred while fetching PerfServer %v", name)
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return id, nil
}
