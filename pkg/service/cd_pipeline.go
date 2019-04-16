package service

import (
	"business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	"business-app-reconciler-controller/pkg/model"
	"business-app-reconciler-controller/pkg/platform"
	"business-app-reconciler-controller/pkg/repository"
	"database/sql"
	"errors"
	"fmt"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"k8s.io/client-go/rest"
	"log"
)

type CdPipelineService struct {
	DB        sql.DB
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

	cdPipelineDb, err := getCDPipelineOrCreate(*txn, edpRestClient, cdPipeline)
	if err != nil {
		log.Printf("Error has occurred during get CD pipeline or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create CD Pipeline entity %v", cdPipeline))
	}
	log.Printf("Id of CD Pipeline to be updated: %v", cdPipelineDb.Id)

	err = updateCDPipelineStatus(*txn, *cdPipelineDb, cdPipeline.Status)
	if err != nil {
		log.Printf("An error has occured while updating CD Pipeline Status for %s pipeline: %s", cdPipelineDb.Name, err)
		return err
	}

	err = updateActionLog(*txn, cdPipeline, cdPipelineDb.Id)
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

func getCDPipelineOrCreate(txn sql.Tx, edpRestClient *rest.RESTClient, cdPipeline model.CDPipeline) (*model.CDPipelineDTO, error) {
	log.Printf("Start retrieving CD Pipeline by name: %v", cdPipeline)
	cdPipelineDto, err := repository.GetCDPipeline(txn, cdPipeline.Name)
	if err != nil {
		return nil, err
	}
	if cdPipelineDto != nil {
		return cdPipelineDto, nil
	}
	log.Printf("Record for CD Pipeline %v has not been found", cdPipeline.Name)
	cdPipelineDTO, err := createCDPipeline(txn, cdPipeline)
	if err != nil {
		return nil, err
	}

	applicationBranches, err := getApplicationBranchesData(cdPipeline, edpRestClient)
	if err != nil {
		log.Printf("An error has occured while getting Application Branch from k8s: %s", err)
		return nil, err
	}
	log.Printf("Fetched Application Branches for %s pipeline: %s", cdPipeline.Name, applicationBranches)

	codebaseBranchesId, err := getCodebaseBranchesId(txn, applicationBranches)
	if err != nil {
		log.Printf("An error has occured while getting codebase branch id: %s", err)
		return nil, err
	}

	err = createCDPipelineCodebaseBranch(txn, cdPipelineDTO.Id, codebaseBranchesId)
	if err != nil {
		log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
		return nil, err
	}

	return cdPipelineDTO, nil
}

func createCDPipeline(txn sql.Tx, cdPipeline model.CDPipeline) (*model.CDPipelineDTO, error) {
	log.Println("Start insertion to the cd_pipeline table...")
	cdPipelineDto, err := repository.CreateCDPipeline(txn, cdPipeline, cdPipeline.Status)
	if err != nil {
		return nil, err
	}

	log.Printf("Id of the newly created CD Pipeline is %v", cdPipelineDto.Id)
	return cdPipelineDto, nil
}

func getApplicationBranchCR(edpRestClient *rest.RESTClient, crName string) (*v1alpha1.ApplicationBranch, error) {
	applicationBranch := &v1alpha1.ApplicationBranch{}
	namespace, err := k8sutil.GetWatchNamespace()
	if err != nil {
		log.Printf("Failed to get watch namespace: %s", err)
		panic(err)
	}
	err = edpRestClient.Get().Namespace(namespace).Resource("applicationbranches").Name(crName).Do().Into(applicationBranch)
	if err != nil {
		log.Printf("An error has occurred while getting Release Branch CR from k8s: %s", err)
		return nil, err
	}
	return applicationBranch, nil
}

func getApplicationBranchesData(cdPipeline model.CDPipeline, edpRestClient *rest.RESTClient) ([]model.ApplicationBranchDTO, error) {
	var applicationBranches []model.ApplicationBranchDTO
	for _, v := range cdPipeline.CodebaseBranch {
		releaseBranchCR, err := getApplicationBranchCR(edpRestClient, v)
		if err != nil {
			return nil, err
		}

		applicationBranches = append(applicationBranches, model.ApplicationBranchDTO{
			AppName:    releaseBranchCR.Spec.AppName,
			BranchName: releaseBranchCR.Spec.BranchName,
		})
	}
	return applicationBranches, nil
}

func updateActionLog(txn sql.Tx, cdPipeline model.CDPipeline, pipelineId int) error {
	log.Println("Start update status of CD Pipeline...")
	actionLogId, err := repository.CreateEventActionLog(txn, cdPipeline.ActionLog)
	if err != nil {
		log.Printf("Error has occurred during status creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot insert status %v", cdPipeline))
	}
	log.Println("ActionLog row has been saved into the repository")

	log.Println("Start update cd_pipeline_codebase_action status of code pipeline entity...")
	err = repository.CreateCDPipelineActionLog(txn, pipelineId, *actionLogId)
	if err != nil {
		log.Printf("Error has occurred during cd_pipeline_action creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create cd_pipeline_action entity %v", cdPipeline))
	}
	log.Println("cd_pipeline_action has been updated")
	return nil
}

func updateCDPipelineStatus(txn sql.Tx, cdPipelineDb model.CDPipelineDTO, status string) error {
	if cdPipelineDb.Status != status {
		log.Printf("Start updating status of %s pipeline to %s", cdPipelineDb.Name, status)
		err := repository.UpdateCDPipelineStatus(txn, cdPipelineDb.Id, status)
		if err != nil {
			return err
		}
	}
	return nil
}

func getCodebaseBranchesId(txn sql.Tx, applicationBranches []model.ApplicationBranchDTO) ([]int, error) {
	var codebaseBranchesId []int
	for _, v := range applicationBranches {
		codebaseBranchId, err := repository.GetCodebaseBranchesId(txn, v)
		if err != nil {
			return nil, err
		}
		codebaseBranchesId = append(codebaseBranchesId, *codebaseBranchId)
	}
	return codebaseBranchesId, nil
}

func createCDPipelineCodebaseBranch(txn sql.Tx, cdPipelineId int, codebaseBranchesId []int) error {
	for _, v := range codebaseBranchesId {
		err := repository.CreateCDPipelineCodebaseBranch(txn, cdPipelineId, v)
		if err != nil {
			log.Printf("An error has occured while inserting CD Pipeline Codebase Branch row: %s", err)
			return err
		}
	}
	return nil
}
