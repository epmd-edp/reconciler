package stage

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/codebasebranch"
	sr "github.com/epmd-edp/reconciler/v2/pkg/repository/stage"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cd_stage_service")

type StageService struct {
	DB        *sql.DB
	ClientSet platform.ClientSet
}

//PutStage creates record in DB for Stage.
//The main cases which method do:
//	- checks if stage can be created (checks if previous stage has been added)
//	- update stage status
//	- add record to Action Log for last operation
func (s StageService) PutStage(stage stage.Stage) error {
	log.V(2).Info("start putting stage into db", "name", stage.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.New("error has occurred during opening transaction")
	}

	if !canStageBeCreated(txn, stage) {
		_ = txn.Rollback()
		return fmt.Errorf("previous stage has not been added yet for stage %v", stage.Name)
	}

	id, err := getStageIdOrCreate(txn, s.ClientSet.EDPRestClient, stage)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "cannot create stage %v", stage.Name)
	}

	if err := updateStageStatus(txn, id, stage); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "cannot create stage %v", stage.Name)
	}

	_ = txn.Commit()

	log.Info("stage has been inserted successfully", "name", stage.Name)
	return nil
}

func createCodebaseDockerStreams(tx *sql.Tx, id int, stage stage.Stage, applicationsToApprove []string) error {
	log.V(2).Info("start creating docker streams for stage", "id", id)
	inputDockerStreams, err := getInputDockerStreams(tx, id, stage)
	if err != nil {
		return errors.Wrapf(err, "cannot get list of input docker streams for stage with id : %v", id)
	}
	if err = createOutputStreamsAndLink(tx, id, stage, inputDockerStreams, applicationsToApprove); err != nil {
		return errors.Wrapf(err, "cannot create output streams for stage with id: %v", id)
	}
	log.Info("docker streams have been successfully created for stage", "stage id", id)
	return nil
}

func UpdateSingleStageCodebaseDockerStreamRelations(tx *sql.Tx, id int, stage stage.Stage, applicationsToApprove []string) error {
	log.V(2).Info("start updating docker streams relation for stage", "stage id", id)
	inputDockerStreams, err := getInputDockerStreams(tx, id, stage)
	if err != nil {
		return errors.Wrapf(err, "cannot get list of input docker streams for stage with id : %v", id)
	}
	if err = updateOutputStreamsRelation(tx, id, stage, inputDockerStreams, applicationsToApprove); err != nil {
		return errors.Wrapf(err, "cannot create output streams for stage with id: %v", id)
	}
	log.Info("docker streams relation have been successfully updated for", "stage id", id)
	return nil
}

func createOutputStreamsAndLink(tx *sql.Tx, id int, stage stage.Stage, dtos []model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.V(2).Info("start creating outputstreams and links for stage", "stage id", id)
	for _, stream := range dtos {
		if err := createSingleOutputStreamAndLink(tx, id, stage, stream, applicationsToApprove); err != nil {
			return err
		}
	}
	return nil
}

func updateOutputStreamsRelation(tx *sql.Tx, id int, stage stage.Stage, dtos []model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.V(2).Info("start updating links for stage ", "stage id", id)
	for _, stream := range dtos {
		if err := updateSingleOutputStreamRelation(tx, id, stage, stream, applicationsToApprove); err != nil {
			return err
		}
	}
	return nil
}

func createSingleOutputStreamAndLink(tx *sql.Tx, stageId int, stage stage.Stage, dto model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.V(2).Info("start creating single outputstream and link for stage", "stage id", stageId, "stream id", dto.CodebaseDockerStreamId)
	ocImageStreamName := fmt.Sprintf("%v-%v-%v-verified", stage.CdPipelineName, stage.Name, dto.CodebaseName)
	branchId, err := repository.GetCodebaseDockerStreamBranchId(*tx, dto.CodebaseDockerStreamId, stage.Tenant)
	if err != nil {
		return errors.Wrapf(err, "cannot get branch id by codebase docker stream id %v", dto.CodebaseDockerStreamId)
	}

	outputId, err := repository.CreateCodebaseDockerStream(*tx, stage.Tenant, branchId, ocImageStreamName)
	if err != nil {
		return errors.Wrap(err, "cannot create codebase docker stream")
	}
	log.Info("docker stream was created", "id", *outputId)

	stage.Id = stageId
	if include(applicationsToApprove, dto.CodebaseName) {
		err = setPreviousStageInputImageStream(tx, stage, dto.CodebaseDockerStreamId, *outputId)
	} else {
		err = setOriginalInputImageStream(tx, stage, dto.CodebaseName, *outputId)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot link codebase docker stream", "id", dto.CodebaseDockerStreamId)
	}
	return nil
}

func updateSingleOutputStreamRelation(tx *sql.Tx, stageId int, stage stage.Stage, dto model.CodebaseDockerStreamReadDTO, applicationsToApprove []string) error {
	log.V(2).Info("start updating single relation outputstream for stage", "stage id", stageId)
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
		return errors.Wrap(err, "cannot link codebase docker stream")
	}
	return nil
}

func tryToCreateOutputCodebaseDockerStreamIfDoesNotExist(tx *sql.Tx, stage stage.Stage, dto model.CodebaseDockerStreamReadDTO) (*int, error) {
	ocImageStreamName := fmt.Sprintf("%v-%v-%v-verified", stage.CdPipelineName, stage.Name, dto.CodebaseName)

	var outputId *int

	outputId, err := repository.GetCodebaseDockerStreamId(*tx, ocImageStreamName, stage.Tenant)
	if err != nil {
		return nil, fmt.Errorf("cannot get Codebase Docker Stream Id %v: %v", ocImageStreamName, err)
	}

	if outputId == nil {
		log.V(2).Info("output stream has not been created. Try to create it ...")

		branchId, err := repository.GetCodebaseDockerStreamBranchId(*tx, dto.CodebaseDockerStreamId, stage.Tenant)
		if err != nil {
			return nil, fmt.Errorf("cannot get branch id by codebase docker stream id %v: %v", dto.CodebaseDockerStreamId, err)
		}

		outputId, err = repository.CreateCodebaseDockerStream(*tx, stage.Tenant, branchId, ocImageStreamName)
		if err != nil {
			return nil, fmt.Errorf("cannot create codebase docker stream for dto: %v", dto)
		}
		log.Info("docker stream was created", "id", *outputId)
	}

	return outputId, nil
}

func setPreviousStageInputImageStream(tx *sql.Tx, stage stage.Stage, inputId int, outputId int) error {
	log.V(2).Info("previous Stage Input Stream", "stage", stage.Id, "input", inputId, "output", outputId)
	return repository.CreateStageCodebaseDockerStream(*tx, stage.Tenant, stage.Id, inputId, outputId)
}

func setOriginalInputImageStream(tx *sql.Tx, stage stage.Stage, codebaseName string, outputId int) error {
	sourceInputStream, err := getOriginalInputImageStream(tx, stage.CdPipelineName, codebaseName, stage.Tenant)
	if err != nil {
		return err
	}
	log.V(2).Info("source Stage Input Stream", "stage", stage.Id, "input", sourceInputStream, "output", outputId)
	return repository.CreateStageCodebaseDockerStream(*tx, stage.Tenant, stage.Id, *sourceInputStream, outputId)
}

func getOriginalInputImageStream(tx *sql.Tx, cdPipelineName, codebaseName, schemaName string) (*int, error) {
	originalInputStream, err := repository.GetSourceInputStream(*tx, cdPipelineName, codebaseName, schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't fetch Original Input Stream for pipeline", "pipe", cdPipelineName, "codebase", codebaseName)
	}
	return originalInputStream, nil
}

func GetCDPipelineCR(edpRestClient *rest.RESTClient, crName string, namespace string) (*v1alpha1.CDPipeline, error) {
	log.V(2).Info("trying to fetch CD Pipeline to get Applications To Promote", "pipe name", crName)
	cdPipeline := &v1alpha1.CDPipeline{}
	err := edpRestClient.Get().Namespace(namespace).Resource("cdpipelines").Name(crName).Do().Into(cdPipeline)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has occurred while getting CD Pipeline CR from cluster")
	}
	log.V(2).Info("CD Pipeline wsa fetched", "name", cdPipeline.Spec.Name)
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

func getInputDockerStreams(tx *sql.Tx, id int, stage stage.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.V(2).Info("start reading input docker streams for stage", "stage id", id)
	if stage.Order == 0 {
		return getInputDockerStreamsForFirstStage(tx, id, stage)
	}
	return getInputDockerStreamsForArbitraryStage(tx, id, stage)
}

func getInputDockerStreamsForArbitraryStage(tx *sql.Tx, id int, stage stage.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.V(2).Info("start reading input docker streams for the arbitrary stage with id: %v", id)
	streams, err := repository.GetDockerStreamsByPipelineNameAndStageOrder(*tx, stage.Tenant, stage.CdPipelineName, stage.Order-1)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has been occurred during the read docker streams",
			"pipeline name", stage.CdPipelineName, "stage order", stage.Order-1)
	}
	log.V(2).Info("streams have been successfully retrieved", "streams", streams)
	return streams, nil
}

func getInputDockerStreamsForFirstStage(tx *sql.Tx, id int, stage stage.Stage) ([]model.CodebaseDockerStreamReadDTO, error) {
	log.V(2).Info("start reading input docker streams for the first stage", "stage id", id)
	streams, err := repository.GetDockerStreamsByPipelineName(*tx, stage.Tenant, stage.CdPipelineName)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has been occurred during the read docker streams",
			"pipeline name", stage.CdPipelineName)
	}
	log.V(2).Info("streams have been successfully retrieved", "streams", streams)
	return streams, nil
}

func canStageBeCreated(tx *sql.Tx, stage stage.Stage) bool {
	if stage.Order == 0 {
		log.V(2).Info("stage is the first in the chain. Returning true..", "name", stage.Name)
		return true
	}
	return prevStageAdded(tx, stage)
}

func prevStageAdded(tx *sql.Tx, stage stage.Stage) bool {
	log.V(2).Info("check previous stage fot stage", "name", stage.Name)
	stageId, err := sr.GetStageIdByPipelineNameAndOrder(*tx, stage.Tenant, stage.CdPipelineName, stage.Order-1)
	if err != nil {
		log.Error(err, "an error has been occurred while retrieving prev stage id : %v", stageId)
		return false
	}
	if stageId == nil {
		log.V(2).Info("previous stage for stage has not been added. Returning false", "target stage", stage.Name)
		return false
	}
	log.V(2).Info("previous stage id", "id", stageId)
	return true
}

func updateStageStatus(tx *sql.Tx, id *int, stage stage.Stage) error {
	log.V(2).Info("start updating status for stage", "id", *id, "status", stage.Status)
	err := sr.UpdateStageStatus(*tx, stage.Tenant, *id, stage.Status)
	if err != nil {
		return errors.Wrapf(err, "an error has been occurred while updating stage status: %v", stage.Name)
	}
	log.V(2).Info("status has been updated", "id", *id, "status", stage.Status)
	return nil
}

func getStageIdOrCreate(tx *sql.Tx, edpRestClient *rest.RESTClient, stage stage.Stage) (*int, error) {
	id, err := sr.GetStageId(*tx, stage.Tenant, stage.Name, stage.CdPipelineName)
	if err != nil {
		return nil, err
	}
	if id != nil {
		log.V(2).Info("stage is already presented. Returning id", "name", stage, "id", *id)
		return id, err
	}
	return createStage(tx, edpRestClient, stage)
}

func createStage(tx *sql.Tx, edpRestClient *rest.RESTClient, stage stage.Stage) (*int, error) {
	log.V(2).Info("start creating stage in db", "name", stage.Name)
	cdPipeline, err := repository.GetCDPipeline(*tx, stage.CdPipelineName, stage.Tenant)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has been occurred while reading cd pipeline %v", stage.CdPipelineName)
	}
	if cdPipeline == nil {
		return nil, fmt.Errorf("record for cd pipeline with name %v has not been found", stage.CdPipelineName)
	}

	if err := setLibraryIdOrDoNothing(tx, &stage.Source, stage.Tenant); err != nil {
		return nil, err
	}

	id, err := sr.CreateStage(*tx, stage, cdPipeline.Id)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create stage id db")
	}

	pipelineCR, err := GetCDPipelineCR(edpRestClient, stage.CdPipelineName, stage.Namespace)
	if err != nil {
		return nil, err
	}

	if err = createCodebaseDockerStreams(tx, *id, stage, pipelineCR.Spec.ApplicationsToPromote); err != nil {
		return nil, errors.Wrapf(err, "couldn't create docker stream for stage %v in CD Pipeline", stage.Name, stage.CdPipelineName)
	}

	if err = insertQualityGateRow(tx, *id, stage.QualityGates, stage.Tenant); err != nil {
		return nil, errors.Wrapf(err, "couldn't create quality gate for stage %v", *id)
	}
	log.Info("stage has been created in db", "id", *id)
	return id, nil
}

func setLibraryIdOrDoNothing(txn *sql.Tx, source *stage.Source, schemaName string) error {
	if source.Type == "default" {
		return nil
	}

	id, err := repository.GetCodebaseId(*txn, source.Library.Name, schemaName)
	if err != nil {
		return errors.Wrapf(err, "an error has occurred while getting library id by %v codebase name",
			source.Library.Name)
	}
	if id == nil {
		return fmt.Errorf("library wasn't found by %v name", source.Library.Name)
	}
	source.Library.Id = id

	bid, err := codebasebranch.GetCodebaseBranchId(*txn, source.Library.Name, source.Library.Branch, schemaName)
	if err != nil {
		return errors.Wrapf(err, "an error has occurred while getting library branch id by %v codebase name and %v branch",
			source.Library.Name, source.Library.Branch)
	}
	if bid == nil {
		return fmt.Errorf("branch wasn't found by %v name", source.Library.Branch)
	}
	source.Library.BranchId = bid

	return nil
}

func insertQualityGateRow(tx *sql.Tx, cdStageId int, gates []stage.QualityGate, schemaName string) error {
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

func insertAutotestQualityGate(tx *sql.Tx, cdStageId int, gate stage.QualityGate, schemaName string) error {
	entityIdsDTO, err := sr.GetCodebaseAndBranchIds(*tx, *gate.AutotestName, *gate.BranchName, schemaName)
	if err != nil {
		return err
	}

	_, err = sr.CreateQualityGate(*tx, gate.QualityGate, gate.JenkinsStepName, cdStageId, &entityIdsDTO.CodebaseId, &entityIdsDTO.BranchId, schemaName)

	return err
}

func insertManualQualityGate(tx *sql.Tx, cdStageId int, gate stage.QualityGate, schemaName string) error {
	_, err := sr.CreateQualityGate(*tx, gate.QualityGate, gate.JenkinsStepName, cdStageId, nil, nil, schemaName)

	return err
}

func (s StageService) DeleteCDStage(pipeName, stageName, schema string) error {
	log.V(2).Info("start deleting cd stage", "pipe name", pipeName, "name", stageName)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.New("error has occurred during opening transaction")
	}

	id, err := sr.SelectCodebaseDockerStreamId(*txn, pipeName, stageName, schema)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't get codebase docker stream id by cd stage %v for cd pipeline", stageName, pipeName)
	}

	if id == nil {
		_ = txn.Rollback()
		log.V(2).Info("docker stream has been deleted", "pipe", pipeName, "stage", stageName)
		return nil
	}

	if err := sr.DeleteCodebaseDockerStream(*txn, *id, schema); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't delete codebase docker stream with %v id", *id)
	}

	if err := sr.DeleteCDStage(*txn, pipeName, stageName, schema); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't delete cd stage %v for cd pipeline", stageName, pipeName)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("cd stage was deleted", "pipe name", pipeName, "name", stageName)
	return nil
}
