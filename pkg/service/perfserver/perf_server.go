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

func (s PerfServerService) PutPerfServer(server perfserver.PerfServer, schema string) error {
	log.Info("start creating PerfServer record in DB", "name", server.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	id, err := perfServerRepo.SelectPerfServer(*txn, server.Name, schema)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while fetching PerfServer %v", server.Name)
	}

	if err := tryToPutPerfServer(txn, id, server, schema); err != nil {
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
