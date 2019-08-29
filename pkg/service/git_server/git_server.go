package git_server

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"reconciler/pkg/model"
	"reconciler/pkg/repository"
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

	id, err := repository.SelectGitServer(*txn, gitServer.GitHost, gitServer.Tenant)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("an error has occurred while fetching Git Server Record %v", gitServer.GitHost))
	}

	if id != nil {
		log.Info("Start updating Git Server", "record", gitServer.GitHost)

		err = repository.UpdateGitServer(*txn, id, gitServer.ActionLog.Result == "success", gitServer.Tenant)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("an error has occurred while updating Git Server Record %v", gitServer.GitHost))
		}
	} else {
		log.Info("Start creating Git Server", "record", gitServer.GitHost)

		_, err = repository.CreateGitServer(*txn, gitServer.GitHost, gitServer.ActionLog.Result == "success", gitServer.Tenant)
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
