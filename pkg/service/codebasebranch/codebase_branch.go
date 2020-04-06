package codebasebranch

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebasebranch"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	cbs "github.com/epmd-edp/reconciler/v2/pkg/repository/codebasebranch"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("codebase-branch-service")

type CodebaseBranchService struct {
	DB *sql.DB
}

func (s CodebaseBranchService) PutCodebaseBranch(codebaseBranch codebasebranch.CodebaseBranch) error {
	log.V(2).Info("start creation of codebase branch", "name", codebaseBranch.Name)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.Wrap(err, "an error has occurred while opening transaction")
	}
	schemaName := codebaseBranch.Tenant

	id, err := getCodebaseBranchIdOrCreate(txn, codebaseBranch, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while getting Codebase Branch id or create",
			"branch", codebaseBranch.Name)
	}

	if err := updateCodebaseBranch(txn, codebaseBranch, *id, schemaName); err != nil {
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot insert codebaseBranch update %v", codebaseBranch))
	}
	log.V(2).Info("CodebaseBranch has been updated", "name", codebaseBranch.Name)

	log.V(2).Info("start update status of codebase branch...")
	actionLogId, err := repository.CreateActionLog(*txn, codebaseBranch.ActionLog, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during status creation", "name", codebaseBranch.Name)
	}
	log.V(2).Info("ActionLog has been saved into the repository")

	log.V(2).Info("Start update codebase_branch_action status of code branch entity...")
	cbId, err := repository.GetCodebaseId(*txn, codebaseBranch.AppName, schemaName)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during retrieving codebase id", "id", cbId)
	}

	if err := repository.CreateCodebaseAction(*txn, *cbId, *actionLogId, schemaName); err != nil {
		_ = txn.Rollback()
		return errors.Wrap(err, "an error has occurred during codebase_branch_action")
	}
	log.V(2).Info("codebase_action has been updated")

	if err := cbs.UpdateStatusByCodebaseBranchId(*txn, *id, codebaseBranch.Status, codebaseBranch.Tenant); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during the update of codebase branch %v", codebaseBranch.Name)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("Codebase Branch has been saved successfully", "name", codebaseBranch.Name)
	return nil
}

func createCodebaseBranch(txn *sql.Tx, codebaseBranch codebasebranch.CodebaseBranch, schemaName string) (*int, error) {
	log.V(2).Info("start codebase_branch insertion", "name", codebaseBranch.Name)
	var streamId *int = nil
	beId, err := repository.GetCodebaseId(*txn, codebaseBranch.AppName, schemaName)
	if err != nil {
		return nil, err
	}
	if beId == nil {
		return nil, fmt.Errorf("%v codebase record has not been found", codebaseBranch.AppName)
	}

	cbType, err := repository.GetCodebaseTypeById(*txn, *beId, schemaName)
	if err != nil {
		return nil, err
	}

	if *cbType == string(codebase.Application) {
		ocImageStreamName := fmt.Sprintf("%v-%v", codebaseBranch.AppName, codebaseBranch.Name)
		streamId, err = repository.CreateCodebaseDockerStream(*txn, schemaName, nil, ocImageStreamName)
		if err != nil {
			return nil, err
		}
		log.V(2).Info("codebase docker stream has been created", "id", streamId)
	}
	id, err := cbs.CreateCodebaseBranch(*txn, codebaseBranch.Name, *beId,
		codebaseBranch.FromCommit, schemaName, streamId, codebaseBranch.Status, codebaseBranch.Version,
		codebaseBranch.BuildNumber, codebaseBranch.LastSuccessBuild, codebaseBranch.Release)
	if err != nil {
		return nil, err
	}

	if *cbType == string(codebase.Application) {
		if err := repository.UpdateBranchIdCodebaseDockerStream(*txn, *streamId, *id, schemaName); err != nil {
			return nil, err
		}
	}
	log.V(2).Info("end codebase_branch insertion", "name", codebaseBranch.Name)
	return id, nil
}

func getCodebaseBranchIdOrCreate(txn *sql.Tx, codebaseBranch codebasebranch.CodebaseBranch, schemaName string) (*int, error) {
	log.V(2).Info("start retrieving Codebase Branch",
		"codebase", codebaseBranch.AppName, "branch", codebaseBranch.Name)
	id, err := cbs.GetCodebaseBranchId(*txn, codebaseBranch.AppName, codebaseBranch.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.V(2).Info("record for Codebase Branch has not been found", "branch", codebaseBranch.Name)
		return createCodebaseBranch(txn, codebaseBranch, schemaName)
	}
	return id, nil
}

func updateCodebaseBranch(txn *sql.Tx, codebaseBranch codebasebranch.CodebaseBranch, id int, schemaName string) error {
	log.V(2).Info("start updating CodebaseBranch by id", "id", id)
	err := cbs.UpdateCodebaseBranch(*txn, id, codebaseBranch.Version, codebaseBranch.BuildNumber,
		codebaseBranch.LastSuccessBuild, schemaName)
	if err != nil {
		return err
	}
	return nil
}

func (s *CodebaseBranchService) Delete(codebase, branch, schema string) error {
	log.V(2).Info("start deleting codebase branch", "codebase", codebase, "branch", branch)
	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}
	if err := cbs.Delete(*txn, codebase, branch, schema); err != nil {
		return errors.Wrapf(err, "couldn't delete %v codebase branch", codebase)
	}
	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("codebase branch has been deleted", "codebase", codebase, "branch", branch)
	return nil
}
