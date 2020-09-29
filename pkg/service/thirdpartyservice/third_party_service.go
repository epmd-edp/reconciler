package thirdpartyservice

import (
	"database/sql"
	"github.com/epmd-edp/reconciler/v2/pkg/model/service"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/thirdpartyservice"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type ThirdPartyService struct {
	DB *sql.DB
}

var log = logf.Log.WithName("third-party-service-layer")

func (s ThirdPartyService) PutService(service service.ServiceDto) error {
	log.Info("start creating ThirdPartyService row in DB", "name", service.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.Wrapf(err, "couldn't open transaction to create record for %v ThirdPartyService", service.Name)
	}

	if err := tryToCreateService(txn, service); err != nil {
		if err := txn.Rollback(); err != nil {
			return errors.Wrapf(err, "couldn't finish rollback while creating %v ThirdPartyService record in DB", service.Name)
		}
		return errors.Wrapf(err, "couldn't create %v ThirdPartyService record in DB", service.Name)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrapf(err, "couldn't commit changes to db for %v ThirdPartyService", service.Name)
	}
	log.Info("ThirdPartyService has been created", "name", service.Name)
	return nil
}

func tryToCreateService(txn *sql.Tx, service service.ServiceDto) error {
	id, err := thirdpartyservice.GetService(*txn, service.Name, service.SchemaName)
	if err != nil {
		return err
	}
	if id == nil {
		return thirdpartyservice.CreateService(*txn, service)
	}
	log.Info("ThirdPartyService already exists. skip creating...", "name", service.Name)
	return nil
}

func (s ThirdPartyService) GetServicesId(txn *sql.Tx, serviceNames []string, schema string) ([]int, error) {
	var servicesId []int
	for _, name := range serviceNames {
		id, err := thirdpartyservice.GetService(*txn, name, schema)
		if err != nil {
			return nil, err
		}
		servicesId = append(servicesId, *id)
	}
	return servicesId, nil
}
