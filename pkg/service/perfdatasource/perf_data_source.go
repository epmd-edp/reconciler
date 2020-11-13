package perfdatasource

import (
	"database/sql"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/perfdatasource"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

type PerfDataSourceService struct {
	DB *sql.DB
}

var log = logf.Log.WithName("perf-data-source-service")

func (s PerfDataSourceService) perfDataSourceExists(dsType, tenant string) (bool, error) {
	log.Info("checking for existence record data source", "type", dsType)
	txn, err := s.DB.Begin()
	if err != nil {
		return false, err
	}

	exists, err := perfdatasource.PerfDataSourceExists(*txn, strings.ToUpper(dsType), tenant)
	if err != nil {
		return false, err
	}

	if err := txn.Commit(); err != nil {
		return false, err
	}

	return exists, nil
}

func (s PerfDataSourceService) InsertPerfDataSources(perf *codebase.Perf, tenant string) error {
	if perf == nil {
		return nil
	}
	log.Info("start inserting data source records")

	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for _, ds := range perf.DataSources {
		exists, err := s.perfDataSourceExists(ds, tenant)
		if err != nil {
			return err
		}

		if exists {
			log.Info("data source already exists. skip creating", "type", ds)
			continue
		}

		if err := perfdatasource.InsertPerfDataSource(*txn, strings.ToUpper(ds), tenant); err != nil {
			_ = txn.Rollback()
			return err
		}
		log.Info("data sources has been added to table", "type", ds)
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s PerfDataSourceService) RemoveCodebaseDataSource(codebase, dataSource, tenant string) error {
	rLog := log.WithValues("codebase", codebase, "data source", dataSource)
	rLog.Info("removing codebase_perf_data_source record")
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	if err := perfdatasource.RemoveCodebaseDataSource(*txn, codebase, dataSource, tenant); err != nil {
		_ = txn.Rollback()
		return err
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	rLog.Info("codebase_perf_data_source record has been removed")
	return nil
}
