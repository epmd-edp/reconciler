package jira_server

import (
	"database/sql"
	jiramodel "github.com/epmd-edp/reconciler/v2/pkg/model/jira-server"
	jiraserver "github.com/epmd-edp/reconciler/v2/pkg/repository/jira-server"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("jira-server-service")

type JiraServerService struct {
	DB *sql.DB
}

func (s JiraServerService) PutJiraServer(jira jiramodel.JiraServer) error {
	rl := log.WithValues("jira server name", jira.Name)
	rl.V(2).Info("Start PutJiraServer method")

	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	id, err := jiraserver.SelectJiraServer(*txn, jira.Name, jira.Tenant)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while fetching Jira Server", jira.Name)
	}

	if err := tryToPutJiraServer(txn, id, jira); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred while put Jira Server", jira.Name)
	}

	if err := txn.Commit(); err != nil {
		return err
	}
	log.Info("Jira Server has been created/updated")
	return nil
}

func tryToPutJiraServer(txn *sql.Tx, id *int, jira jiramodel.JiraServer) error {
	if id != nil {
		log.V(2).Info("Start updating Jira Server")
		return jiraserver.UpdateJiraServer(*txn, id, jira.Available, jira.Tenant)
	}
	log.V(2).Info("Start creating Jira Server")
	return jiraserver.CreateJiraServer(*txn, jira.Name, jira.Available, jira.Tenant)
}
