package service

import (
	"database/sql"
	"errors"
	"fmt"
	"k8s.io/client-go/rest"
	"log"
	"reconciler/pkg/apis/edp/v1alpha1"
	"reconciler/pkg/model"
	"reconciler/pkg/platform"
	"reconciler/pkg/repository"
	"sort"
)

type CdPipelineService struct {
	DB        *sql.DB
	ClientSet platform.ClientSet
}

func (service CdPipelineService) PutCDPipeline(cdPipeline model.CDPipeline) error {
	log.Printf("Start creation of CD Pipeline %v...", cdPipeline)
	log.Println("Start transaction...")

	txn, err := service.DB.Begin()
	edpRestClient := service.ClientSet.EDPRestClient
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New("error has occurred during opening transaction")
	}

	schemaName := cdPipeline.Tenant

	cdPipelineDb, err := getCDPipelineOrCreate(*txn, edpRestClient, cdPipeline, schemaName)
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

func getCDPipelineOrCreate(txn sql.Tx, edpRestClient *rest.RESTClient, cdPipeline model.CDPipeline, schemaName string) (*model.CDPipelineDTO, error) {
	log.Printf("Start retrieving CD Pipeline by name: %v", cdPipeline)
	cdPipelineReadModel, err := repository.GetCDPipeline(txn, cdPipeline.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if cdPipelineReadModel != nil {

		err = repository.DeleteBranches(txn, cdPipelineReadModel.Id, schemaName)
		if err != nil {
			log.Printf("An error has occurred while deleting pipeline's branches: %s", err)
			return nil, err
		}

		err = createCDPipelineCodebaseBranch(txn, edpRestClient, cdPipelineReadModel, cdPipeline, schemaName)
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

		err = tryToCreateCodebaseDockerStream(txn, edpRestClient, stages, cdPipelineReadModel.Name, schemaName)
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

	err = createCDPipelineCodebaseBranch(txn, edpRestClient, cdPipelineDTO, cdPipeline, schemaName)
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

func createCodebaseDockerStreamsRow(txn sql.Tx, edpRestClient *rest.RESTClient, stages []model.Stage, pipelineName string, schemaName string) error {
	log.Printf("Try to create Codebase Docker Streams for stages: %v", stages)

	for i := range stages {
		stages[i].Tenant = schemaName
		stages[i].CdPipelineName = pipelineName

		pipelineCR, err := getCDPipelineCR(edpRestClient, stages[i].CdPipelineName, stages[i].Tenant+"-edp-cicd")
		if err != nil {
			return err
		}

		err = createCodebaseDockerStreams(txn, stages[i].Id, stages[i], pipelineCR.Spec.ApplicationsToPromote)
		if err != nil {
			log.Printf("Error has occurred while creating Codebase Docker Stream row: %v", err)
			return err
		}
	}

	log.Printf("Codebase Docker Stream Row has been created for %v pipeline", pipelineName)

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

func deleteStageCodebaseDockerStream(txn sql.Tx, stages []model.Stage, schemaName string) error {
	var outputStreamIdsToRemove []int
	var stagesToLog []string

	for _, stage := range stages {
		outputStreamIds, err := repository.DeleteStageCodebaseDockerStream(txn, stage.Id, schemaName)
		outputStreamIdsToRemove = append(outputStreamIdsToRemove, outputStreamIds...)
		if err != nil {
			log.Printf("An error has occured while deleting stage codebase docker stream row: %v", err)
			return err
		}
		stagesToLog = append(stagesToLog, stage.Name)
	}
	log.Printf("Collected Output Stream Ids to delete: %v", outputStreamIdsToRemove)

	err := deleteCodebaseDockerStreams(txn, outputStreamIdsToRemove, schemaName)
	if err != nil {
		return err
	}

	log.Printf("All records in StageCodebaseDockerStream and CodebaseDockerStream for %v Stages are removed.", stagesToLog)

	return nil
}

func deleteCodebaseDockerStreams(txn sql.Tx, outputStreamIds []int, schemaName string) error {
	for _, id := range outputStreamIds {
		err := repository.DeleteCodebaseDockerStream(txn, id, schemaName)
		if err != nil {
			log.Printf("An error has occured while deleting codebase docker stream row: %v", err)
			return err
		}
	}
	return nil
}

func tryToCreateCodebaseDockerStream(txn sql.Tx, edpRestClient *rest.RESTClient, stages []model.Stage, pipelineName string, schemaName string) error {
	if stages == nil {
		log.Printf("There're no stages for %v CD Pipeline. Updating of Codebase Docker stream will not be executed.", pipelineName)
		return nil
	}

	err := deleteStageCodebaseDockerStream(txn, stages, schemaName)
	if err != nil {
		return err
	}

	err = createCodebaseDockerStreamsRow(txn, edpRestClient, stages, pipelineName, schemaName)
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

func getCodebaseBranchCR(edpRestClient *rest.RESTClient, crName string, namespace string) (*v1alpha1.CodebaseBranch, error) {
	codebaseBranch := &v1alpha1.CodebaseBranch{}
	err := edpRestClient.Get().Namespace(namespace).Resource("codebasebranches").Name(crName).Do().Into(codebaseBranch)
	if err != nil {
		log.Printf("An error has occurred while getting Release Branch CR from k8s: %s", err)
		return nil, err
	}
	return codebaseBranch, nil
}

func getCodebaseBranchesData(cdPipeline model.CDPipeline, edpRestClient *rest.RESTClient) ([]model.CodebaseBranchDTO, error) {
	var codebaseBranches []model.CodebaseBranchDTO
	for _, v := range cdPipeline.CodebaseBranch {
		releaseBranchCR, err := getCodebaseBranchCR(edpRestClient, v, cdPipeline.Namespace)
		if err != nil {
			return nil, err
		}

		codebaseBranches = append(codebaseBranches, model.CodebaseBranchDTO{
			CodebaseName: releaseBranchCR.Spec.CodebaseName,
			BranchName:   releaseBranchCR.Spec.BranchName,
		})
	}
	return codebaseBranches, nil
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

func getCodebaseBranchesId(txn sql.Tx, codebaseBranches []model.CodebaseBranchDTO, schemaName string) ([]int, error) {
	var codebaseBranchesId []int
	for _, v := range codebaseBranches {
		codebaseBranchId, err := repository.GetCodebaseBranchesId(txn, v, schemaName)
		if err != nil {
			return nil, err
		}
		codebaseBranchesId = append(codebaseBranchesId, *codebaseBranchId)
	}
	return codebaseBranchesId, nil
}

func insertCDPipelineCodebaseBranch(txn sql.Tx, cdPipelineId int, codebaseBranchesId []int, schemaName string) error {
	for _, v := range codebaseBranchesId {
		err := repository.CreateCDPipelineCodebaseBranch(txn, cdPipelineId, v, schemaName)
		if err != nil {
			log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
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

func createCDPipelineCodebaseBranch(txn sql.Tx, edpRestClient *rest.RESTClient, cdPipelineReadModel *model.CDPipelineDTO, cdPipeline model.CDPipeline, schemaName string) error {
	codebaseBranches, err := getCodebaseBranchesData(cdPipeline, edpRestClient)
	if err != nil {
		log.Printf("An error has occured while getting Codebase Branch from k8s: %s", err)
		return err
	}
	log.Printf("Fetched Codebase Branches for %s pipeline: %s", cdPipeline.Name, codebaseBranches)

	codebaseBranchesId, err := getCodebaseBranchesId(txn, codebaseBranches, schemaName)
	if err != nil {
		log.Printf("An error has occured while getting codebase branch id: %s", err)
		return err
	}

	err = insertCDPipelineCodebaseBranch(txn, cdPipelineReadModel.Id, codebaseBranchesId, schemaName)
	if err != nil {
		log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
		return err
	}

	return nil
}
