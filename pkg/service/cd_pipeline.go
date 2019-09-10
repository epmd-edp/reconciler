package service

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"log"
	"sort"
)

type CdPipelineService struct {
	DB        *sql.DB
	ClientSet platform.ClientSet
}

func (s CdPipelineService) PutCDPipeline(cdPipeline model.CDPipeline) error {
	log.Printf("Start creation of CD Pipeline %v...", cdPipeline)
	log.Println("Start transaction...")

	txn, err := s.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New("error has occurred during opening transaction")
	}

	schemaName := cdPipeline.Tenant

	cdPipelineDb, err := s.getCDPipelineOrCreate(*txn, cdPipeline, schemaName)
	if err != nil {
		log.Printf("Error has occurred during get CD pipeline or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create CD Pipeline entity %v", cdPipeline))
	}
	log.Printf("Id of CD Pipeline to be updated: %v", cdPipelineDb.Id)

	err = updateCDPipelineStatus(*txn, *cdPipelineDb, cdPipeline.Status, schemaName)
	if err != nil {
		log.Printf("An error has occured while updating CD Pipeline Status for %s pipeline: %s", cdPipelineDb.Name, err)
		_ = txn.Rollback()
		return err
	}

	err = updateActionLog(*txn, cdPipeline, cdPipelineDb.Id, schemaName)
	if err != nil {
		log.Printf("An error has occured while updating CD Pipeline Action Event Log for %s pipeline: %s", cdPipeline.Name, err)
		return err
	}

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Println("CD Pipeline has been saved successfully")

	return nil
}

func (s CdPipelineService) getCDPipelineOrCreate(txn sql.Tx, cdPipeline model.CDPipeline, schemaName string) (*model.CDPipelineDTO, error) {
	log.Printf("Start retrieving CD Pipeline by name: %v", cdPipeline)
	cdPipelineReadModel, err := repository.GetCDPipeline(txn, cdPipeline.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if cdPipelineReadModel != nil {

		err = repository.DeleteCDPipelineDockerStreams(txn, cdPipelineReadModel.Id, schemaName)
		if err != nil {
			log.Printf("An error has occurred while deleting pipeline's docker streams: %s", err)
			return nil, err
		}

		err = createCDPipelineDockerStream(txn, cdPipelineReadModel.Id, cdPipeline.InputDockerStreams, schemaName)
		if err != nil {
			return nil, err
		}

		stages, err := getStages(txn, cdPipelineReadModel.Name, schemaName)
		if err != nil {
			return nil, err
		}

		sort.SliceStable(stages, func(i, j int) bool {
			return stages[i].Order < stages[j].Order
		})

		err = s.updateStageCodebaseDockerStream(txn, stages, cdPipelineReadModel.Name, schemaName)
		if err != nil {
			return nil, err
		}

		err = updateApplicationsToPromote(txn, cdPipelineReadModel.Id, cdPipeline.ApplicationsToPromote, schemaName)
		if err != nil {
			return nil, err
		}

		return cdPipelineReadModel, nil
	}
	log.Printf("Record for CD Pipeline %v has not been found", cdPipeline.Name)
	cdPipelineDTO, err := createCDPipeline(txn, cdPipeline, schemaName)
	if err != nil {
		return nil, err
	}

	err = createCDPipelineDockerStream(txn, cdPipelineDTO.Id, cdPipeline.InputDockerStreams, schemaName)
	if err != nil {
		return nil, err
	}

	if cdPipeline.ThirdPartyServices != nil && len(cdPipeline.ThirdPartyServices) != 0 {
		log.Printf("Try to create records in ThirdPartyServices: %v", cdPipeline.ThirdPartyServices)

		servicesId, err := getServicesId(txn, cdPipeline.ThirdPartyServices, schemaName)
		if err != nil {
			log.Printf("An error has occured while getting services id: %s", err)
			return nil, err
		}

		log.Printf("Try to create record for %v service", servicesId)

		err = createCDPipelineThirdPartyService(txn, cdPipelineDTO.Id, servicesId, schemaName)
		if err != nil {
			log.Printf("An error has occured while inserting record into cd_pipeline_third_party_service: %v", err)
			return nil, err
		}
	}

	err = createApplicationToPromoteRow(txn, cdPipelineDTO.Id, cdPipeline.ApplicationsToPromote, schemaName)
	if err != nil {
		log.Printf("An error has occured while inserting record into applications_to_promote: %v", err)
		return nil, err
	}

	return cdPipelineDTO, nil
}

func updateApplicationsToPromote(tx sql.Tx, cdPipelineId int, applicationsToPromote []string, schemaName string) error {
	err := repository.RemoveApplicationsToPromote(tx, cdPipelineId, schemaName)
	if err != nil {
		return fmt.Errorf("an error has occurred while removing Application To Promote records for %v Stage: %v", cdPipelineId, err)
	}

	err = createApplicationToPromoteRow(tx, cdPipelineId, applicationsToPromote, schemaName)
	if err != nil {
		return fmt.Errorf("an error has occurred while creating Application To Promote record for %v Stage: %v", cdPipelineId, err)
	}

	return nil
}

func createApplicationToPromoteRow(txn sql.Tx, cdPipelineId int, applicationsToPromote []string, schemaName string) error {
	log.Printf("Try to create record in ApplicationToPromote table %v ...", applicationsToPromote)

	for _, appToPromote := range applicationsToPromote {
		id, err := repository.GetApplicationId(txn, appToPromote, schemaName)
		if err != nil {
			return err
		}

		log.Printf("Application Id %v by %v app name", id, appToPromote)

		err = repository.CreateApplicationsToPromote(txn, cdPipelineId, *id, schemaName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s CdPipelineService) updateStageCodebaseDockerStreamRelations(txn sql.Tx, stages []model.Stage, pipelineName string, schemaName string) error {
	log.Printf("Try to update Stage Codebase Docker Streams relations for stages: %v", stages)

	for i := range stages {
		stages[i].Tenant = schemaName
		stages[i].CdPipelineName = pipelineName

		pipelineCR, err := getCDPipelineCR(s.ClientSet.EDPRestClient, stages[i].CdPipelineName, stages[i].Tenant+"-edp-cicd")
		if err != nil {
			return err
		}

		err = updateSingleStageCodebaseDockerStreamRelations(txn, stages[i].Id, stages[i], pipelineCR.Spec.ApplicationsToPromote)
		if err != nil {
			log.Printf("Error has occurred while creating Codebase Docker Stream row: %v", err)
			return err
		}
	}

	log.Printf("Relations have been updated for %v pipeline", pipelineName)

	return nil
}

func getStages(txn sql.Tx, cdPipelineName string, schemaName string) ([]model.Stage, error) {
	stages, err := repository.GetStages(txn, cdPipelineName, schemaName)
	if err != nil {
		log.Printf("An error has occured while getting Stages for CD Pipeline %v : %v", cdPipelineName, err)
		return nil, err
	}
	log.Printf("Fetched Stages %v for CD Pipeline %v", stages, cdPipelineName)

	return stages, nil
}

func deleteStageCodebaseDockerStream(txn sql.Tx, stages []model.Stage, schemaName string) ([]int, error) {
	var outputStreamIdsToRemove []int
	var stagesToLog []string

	for _, stage := range stages {
		outputStreamIds, err := repository.DeleteStageCodebaseDockerStream(txn, stage.Id, schemaName)
		outputStreamIdsToRemove = append(outputStreamIdsToRemove, outputStreamIds...)
		if err != nil {
			log.Printf("An error has occured while deleting stage codebase docker stream row: %v", err)
			return nil, err
		}
		stagesToLog = append(stagesToLog, stage.Name)
	}

	log.Printf("Collected Output Stream Ids to delete: %v", outputStreamIdsToRemove)

	return outputStreamIdsToRemove, nil
}

func (s CdPipelineService) updateStageCodebaseDockerStream(txn sql.Tx, stages []model.Stage, pipelineName string, schemaName string) error {
	if stages == nil {
		log.Printf("There're no stages for %v CD Pipeline. Updating of Codebase Docker stream will not be executed.", pipelineName)
		return nil
	}

	_, err := deleteStageCodebaseDockerStream(txn, stages, schemaName)
	if err != nil {
		return err
	}

	err = s.updateStageCodebaseDockerStreamRelations(txn, stages, pipelineName, schemaName)
	if err != nil {
		return err
	}

	return nil
}

func createCDPipeline(txn sql.Tx, cdPipeline model.CDPipeline, schemaName string) (*model.CDPipelineDTO, error) {
	log.Println("Start insertion to the cd_pipeline table...")
	cdPipelineDto, err := repository.CreateCDPipeline(txn, cdPipeline, cdPipeline.Status, schemaName)
	if err != nil {
		return nil, err
	}

	log.Printf("Id of the newly created CD Pipeline is %v", cdPipelineDto.Id)
	return cdPipelineDto, nil
}

func updateActionLog(txn sql.Tx, cdPipeline model.CDPipeline, pipelineId int, schemaName string) error {
	log.Println("Start update status of CD Pipeline...")
	actionLogId, err := repository.CreateEventActionLog(txn, cdPipeline.ActionLog, schemaName)
	if err != nil {
		log.Printf("Error has occurred during status creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot insert status %v", cdPipeline))
	}
	log.Println("ActionLog row has been saved into the repository")

	log.Println("Start update cd_pipeline_codebase_action status of code pipeline entity...")
	err = repository.CreateCDPipelineActionLog(txn, pipelineId, *actionLogId, schemaName)
	if err != nil {
		log.Printf("Error has occurred during cd_pipeline_action creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create cd_pipeline_action entity %v", cdPipeline))
	}
	log.Println("cd_pipeline_action has been updated")
	return nil
}

func updateCDPipelineStatus(txn sql.Tx, cdPipelineDb model.CDPipelineDTO, status string, schemaName string) error {
	if cdPipelineDb.Status != status {
		log.Printf("Start updating status of %s pipeline to %s", cdPipelineDb.Name, status)
		err := repository.UpdateCDPipelineStatus(txn, cdPipelineDb.Id, status, schemaName)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCDPipelineThirdPartyService(txn sql.Tx, cdPipelineId int, servicesId []int, schemaName string) error {
	for _, serviceId := range servicesId {
		err := repository.CreateCDPipelineThirdPartyService(txn, cdPipelineId, serviceId, schemaName)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCDPipelineDockerStream(txn sql.Tx, cdPipelineId int, dockerStreams []string, schemaName string) error {
	var dockerStreamIds []int
	for _, dockerStream := range dockerStreams {
		id, err := repository.GetCodebaseDockerStreamId(txn, dockerStream, schemaName)
		if err != nil {
			log.Printf("An error has occured while getting id of docker stream %v: %v", dockerStream, err)
			return err
		}
		dockerStreamIds = append(dockerStreamIds, *id)
	}

	err := insertCDPipelineDockerStream(txn, cdPipelineId, dockerStreamIds, schemaName)
	if err != nil {
		log.Printf("An error has occured while inserting CD Pipeline Docker Stream row: %s", err)
		return err
	}

	return nil
}

func insertCDPipelineDockerStream(txn sql.Tx, cdPipelineId int, dockerStreams []int, schemaName string) error {
	for _, id := range dockerStreams {
		err := repository.CreateCDPipelineDockerStream(txn, cdPipelineId, id, schemaName)
		if err != nil {
			log.Printf("An error has occured while inserting CD Pipeline Docker Stream row: %s", err)
			return err
		}
	}
	return nil
}
