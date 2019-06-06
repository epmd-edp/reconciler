package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reconciler/pkg/model"
	"reconciler/pkg/repository"
)

type BEService struct {
	DB *sql.DB
}

func (service BEService) PutBE(be model.Codebase) error {
	log.Printf("Start creation of business entity %v...", be)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}

	schemaName := be.Tenant

	id, err := getBeIdOrCreate(*txn, be, schemaName)
	if err != nil {
		log.Printf("Error has occurred during get BE id or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of BE to be updated: %v", *id)

	isPresent, err := checkActionLogDuplicate(*txn, be, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return err
	}

	if !isPresent {
		log.Println("Start update status of codebase...")
		codebaseActionId, err := repository.CreateActionLog(*txn, be.ActionLog, schemaName)
		if err != nil {
			log.Printf("Error has occurred during status creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot insert status %v", be))
		}
		log.Println("ActionLog has been saved into the repository")

		log.Println("Start update codebase_action status of codebase...")
		err = repository.CreateCodebaseAction(*txn, *id, *codebaseActionId, schemaName)
		if err != nil {
			log.Printf("Error has occurred during codebase_action creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot create codebase_action entity %v", be))
		}
		log.Println("codebase_action has been updated")
	}

	err = repository.UpdateStatusByCodebaseId(*txn, *id, be.Status, be.Tenant)
	if err != nil {
		log.Printf("Error has occurred during the update of codebase: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create codebase with name %v", be.Name))
	}

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Println("Business entity has been saved successfully")

	return nil
}

func getBeIdOrCreate(txn sql.Tx, be model.Codebase, schemaName string) (*int, error) {
	log.Printf("Start retrieving BE by name, tenant and type: %v", be)
	id, err := repository.GetCodebaseId(txn, be.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for BE %v has not been found", be)
		return createBE(txn, be, schemaName)
	}
	return id, nil
}

func createBE(txn sql.Tx, be model.Codebase, schemaName string) (*int, error) {
	log.Println("Start insertion in the repository business entity...")
	id, err := repository.CreateCodebase(txn, be, schemaName)
	if err != nil {
		log.Printf("Error has occurred during business entity creation: %v", err)
		return nil, errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of the newly created business entity is %v", *id)
	return id, nil
}

func checkActionLogDuplicate(txn sql.Tx, be model.Codebase, schemaName string) (bool, error) {
	log.Println("Checks duplicate in action log table")
	lastId, err := repository.GetLastIdActionLog(txn, be, schemaName)
	if err != nil {
		log.Printf("Error has occurred while checking on duplicate: %v", err)
		return false, errors.New(fmt.Sprintf("cannot check duplication %v", be))
	}
	if lastId == nil {
		return false, nil
	}
	return true, nil
}
