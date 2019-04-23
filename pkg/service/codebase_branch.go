package service

import (
	"business-app-reconciler-controller/pkg/model"
	"business-app-reconciler-controller/pkg/repository"
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type CodebaseBranchService struct {
	DB sql.DB
}

func (service CodebaseBranchService) PutCodebaseBranch(codebaseBranch model.CodebaseBranch) error {
	log.Printf("Start creation of codebase branch %v...", codebaseBranch)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create codebase branch %v", codebaseBranch))
	}

	schemaName, err := repository.GetSchema(*txn, codebaseBranch.Tenant)
	if err != nil {
		return err
	}


	tenantName, err := repository.GetCodebaseTenantName(*txn, codebaseBranch.AppName, *schemaName)
	if err != nil {
		log.Printf("Error has occurred while getting tenant name: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot get tenant name for %s application", codebaseBranch.AppName))
	}

	if tenantName == nil {
		errMsg := fmt.Sprintf("Application with name %v has not been found", codebaseBranch.AppName)
		log.Printf(errMsg)
		_ = txn.Rollback()
		return errors.New(errMsg)
	}

	id, err := getCodebaseBranchIdOrCreate(*txn, codebaseBranch, *tenantName, *schemaName)
	if err != nil {
		log.Printf("Error has occurred during get Codebase Branch id or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", codebaseBranch))
	}
	log.Printf("Id of Codebase Branch to be updated: %v", *id)

	isPresent, err := checkCodebaseBranchActionLogDuplicate(*txn, codebaseBranch, *schemaName)
	if err != nil {
		_ = txn.Rollback()
		return err
	}

	if !isPresent {
		log.Println("Start update status of codebase branch...")
		actionLogId, err := repository.CreateCodebaseActionLog(*txn, codebaseBranch.ActionLog, *schemaName)
		if err != nil {
			log.Printf("Error has occurred during status creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot insert status %v", codebaseBranch))
		}
		log.Println("ActionLog has been saved into the repository")

		log.Println("Start update codebase_branch_action status of code branch entity...")
		err = repository.CreateCodebaseBranchAction(*txn, *id, *actionLogId, *schemaName)
		if err != nil {
			log.Printf("Error has occurred during codebase_branch_action creation: %v", err)
			_ = txn.Rollback()
			return errors.New(fmt.Sprintf("cannot create codebase_branch_action entity %v", codebaseBranch))
		}
		log.Println("codebase_action has been updated")
	}

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Println("Codebase Branch has been saved successfully")

	return nil
}

func createCodebaseBranch(txn sql.Tx, codebaseBranch model.CodebaseBranch, tenantName string, schemaName string) (*int, error) {
	log.Println("Start insertion to the codebase_branch table...")
	beId, err := repository.GetCodebaseId(txn, "application", codebaseBranch.AppName, tenantName, schemaName)
	if err != nil {
		return nil, err
	}
	if beId == nil {
		return nil, errors.New(fmt.Sprintf("record for codebase has not been found with %s appName and %s tenantName parameters", codebaseBranch.AppName, tenantName))
	}

	id, err := repository.CreateCodebaseBranch(txn, codebaseBranch.Name, *beId, codebaseBranch.FromCommit, schemaName)
	if err != nil {
		return nil, err
	}

	log.Printf("Id of the newly created codebase branch is %v", *id)
	return id, nil
}

func getCodebaseBranchIdOrCreate(txn sql.Tx, codebaseBranch model.CodebaseBranch, tenantName string, schemaName string) (*int, error) {
	log.Printf("Start retrieving Codebase Branch by name, tenant and appName: %v", codebaseBranch)
	id, err := repository.GetCodebaseBranchId(txn, codebaseBranch.AppName, codebaseBranch.Name, tenantName, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for Codebase Branch %v has not been found", codebaseBranch)
		return createCodebaseBranch(txn, codebaseBranch, tenantName, schemaName)
	}
	return id, nil
}

func checkCodebaseBranchActionLogDuplicate(txn sql.Tx, codebaseBranch model.CodebaseBranch, schemaName string) (bool, error) {
	log.Println("Checks duplicate in action log table")
	lastId, err := repository.GetLastIdCodebaseBranchActionLog(txn, codebaseBranch, schemaName)
	if err != nil {
		log.Printf("Error has occurred while checking on duplicate: %v", err)
		return false, errors.New(fmt.Sprintf("cannot check duplication %v", codebaseBranch))
	}
	if lastId == nil {
		return false, nil
	}
	return true, nil
}
