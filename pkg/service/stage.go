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

	prevStageAdded := canStageBeCreated(*txn, stage)

	if !prevStageAdded {
		log.Printf("previous stage has not been added yet for stage %v", stage)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}

	id, err := getStageIdOrCreate(*txn, stage)
	if err != nil {
		log.Printf("error has occured during retrieving or creation stage: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}
	log.Printf("Id of stage to be updated: %v", *id)

	err = createCodebaseDockerStreams(*txn, *id, stage)
	if err != nil {
		log.Printf("error has occured during the creation docker streams: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}

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

func createCodebaseDockerStreams(tx sql.Tx, id int, stage model.Stage) error {
	log.Printf("Start creation of the docker streams for stage with id: %v", id)

	inputDockerStreams, err := getInputDockerStreams(tx, id, stage)
	if err != nil {
		log.Printf("Cannot get list of input docker streams for stage with id : %v", id)
		return err
	}

	err = createOutputStreamsAndLink(tx, id, stage, inputDockerStreams)

	if err != nil {
		log.Printf("Cannot create output streams for stage with id: %v", id)
		return err
	}

	log.Printf("Docker streams have been successfully created for stage with id: %v", id)
	return nil
}

func createOutputStreamsAndLink(tx sql.Tx, id int, stage model.Stage, dtos []model.CodebaseDockerStreamReadDTO) error {
	log.Printf("Start creation of outputstreams and links for stage with id: %v", id)
	for _, stream := range dtos {
		err := createSingleOutputStreamAndLink(tx, id, stage, stream)
		if err != nil {
			return err
		}
	}
	return nil
}

func createSingleOutputStreamAndLink(tx sql.Tx, stageId int, stage model.Stage, dto model.CodebaseDockerStreamReadDTO) error {
	log.Printf("Start creation single outputstream and link for stage with id %v and stream: %v", stageId, dto)

	ocImageStreamName := fmt.Sprintf("%v-%v-%v-verified", stage.CdPipelineName, stage.Name, dto.CodebaseName)

	outputId, err := repository.CreateCodebaseDockerStream(tx, stage.Tenant, dto.CodebaseId, ocImageStreamName)
	if err != nil {
		log.Printf("Cannot create codebase docker stream for dto: %v", dto)
		return err
	}
	log.Printf("Id of newly created docker stream is: %v", *outputId)

	err = repository.CreateStageCodebaseDockerStream(tx, stage.Tenant, stageId, dto.CodebaseDockerStreamId, *outputId)

	if err != nil {
		log.Printf("Cannot link codebase docker stream for dto: %v", dto)
		return err
	}

	log.Printf("End creation single outputstream and link for stage with id %v and stream: %v", stageId, dto)
	return nil
}

func getInputDockerStreams(tx sql.Tx, id int, stage model.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.Printf("Start read input docker streams for stage with id: %v", id)
	if stage.Order == 0 {
		return getInputDockerStreamsForFirstStage(tx, id, stage)
	}
	return getInputDockerStreamsForArbitraryStage(tx, id, stage)
}

func getInputDockerStreamsForArbitraryStage(tx sql.Tx, id int, stage model.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.Printf("Start read input docker streams for the arbitrary stage with id: %v", id)
	streams, err := repository.GetDockerStreamsByPipelineNameAndStageOrder(tx, stage.Tenant, stage.CdPipelineName, stage.Order-1)
	if err != nil {
		log.Printf("Error has been occured during the read docker streams by pipiline name %v and stage order: %v", stage.CdPipelineName, stage.Order-1)
		return nil, err
	}
	log.Printf("Streams have been successfully retrieved: %v", streams)
	return streams, nil
}

func getInputDockerStreamsForFirstStage(tx sql.Tx, id int, stage model.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.Printf("Start read input docker streams for the first stage with id: %v", id)
	streams, err := repository.GetDockerStreamsByPipelineName(tx, stage.Tenant, stage.CdPipelineName)
	if err != nil {
		log.Printf("Error has been occured during the read docker streams by pipiline name : %v", stage.CdPipelineName)
		return nil, err
	}
	log.Printf("Streams have been successfully retrieved: %v", streams)
	return streams, nil
}

func canStageBeCreated(tx sql.Tx, stage model.Stage) bool {
	if stage.Order == 0 {
		log.Printf("Stage %v is the first in the chain. Returning true..", stage)
		return true
	}
	return prevStageAdded(tx, stage)
}

func prevStageAdded(tx sql.Tx, stage model.Stage) bool {
	log.Printf("Check previous stage fot stage: %v", stage)
	stageId, err := repository.GetStageIdByPipelineNameAndOrder(tx, stage.Tenant, stage.CdPipelineName, stage.Order-1)

	if err != nil {
		log.Printf("Error has been occured during the retrieving prev stage id : %v", err)
		return false
	}

	if stageId == nil {
		log.Printf("Previous stage for stage %v has not been added. Returning false", stage)
		return false
	}

	log.Printf("Id of previous stage is %v", stageId)
	return true
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
