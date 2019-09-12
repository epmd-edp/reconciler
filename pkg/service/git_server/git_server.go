package git_server

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("git-server-service")

type GitServerService struct {
	DB *sql.DB
}

func (s GitServerService) CreateOrUpdateGitServerRecord(gitServer model.GitServer) error {
	log.Info("Start CreateOrUpdateGitServerRecord method", "Git host", gitServer.GitHost)

	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	id, err := repository.SelectGitServer(*txn, gitServer.Name, gitServer.Tenant)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("an error has occurred while fetching Git Server Record %v", gitServer.Name))
	}

	if id != nil {
		log.Info("Start updating Git Server", "record", gitServer.Name)

		err = repository.UpdateGitServer(*txn, id, gitServer.ActionLog.Result == "success", gitServer.Tenant)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("an error has occurred while updating Git Server Record %v", gitServer.Name))
		}
	} else {
		log.Info("Start creating Git Server", "record", gitServer.Name)

		_, err = repository.CreateGitServer(*txn, gitServer.Name, gitServer.GitHost, gitServer.ActionLog.Result == "success", gitServer.Tenant)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("an error has occurred while creating Git Server Record %v", gitServer.GitHost))
		}
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	log.Info("End CreateOrUpdateGitServerRecord method", "Git host", gitServer.GitHost)

	return nil
}
