package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reconciler/pkg/model"
	"reconciler/pkg/repository"
)

type StageService struct {
	DB sql.DB
}

func (service StageService) PutStage(stage model.Stage) error {
	log.Printf("Start put stage: %v ...", stage)

	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New("error has occurred during opening transaction")
	}

	id, err := getStageIdOrCreate(*txn, stage)
	if err != nil {
		log.Printf("error has occured during retrieving or creation stage: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}
	log.Printf("Id of stage to be updated: %v", *id)

	err = updateStageStatus(*txn, id, stage)

	if err != nil {
		log.Printf("error has occured during the updating stage status: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}

	err = addActionLog(*txn, id, stage)

	if err != nil {
		log.Printf("error has occured during the adding action log: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}

	_ = txn.Commit()

	log.Printf("Stage %v has been inserted successfully", stage)
	return nil
}

func addActionLog(tx sql.Tx, id *int, stage model.Stage) error {
	log.Printf("Start adding action log: %v for stage with id %v", stage.ActionLog, *id)
	actionLogId, err := repository.CreateEventActionLog(tx, stage.ActionLog, stage.Tenant)
	if err != nil {
		return err
	}
	log.Printf("Action log has been added. Id of newly created al is %v", *actionLogId)
	err = repository.CreateStageActionLog(tx, stage.Tenant, *id, *actionLogId)
	if err != nil {
		return err
	}
	log.Printf("Action log %v for stage with id %v has been updated successfully", stage.ActionLog, *id)
	return nil
}

func updateStageStatus(tx sql.Tx, id *int, stage model.Stage) error {
	log.Printf("Start updating status: %v for stage with id %v", stage.Status, *id)
	err := repository.UpdateStageStatus(tx, stage.Tenant, *id, stage.Status)
	if err != nil {
		return err
	}
	log.Printf("Status for stage with id %v has been successfully updated to %v", *id, stage.Status)
	return nil
}

func getStageIdOrCreate(tx sql.Tx, stage model.Stage) (*int, error) {
	log.Printf("Start get stage id or create for stage: %v", stage)
	id, err := repository.GetStageId(tx, stage.Tenant, stage.Name, stage.CdPipelineName)
	if err != nil {
		return nil, err
	}
	if id != nil {
		log.Printf("Stage %v is already presented. Returning id; %v", stage, *id)
		return id, err
	}
	return createStage(tx, stage)
}

func createStage(tx sql.Tx, stage model.Stage) (*int, error) {
	log.Printf("Start create stage %v", stage)
	cdPipeline, err := repository.GetCDPipeline(tx, stage.CdPipelineName, stage.Tenant)
	if err != nil {
		log.Printf("Error has been occured during the reading cd pipeline by name: %v", cdPipeline)
		return nil, err
	}
	if cdPipeline == nil {
		return nil, fmt.Errorf("record for cd pipeline with name %v has not been found", stage.CdPipelineName)
	}
	id, err := repository.CreateStage(tx, stage.Tenant, stage, cdPipeline.Id)
	if err != nil {
		return nil, err
	}
	log.Printf("Id of newly created stage is %v", *id)
	return id, nil
}
