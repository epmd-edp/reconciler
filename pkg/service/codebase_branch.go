package service

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reconciler/pkg/model"
	"reconciler/pkg/repository"
)

type CodebaseBranchService struct {
	DB *sql.DB
}

func (service CodebaseBranchService) PutCodebaseBranch(codebaseBranch model.CodebaseBranch) error {
	log.Printf("Start creation of codebase branch %v...", codebaseBranch)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create codebase branch %v", codebaseBranch))
	}

	schemaName := codebaseBranch.Tenant

	id, err := getCodebaseBranchIdOrCreate(*txn, codebaseBranch, schemaName)
	if err != nil {
		log.Printf("Error has occurred during get Codebase Branch id or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", codebaseBranch))
	}
	log.Printf("Id of Codebase Branch to be updated: %v", *id)

	log.Println("Start update status of codebase branch...")
	actionLogId, err := repository.CreateActionLog(*txn, codebaseBranch.ActionLog, schemaName)
	if err != nil {
		log.Printf("Error has occurred during status creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot insert status %v", codebaseBranch))
	}
	log.Println("ActionLog has been saved into the repository")

	log.Println("Start update codebase_branch_action status of code branch entity...")
	cbId, err := repository.GetCodebaseId(*txn, codebaseBranch.AppName, schemaName)
	if err != nil {
		log.Printf("Error has occurred during retrieving codebase id: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot update codebase branch %v", codebaseBranch))
	}
	err = repository.CreateCodebaseAction(*txn, *cbId, *actionLogId, schemaName)
	if err != nil {
		log.Printf("Error has occurred during codebase_branch_action creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create codebase_branch_action entity %v", codebaseBranch))
	}
	log.Println("codebase_action has been updated")

	err = repository.UpdateStatusByCodebaseBranchId(*txn, *id, codebaseBranch.Status, codebaseBranch.Tenant)
	if err != nil {
		log.Printf("Error has occurred during the update of codebase branch: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create codebase branch with name %v", codebaseBranch.Name))
	}

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Println("Codebase Branch has been saved successfully")

	return nil
}

func createCodebaseBranch(txn sql.Tx, codebaseBranch model.CodebaseBranch, schemaName string) (*int, error) {
	log.Println("Start insertion to the codebase_branch table...")
	var streamId *int = nil

	beId, err := repository.GetCodebaseId(txn, codebaseBranch.AppName, schemaName)
	if err != nil {
		return nil, err
	}
	if beId == nil {
		return nil, errors.New(fmt.Sprintf("record for codebase has not been found with %s appName parameters", codebaseBranch.AppName))
	}

	cbType, err := repository.GetCodebaseTypeById(txn, *beId, schemaName)
	if err != nil {
		return nil, err
	}

	if *cbType == string(model.Application) {
		ocImageStreamName := fmt.Sprintf("%v-%v", codebaseBranch.AppName, codebaseBranch.Name)

		streamId, err = repository.CreateCodebaseDockerStream(txn, schemaName, *beId, ocImageStreamName)
		if err != nil {
			return nil, err
		}

		log.Printf("Id of newly created codebase docker stream: %v", streamId)

	}
	id, err := repository.CreateCodebaseBranch(txn, codebaseBranch.Name, *beId,
		codebaseBranch.FromCommit, schemaName, streamId, codebaseBranch.Status)
	if err != nil {
		return nil, err
	}

	if *cbType == string(model.Application) {
		err := repository.UpdateBranchIdCodebaseDockerStream(txn, *streamId, *id, schemaName)
		if err != nil {
			return nil, err
		}
	}

	log.Printf("Id of the newly created codebase branch is %v", *id)
	return id, nil
}

func getCodebaseBranchIdOrCreate(txn sql.Tx, codebaseBranch model.CodebaseBranch, schemaName string) (*int, error) {
	log.Printf("Start retrieving Codebase Branch by name, tenant and appName: %v", codebaseBranch)
	id, err := repository.GetCodebaseBranchId(txn, codebaseBranch.AppName, codebaseBranch.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for Codebase Branch %v has not been found", codebaseBranch)
		return createCodebaseBranch(txn, codebaseBranch, schemaName)
	}
	return id, nil
}
