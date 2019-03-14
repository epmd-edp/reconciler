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

func (service BEService) CreateBE(be model.BusinessEntity) error {
	log.Printf("Start creation of business entity %+v...", be)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Println("Start insertion in the repository business entity...")
	id, err := repository.CreateBE(*txn, be)
	if err != nil {
		log.Printf("Error has occurred during business entity creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of the newly created business entity is %v", *id)
	log.Println("Start property creation...")
	err = repository.CreateProps(*txn, *id, be.Props)
	if err != nil {
		log.Printf("Error has occurred during property creation %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Println("Properties has been saved successfully")
	log.Printf("Start retrieving status id by message %v", be.Status.Message)
	statusId, err := repository.GetStatusId(*txn, be.Status.Message)
	if err != nil {
		log.Printf("Error has occurred during status retrieving: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Status id %v has been retrived for message %v", *statusId, be.Status.Message)
	be.Status.Id = *statusId
	log.Println("Start update status of business application...")
	err = repository.CreateStatus(*txn, *id, be.Status)
	if err != nil {
		log.Printf("Error has occurred during status creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Println("Status has been saved into the repository")

	_ = txn.Commit()

	log.Println("Business entity has been saved successfully")

	return nil
}
