package service

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/thirdpartyservice"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"log"
)

type ThirdPartyService struct {
	DB *sql.DB
}

func (c ThirdPartyService) PutService(service thirdpartyservice.ThirdPartyService) error {
	log.Printf("Start ThirdPartyService entity creation %v...", service)
	log.Println("Start transaction...")
	txn, err := c.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create service entity %v", service))
	}

	schemaName := service.Tenant

	id, err := skipCatalogOrCreate(txn, service, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return err
	}
	log.Printf("Id of the newly created ThirdPartyService entity is %v", *id)

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Printf("ThirdPartyService entity has been saved successfully: %v", service)

	return nil

}

func skipCatalogOrCreate(txn *sql.Tx, service thirdpartyservice.ThirdPartyService, schemaName string) (*int, error) {
	log.Printf("Start retrieving ThirdPartyService by name and tenant: %v", service)
	id, err := repository.GetThirdPartyService(*txn, service.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for ThirdPartyService %v has not been found", service)
		return createService(txn, service, schemaName)
	}
	return id, nil
}

func createService(txn *sql.Tx, service thirdpartyservice.ThirdPartyService, schemaName string) (*int, error) {
	log.Println("Start ThirdPartyService entity saving...")
	id, err := repository.CreateThirdPartyService(*txn, service, schemaName)
	if err != nil {
		log.Printf("Error has occurred during ThirdPartyService entity creation: %v", err)
		return nil, errors.New(fmt.Sprintf("cannot create ThirdPartyService entity %v", service))
	}
	return id, nil
}

func GetServicesId(txn *sql.Tx, serviceNames []string, schemaName string) ([]int, error) {
	var servicesId []int
	for _, name := range serviceNames {
		id, err := repository.GetThirdPartyService(*txn, name, schemaName)
		if err != nil {
			return nil, err
		}
		servicesId = append(servicesId, *id)
	}
	return servicesId, nil
}
