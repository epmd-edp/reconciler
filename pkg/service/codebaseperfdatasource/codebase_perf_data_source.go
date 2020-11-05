package codebaseperfdatasource

import (
	"database/sql"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/codebaseperfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/perfdatasource"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

type CodebasePerfDataSourceService struct {
	DB *sql.DB
}

var log = logf.Log.WithName("codebase-perf-data-source-service")

func (s CodebasePerfDataSourceService) InsertCodebasePerfDataSources(codebaseId int, perf *codebase.Perf, tenant string) error {
	if perf == nil {
		return nil
	}
	log.Info("insert CodebasePerfDataSource record", "codebase id", codebaseId)

	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for _, ds := range perf.DataSources {
		id, err := perfdatasource.GetDataSourceId(*txn, strings.ToUpper(ds), tenant)
		if err != nil {
			return err
		}

		exists, err := s.codebasePerfDataSourceExists(codebaseId, *id, tenant)
		if err != nil {
			return err
		}

		if exists {
			continue
		}

		if err := codebaseperfdatasource.InsertCodebasePerfDataSource(*txn, codebaseId, *id, tenant); err != nil {
			_ = txn.Rollback()
			return err
		}
		log.Info("CodebasePerfDataSource has been added to table",
			"codebase id", codebaseId, "data source id", *id)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	return nil
}

func (s CodebasePerfDataSourceService) codebasePerfDataSourceExists(codebaseId, dsId int, tenant string) (bool, error) {
	log.Info("checking for existence CodebasePerfDataSource record",
		"codebase id", codebaseId, "data source id", dsId)
	txn, err := s.DB.Begin()
	if err != nil {
		return false, err
	}

	exists, err := codebaseperfdatasource.CodebasePerfDataSourceExists(*txn, codebaseId, dsId, tenant)
	if err != nil {
		return false, err
	}

	if err := txn.Commit(); err != nil {
		return false, err
	}

	return exists, nil
}
