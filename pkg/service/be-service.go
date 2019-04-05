package service

import (
	"business-app-reconciler-controller/pkg/model"
	"business-app-reconciler-controller/pkg/repository"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type BEService struct {
	DB sql.DB
}

func (service BEService) PutBE(be model.BusinessEntity) error {
	log.Printf("Start creation of business entity %v...", be)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	id, err := getBeIdOrCreate(*txn, be)
	if err != nil {
		log.Printf("Error has occurred during get BE id or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of BE to be updated: %v", *id)

	isPresent, err := checkActionLogDuplicate(*txn, be)
	if err != nil {
		_ = txn.Rollback()
		return err
	}

	if !isPresent {
		log.Println("Start update status of business application...")
		codebaseActionId, err := repository.CreateActionLog(*txn, be.ActionLog)
		if err != nil {
			log.Printf("Error has occurred during status creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot create business entity %v", be))
		}
		log.Println("ActionLog has been saved into the repository")

		log.Println("Start update codebase_action status of business application...")
		err = repository.CreateCodebaseAction(*txn, *id, *codebaseActionId)
		if err != nil {
			log.Printf("Error has occurred during codebase_action creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot create codebase_action entity %v", be))
		}
		log.Println("codebase_action has been updated")
	}

	_ = txn.Commit()

	log.Println("Business entity has been saved successfully")

	return nil
}

func getBeIdOrCreate(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	log.Printf("Start retrieving BE by name, tenant and type: %v", be)
	id, err := repository.GetBEId(txn, be)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for BE %v has not been found", be)
		return createBE(txn, be)
	}
	return id, nil
}

func createBE(txn sql.Tx, be model.BusinessEntity) (*int, error) {
	log.Println("Start insertion in the repository business entity...")
	id, err := repository.CreateBE(txn, be)
	if err != nil {
		log.Printf("Error has occurred during business entity creation: %v", err)
		return nil, errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of the newly created business entity is %v", *id)
	return id, nil
}

func checkActionLogDuplicate(txn sql.Tx, be model.BusinessEntity) (bool, error) {
	log.Println("Checks duplicate in action log table")
	lastId, err := repository.GetLastIdActionLog(txn, be)
	if err != nil {
		log.Printf("Error has occurred while checking on duplicate: %v", err)
		return false, errors.New(fmt.Sprintf("cannot check duplication %v", be))
	}
	if lastId == nil {
		return false, nil
	}
	return true, nil
}
