package cd_pipeline

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/model/cdpipeline"
	"github.com/epmd-edp/reconciler/v2/pkg/model/stage"
	"github.com/epmd-edp/reconciler/v2/pkg/platform"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	sr "github.com/epmd-edp/reconciler/v2/pkg/repository/stage"
	"github.com/epmd-edp/reconciler/v2/pkg/service"
	service_stage "github.com/epmd-edp/reconciler/v2/pkg/service/stage"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sort"
)

var log = logf.Log.WithName("cd_pipeline_service")

type CdPipelineService struct {
	DB        *sql.DB
	ClientSet platform.ClientSet
}

func (s CdPipelineService) PutCDPipeline(cdPipeline cdpipeline.CDPipeline) error {
	log.V(2).Info("start CD Pipeline creation", "name", cdPipeline.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.New("an error has occurred while opening transaction")
	}
	schemaName := cdPipeline.Tenant

	cdPipelineDb, err := s.getCDPipelineOrCreate(txn, cdPipeline, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't get/create cd pipeline %v", cdPipeline.Name)
	}
	log.Info("Id of CD Pipeline to be updated: %v", cdPipelineDb.Id)

	if err := updateCDPipelineStatus(txn, *cdPipelineDb, cdPipeline.Status, schemaName); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while updating %v CD Pipeline Status", cdPipelineDb.Name)
	}

	if err := updateActionLog(txn, cdPipeline, cdPipelineDb.Id, schemaName); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while updating CD Pipelin %ve Action Event Log", cdPipeline.Name)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrap(err, "an error has occurred while closing transaction")
	}
	log.Info("CD Pipeline has been saved successfully", "name", cdPipelineDb.Name)
	return nil
}

func (s CdPipelineService) getCDPipelineOrCreate(txn *sql.Tx, cdPipeline cdpipeline.CDPipeline, schemaName string) (*model.CDPipelineDTO, error) {
	log.V(2).Info("start retrieving CD Pipeline", "name", cdPipeline.Name)
	cdPipelineReadModel, err := repository.GetCDPipeline(*txn, cdPipeline.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if cdPipelineReadModel != nil {
		if err := repository.DeleteCDPipelineDockerStreams(*txn, cdPipelineReadModel.Id, schemaName); err != nil {
			return nil, errors.Wrap(err, "an error has occurred while deleting pipeline's docker streams")
		}

		if err := createCDPipelineDockerStream(txn, cdPipelineReadModel.Id, cdPipeline.InputDockerStreams, schemaName); err != nil {
			return nil, err
		}

		stages, err := getStages(txn, cdPipelineReadModel.Name, schemaName)
		if err != nil {
			return nil, err
		}

		sort.SliceStable(stages, func(i, j int) bool {
			return stages[i].Order < stages[j].Order
		})

		for i := 0; i < len(stages); i++ {
			stages[i].Namespace = cdPipeline.Namespace
		}

		if err := s.updateStageCodebaseDockerStream(txn, stages, cdPipelineReadModel.Name, schemaName); err != nil {
			return nil, err
		}

		if err := updateApplicationsToPromote(txn, cdPipelineReadModel.Id, cdPipeline.ApplicationsToPromote, schemaName); err != nil {
			return nil, err
		}

		return cdPipelineReadModel, nil
	}
	log.V(2).Info("record for CD Pipeline has not been found", "name", cdPipeline.Name)

	cdPipelineDTO, err := createCDPipeline(txn, cdPipeline, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return nil, err
	}

	if err := createCDPipelineDockerStream(txn, cdPipelineDTO.Id, cdPipeline.InputDockerStreams, schemaName); err != nil {
		_ = txn.Rollback()
		return nil, err
	}

	if cdPipeline.ThirdPartyServices != nil && len(cdPipeline.ThirdPartyServices) != 0 {
		log.V(2).Info("try to create records in ThirdPartyServices", "values", cdPipeline.ThirdPartyServices)
		servicesId, err := service.GetServicesId(txn, cdPipeline.ThirdPartyServices, schemaName)
		if err != nil {
			_ = txn.Rollback()
			return nil, errors.Wrap(err, "an error has occurred while getting services id:")
		}

		if err := createCDPipelineThirdPartyService(txn, cdPipelineDTO.Id, servicesId, schemaName); err != nil {
			_ = txn.Rollback()
			return nil, errors.Wrap(err, "an error has occurred while inserting record into cd_pipeline_third_party_service")
		}
	}

	if err := createApplicationToPromoteRow(txn, cdPipelineDTO.Id, cdPipeline.ApplicationsToPromote, schemaName); err != nil {
		_ = txn.Rollback()
		return nil, errors.Wrap(err, "an error has occurred while inserting record into applications_to_promote")
	}
	return cdPipelineDTO, nil
}

func updateApplicationsToPromote(tx *sql.Tx, cdPipelineId int, applicationsToPromote []string, schemaName string) error {
	if err := repository.RemoveApplicationsToPromote(*tx, cdPipelineId, schemaName); err != nil {
		return errors.Wrapf(err, "an error has occurred while removing Application To Promote records for Stage",
			"pipe id", cdPipelineId)
	}
	if err := createApplicationToPromoteRow(tx, cdPipelineId, applicationsToPromote, schemaName); err != nil {
		return fmt.Errorf("an error has occurred while creating Application To Promote record for %v Stage: %v", cdPipelineId, err)
	}
	return nil
}

func createApplicationToPromoteRow(txn *sql.Tx, cdPipelineId int, applicationsToPromote []string, schemaName string) error {
	log.V(2).Info("try to create record in ApplicationToPromote table", "applicationsToPromote", applicationsToPromote)
	for _, appToPromote := range applicationsToPromote {
		id, err := repository.GetApplicationId(*txn, appToPromote, schemaName)
		if err != nil {
			return err
		}

		if err := repository.CreateApplicationsToPromote(*txn, cdPipelineId, *id, schemaName); err != nil {
			return err
		}
	}

	return nil
}

func (s CdPipelineService) updateStageCodebaseDockerStreamRelations(txn *sql.Tx, stages []stage.Stage, pipelineName string, schemaName string) error {
	log.V(2).Info("try to update Stage Codebase Docker Streams relations for stages", "stages", stages)
	for i := range stages {
		stages[i].Tenant = schemaName
		stages[i].CdPipelineName = pipelineName

		pipelineCR, err := service_stage.GetCDPipelineCR(s.ClientSet.EDPRestClient, stages[i].CdPipelineName, stages[i].Namespace)
		if err != nil {
			return err
		}

		if err := service_stage.UpdateSingleStageCodebaseDockerStreamRelations(txn, stages[i].Id, stages[i], pipelineCR.Spec.ApplicationsToPromote); err != nil {
			return err
		}
	}
	log.V(2).Info("relations have been updated for pipeline", "name", pipelineName)
	return nil
}

func getStages(txn *sql.Tx, cdPipelineName string, schemaName string) ([]stage.Stage, error) {
	stages, err := sr.GetStages(*txn, cdPipelineName, schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "an error has occurred while getting Stages for CD Pipeline %v", cdPipelineName)
	}
	log.V(2).Info("stages have been fetched", "pipe", cdPipelineName, "stages", stages)
	return stages, nil
}

func deleteStageCodebaseDockerStream(txn *sql.Tx, stages []stage.Stage, schemaName string) ([]int, error) {
	var outputStreamIdsToRemove []int
	var stagesToLog []string

	for _, stage := range stages {
		outputStreamIds, err := repository.DeleteStageCodebaseDockerStream(*txn, stage.Id, schemaName)
		outputStreamIdsToRemove = append(outputStreamIdsToRemove, outputStreamIds...)
		if err != nil {
			return nil, errors.Wrap(err, "an error has occurred while deleting stage codebase docker stream row")
		}
		stagesToLog = append(stagesToLog, stage.Name)
	}
	log.V(2).Info("Output Stream Ids to delete have been collected", "id's", outputStreamIdsToRemove)
	return outputStreamIdsToRemove, nil
}

func (s CdPipelineService) updateStageCodebaseDockerStream(txn *sql.Tx, stages []stage.Stage, pipelineName string, schemaName string) error {
	if stages == nil {
		log.V(2).Info("There're no stages for CD Pipeline. Updating of Codebase Docker stream will not be executed.",
			"pipe", pipelineName)
		return nil
	}

	if _, err := deleteStageCodebaseDockerStream(txn, stages, schemaName); err != nil {
		return err
	}

	if err := s.updateStageCodebaseDockerStreamRelations(txn, stages, pipelineName, schemaName); err != nil {
		return err
	}

	return nil
}

func createCDPipeline(txn *sql.Tx, cdPipeline cdpipeline.CDPipeline, schemaName string) (*model.CDPipelineDTO, error) {
	log.V(2).Info("start insertion cd_pipeline to table", "name", cdPipeline.Name)
	cdPipelineDto, err := repository.CreateCDPipeline(*txn, cdPipeline, cdPipeline.Status, schemaName)
	if err != nil {
		return nil, err
	}
	log.Info("cd pipeline has been created", "id", cdPipelineDto.Id)
	return cdPipelineDto, nil
}

func updateActionLog(txn *sql.Tx, cdPipeline cdpipeline.CDPipeline, pipelineId int, schemaName string) error {
	log.V(2).Info("start updating status of CD Pipeline", "name", cdPipeline.Name)
	actionLogId, err := repository.CreateEventActionLog(*txn, cdPipeline.ActionLog, schemaName)
	if err != nil {
		return errors.Wrapf(err, "cannot insert status %v", cdPipeline)
	}

	log.V(2).Info("start updating cd_pipeline_codebase_action status of code pipeline entity...")
	if err := repository.CreateCDPipelineActionLog(*txn, pipelineId, *actionLogId, schemaName); err != nil {
		return errors.Wrapf(err, "cannot create cd_pipeline_action entity %v", cdPipeline)
	}
	log.Info("cd_pipeline_action has been updated")
	return nil
}

func updateCDPipelineStatus(txn *sql.Tx, cdPipelineDb model.CDPipelineDTO, status string, schemaName string) error {
	if cdPipelineDb.Status != status {
		log.V(2).Info("start updating status of cd pipeline",
			"pipe name", cdPipelineDb.Name, "status", status)
		if err := repository.UpdateCDPipelineStatus(*txn, cdPipelineDb.Id, status, schemaName); err != nil {
			return err
		}
	}
	return nil
}

func createCDPipelineThirdPartyService(txn *sql.Tx, cdPipelineId int, servicesId []int, schemaName string) error {
	for _, serviceId := range servicesId {
		err := repository.CreateCDPipelineThirdPartyService(*txn, cdPipelineId, serviceId, schemaName)
		if err != nil {
			return err
		}
	}
	return nil
}

func createCDPipelineDockerStream(txn *sql.Tx, cdPipelineId int, dockerStreams []string, schemaName string) error {
	var dockerStreamIds []int
	for _, dockerStream := range dockerStreams {
		id, err := repository.GetCodebaseDockerStreamId(*txn, dockerStream, schemaName)
		if err != nil {
			return errors.Wrapf(err, "an error has occurred while getting id of docker stream %v", dockerStream)
		}
		if id == nil {
			return fmt.Errorf("cannot find docker stream by name: %v in the schema: %v", dockerStream, schemaName)
		}
		dockerStreamIds = append(dockerStreamIds, *id)
	}

	if err := insertCDPipelineDockerStream(txn, cdPipelineId, dockerStreamIds, schemaName); err != nil {
		return err
	}

	return nil
}

func insertCDPipelineDockerStream(txn *sql.Tx, cdPipelineId int, dockerStreams []int, schemaName string) error {
	for _, id := range dockerStreams {
		if err := repository.CreateCDPipelineDockerStream(*txn, cdPipelineId, id, schemaName); err != nil {
			return errors.Wrapf(err, "an error has occurred while inserting CD Pipeline Docker Stream row %v", id)
		}
	}
	return nil
}

func (s CdPipelineService) DeleteCDPipeline(pipeName, schema string) error {
	log.V(2).Info("start deleting cd pipeline", "name", pipeName)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	if err := sr.DeleteCodebaseDockerStreams(*txn, pipeName, schema); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't delete codebase docker streams for %v cd pipeline", pipeName)
	}

	if err := repository.DeleteCDPipeline(*txn, pipeName, schema); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "couldn't delete cd pipeline %v", pipeName)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("cd pipeline has been deleted", "pipe name", pipeName)
	return nil
}
