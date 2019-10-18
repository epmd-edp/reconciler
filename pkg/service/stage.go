package service

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"k8s.io/client-go/rest"
	"log"
)

type StageService struct {
	DB        *sql.DB
	ClientSet platform.ClientSet
}

//PutStage creates record in DB for Stage.
//The main cases which method do:
//	- checks if stage can be created (checks if previous stage has been added)
//	- update stage status
//	- add record to Action Log for last operation
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

	edpRestClient := service.ClientSet.EDPRestClient

	id, err := getStageIdOrCreate(*txn, edpRestClient, stage)
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

	cdPipelineReadModel, err := repository.GetCDPipeline(*txn, stage.CdPipelineName, stage.Tenant)
	if err != nil {
		log.Printf("error has occured while fetching CD Pipeline %v: %v", stage.CdPipelineName, err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot fetch CD Pipeline: %v", stage.CdPipelineName)
	}

	err = addActionLog(*txn, &cdPipelineReadModel.Id, stage)

	if err != nil {
		log.Printf("error has occured during the adding action log: %v", err)
		_ = txn.Rollback()
		return fmt.Errorf("cannot create stage: %v", stage)
	}

	_ = txn.Commit()

	log.Printf("Stage %v has been inserted successfully", stage)
	return nil
}

func createCodebaseDockerStreams(tx sql.Tx, id int, stage model.Stage, applicationsToApprove []string) error {
	log.Printf("Start creation of the docker streams for stage with id: %v", id)

	inputDockerStreams, err := getInputDockerStreams(tx, id, stage)
	if err != nil {
		log.Printf("Cannot get list of input docker streams for stage with id : %v", id)
		return err
	}

	err = createOutputStreamsAndLink(tx, id, stage, inputDockerStreams, applicationsToApprove)

	if err != nil {
		log.Printf("Cannot create output streams for stage with id: %v", id)
		return err
	}

	log.Printf("Docker streams have been successfully created for stage with id: %v", id)
	return nil
}

func updateSingleStageCodebaseDockerStreamRelations(tx sql.Tx, id int, stage model.Stage, applicationsToApprove []string) error {
	log.Printf("Start update of the docker streams relation for stage with id: %v", id)

	inputDockerStreams, err := getInputDockerStreams(tx, id, stage)
	if err != nil {
		log.Printf("Cannot get list of input docker streams for stage with id : %v", id)
		return err
	}

	err = updateOutputStreamsRelation(tx, id, stage, inputDockerStreams, applicationsToApprove)

	if err != nil {
		log.Printf("Cannot create output streams for stage with id: %v", id)
		return err
	}

	log.Printf("Docker streams relation have been successfully updated for stage with id: %v", id)
	return nil
}

func createOutputStreamsAndLink(tx sql.Tx, id int, stage model.Stage, dtos []model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.Printf("Start creation of outputstreams and links for stage with id: %v", id)
	for _, stream := range dtos {
		err := createSingleOutputStreamAndLink(tx, id, stage, stream, applicationsToApprove)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateOutputStreamsRelation(tx sql.Tx, id int, stage model.Stage, dtos []model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.Printf("Start update of links for stage with id: %v", id)
	for _, stream := range dtos {
		err := updateSingleOutputStreamRelation(tx, id, stage, stream, applicationsToApprove)
		if err != nil {
			return err
		}
	}
	return nil
}

func createSingleOutputStreamAndLink(tx sql.Tx, stageId int, stage model.Stage, dto model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.Printf("Start creation single outputstream and link for stage with id %v and stream: %v", stageId, dto)

	ocImageStreamName := fmt.Sprintf("%v-%v-%v-verified", stage.CdPipelineName, stage.Name, dto.CodebaseName)

	branchId, err := repository.GetCodebaseDockerStreamBranchId(tx, dto.CodebaseDockerStreamId, stage.Tenant)
	if err != nil {
		log.Printf("Cannot get branch id by codebase docker stream id %v: %v", dto.CodebaseDockerStreamId, err)
		return err
	}

	outputId, err := repository.CreateCodebaseDockerStream(tx, stage.Tenant, branchId, ocImageStreamName)
	if err != nil {
		log.Printf("Cannot create codebase docker stream for dto: %v", dto)
		return err
	}
	log.Printf("Id of newly created docker stream is: %v", *outputId)

	stage.Id = stageId
	if include(applicationsToApprove, dto.CodebaseName) {
		err = setPreviousStageInputImageStream(tx, stage, dto.CodebaseDockerStreamId, *outputId)
	} else {
		err = setOriginalInputImageStream(tx, stage, dto.CodebaseName, *outputId)
	}

	if err != nil {
		log.Printf("Cannot link codebase docker stream for dto: %v", dto)
		return err
	}

	log.Printf("End creation single outputstream and link for stage with id %v and stream: %v", stageId, dto)
	return nil
}

func updateSingleOutputStreamRelation(tx sql.Tx, stageId int, stage model.Stage, dto model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.Printf("Start update single relation outputstream for stage with id %v and stream: %v", stageId, dto)

	outputId, err := tryToCreateOutputCodebaseDockerStreamIfDoesNotExist(tx, stage, dto)
	if err != nil {
		return err
	}

	stage.Id = stageId
	if include(applicationsToApprove, dto.CodebaseName) {
		err = setPreviousStageInputImageStream(tx, stage, dto.CodebaseDockerStreamId, *outputId)
	} else {
		err = setOriginalInputImageStream(tx, stage, dto.CodebaseName, *outputId)
	}

	if err != nil {
		log.Printf("Cannot link codebase docker stream for dto: %v", dto)
		return err
	}

	log.Printf("End update single relation outputstream for stage with id %v and stream: %v", stageId, dto)
	return nil
}

func tryToCreateOutputCodebaseDockerStreamIfDoesNotExist(tx sql.Tx, stage model.Stage, dto model.CodebaseDockerStreamReadDTO) (*int, error) {
	ocImageStreamName := fmt.Sprintf("%v-%v-%v-verified", stage.CdPipelineName, stage.Name, dto.CodebaseName)

	var outputId *int

	outputId, err := repository.GetCodebaseDockerStreamId(tx, ocImageStreamName, stage.Tenant)
	if err != nil {
		return nil, fmt.Errorf("cannot get Codebase Docker Stream Id %v: %v", ocImageStreamName, err)
	}

	if outputId == nil {
		log.Println("Output stream has not been created. Try to create it ...")

		branchId, err := repository.GetCodebaseDockerStreamBranchId(tx, dto.CodebaseDockerStreamId, stage.Tenant)
		if err != nil {
			return nil, fmt.Errorf("cannot get branch id by codebase docker stream id %v: %v", dto.CodebaseDockerStreamId, err)
		}

		outputId, err = repository.CreateCodebaseDockerStream(tx, stage.Tenant, branchId, ocImageStreamName)
		if err != nil {
			return nil, fmt.Errorf("cannot create codebase docker stream for dto: %v", dto)
		}
		log.Printf("Id of newly created docker stream is: %v", *outputId)
	}

	return outputId, nil
}

func setPreviousStageInputImageStream(tx sql.Tx, stage model.Stage, inputId int, outputId int) error {
	log.Printf("Previous Stage Input Stream. CD Stage {%v:%v} has InputDockerStream - %v and OutputDockerStream - %v", stage.Id, stage.Name, inputId, outputId)
	return repository.CreateStageCodebaseDockerStream(tx, stage.Tenant, stage.Id, inputId, outputId)
}

func setOriginalInputImageStream(tx sql.Tx, stage model.Stage, codebaseName string, outputId int) error {
	sourceInputStream, err := getOriginalInputImageStream(tx, stage.CdPipelineName, codebaseName, stage.Tenant)
	if err != nil {
		return err
	}

	log.Printf("Source Input Stream. CD Stage {%v:%v} has InputDockerStream - %v and OutputDockerStream - %v", stage.Id, stage.Name, sourceInputStream, outputId)
	return repository.CreateStageCodebaseDockerStream(tx, stage.Tenant, stage.Id, *sourceInputStream, outputId)
}

func getOriginalInputImageStream(tx sql.Tx, cdPipelineName, codebaseName, schemaName string) (*int, error) {
	originalInputStream, err := repository.GetSourceInputStream(tx, cdPipelineName, codebaseName, schemaName)
	if err != nil {
		log.Printf("Couldn't fetch Original Input Stream for %v pipeline and %v codebase: %v", cdPipelineName, codebaseName, err)
		return nil, err
	}
	return originalInputStream, nil
}

func getCDPipelineCR(edpRestClient *rest.RESTClient, crName string, namespace string) (*v1alpha1.CDPipeline, error) {
	log.Printf("Trying to fetch CD Pipeline %v to get Applications To Promote", crName)

	cdPipeline := &v1alpha1.CDPipeline{}
	err := edpRestClient.Get().Namespace(namespace).Resource("cdpipelines").Name(crName).Do().Into(cdPipeline)
	if err != nil {
		log.Printf("An error has occurred while getting CD Pipeline CR from k8s: %s", err)
		return nil, err
	}

	log.Printf("Fetched CD Pipeline: %v", cdPipeline.Spec)

	return cdPipeline, nil
}

func include(applicationsToPromote []string, application string) bool {
	for _, app := range applicationsToPromote {
		if app == application {
			return true
		}
	}
	return false
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
	err = repository.CreateCDPipelineActionLog(tx, *id, *actionLogId, stage.Tenant)
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

func getStageIdOrCreate(tx sql.Tx, edpRestClient *rest.RESTClient, stage model.Stage) (*int, error) {
	log.Printf("Start get stage id or create for stage: %v", stage)
	id, err := repository.GetStageId(tx, stage.Tenant, stage.Name, stage.CdPipelineName)
	if err != nil {
		return nil, err
	}
	if id != nil {
		log.Printf("Stage %v is already presented. Returning id; %v", stage, *id)
		return id, err
	}
	return createStage(tx, edpRestClient, stage)
}

func createStage(tx sql.Tx, edpRestClient *rest.RESTClient, stage model.Stage) (*int, error) {
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

	pipelineCR, err := getCDPipelineCR(edpRestClient, stage.CdPipelineName, stage.Tenant+"-edp-cicd")
	if err != nil {
		return nil, err
	}

	err = createCodebaseDockerStreams(tx, *id, stage, pipelineCR.Spec.ApplicationsToPromote)
	if err != nil {
		return nil, fmt.Errorf("error has occured during the creation docker streams for stage %v in CD Pipeline %v", stage.Name, stage.CdPipelineName)
	}

	err = insertQualityGateRow(tx, *id, stage.QualityGates, stage.Tenant)
	if err != nil {
		return nil, fmt.Errorf("an error has occurred while creating Quality Gate for %v Stage: %v", *id, err)
	}

	return id, nil
}

func insertQualityGateRow(tx sql.Tx, cdStageId int, gates []model.QualityGate, schemaName string) error {
	for _, gate := range gates {
		if gate.QualityGate == "autotests" {
			err := insertAutotestQualityGate(tx, cdStageId, gate, schemaName)
			if err != nil {
				return err
			}

			continue
		}

		err := insertManualQualityGate(tx, cdStageId, gate, schemaName)
		if err != nil {
			return err
		}
	}

	return nil
}

func insertAutotestQualityGate(tx sql.Tx, cdStageId int, gate model.QualityGate, schemaName string) error {
	entityIdsDTO, err := repository.GetCodebaseAndBranchIds(tx, *gate.AutotestName, *gate.BranchName, schemaName)
	if err != nil {
		return err
	}

	_, err = repository.CreateQualityGate(tx, gate.QualityGate, gate.JenkinsStepName, cdStageId, &entityIdsDTO.CodebaseId, &entityIdsDTO.BranchId, schemaName)

	return err
}

func insertManualQualityGate(tx sql.Tx, cdStageId int, gate model.QualityGate, schemaName string) error {
	_, err := repository.CreateQualityGate(tx, gate.QualityGate, gate.JenkinsStepName, cdStageId, nil, nil, schemaName)

	return err
}
