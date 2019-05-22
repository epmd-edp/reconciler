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
	cdPipelineDto, err := repository.GetCDPipeline(txn, cdPipeline.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if cdPipelineDto != nil {
		return cdPipelineDto, nil
	}
	log.Printf("Record for CD Pipeline %v has not been found", cdPipeline.Name)
	cdPipelineDTO, err := createCDPipeline(txn, cdPipeline, schemaName)
	if err != nil {
		return nil, err
	}

	codebaseBranches, err := getCodebaseBranchesData(cdPipeline, edpRestClient)
	if err != nil {
		log.Printf("An error has occured while getting Codebase Branch from k8s: %s", err)
		return nil, err
	}
	log.Printf("Fetched Codebase Branches for %s pipeline: %s", cdPipeline.Name, codebaseBranches)

	codebaseBranchesId, err := getCodebaseBranchesId(txn, codebaseBranches, schemaName)
	if err != nil {
		log.Printf("An error has occured while getting codebase branch id: %s", err)
		return nil, err
	}

	err = createCDPipelineCodebaseBranch(txn, cdPipelineDTO.Id, codebaseBranchesId, schemaName)
	if err != nil {
		log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
		return nil, err
	}

	return cdPipelineDTO, nil
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
			CodebaseName:    releaseBranchCR.Spec.CodebaseName,
			BranchName: releaseBranchCR.Spec.BranchName,
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

func createCDPipelineCodebaseBranch(txn sql.Tx, cdPipelineId int, codebaseBranchesId []int, schemaName string) error {
	for _, v := range codebaseBranchesId {
		err := repository.CreateCDPipelineCodebaseBranch(txn, cdPipelineId, v, schemaName)
		if err != nil {
			log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
			return err
		}
	}
	return nil
}
