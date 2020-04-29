package service

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/jenkins-slave"
	jiraserver "github.com/epmd-edp/reconciler/v2/pkg/repository/jira-server"
	jp "github.com/epmd-edp/reconciler/v2/pkg/repository/job-provisioning"
	"github.com/pkg/errors"
	"log"
)

type BEService struct {
	DB *sql.DB
}

func (service BEService) PutBE(be codebase.Codebase) error {
	log.Printf("Start creation of business entity %v...", be)
	log.Println("Start transaction...")
	txn, err := service.DB.Begin()
	if err != nil {
		log.Printf("Error has occurred during opening transaction: %v", err)
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}

	schemaName := be.Tenant

	id, err := getBeIdOrCreate(txn, be, schemaName)
	if err != nil {
		log.Printf("Error has occurred during get BE id or create: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create business entity %v", be))
	}
	log.Printf("Id of BE to be updated: %v", *id)

	log.Println("Start update status of codebase...")
	codebaseActionId, err := repository.CreateActionLog(*txn, be.ActionLog, schemaName)
	if err != nil {
		log.Printf("Error has occurred during status creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot insert status %v", be))
	}
	log.Println("ActionLog has been saved into the repository")

	log.Println("Start update codebase_action status of codebase...")
	err = repository.CreateCodebaseAction(*txn, *id, *codebaseActionId, schemaName)
	if err != nil {
		log.Printf("Error has occurred during codebase_action creation: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create codebase_action entity %v", be))
	}
	log.Println("codebase_action has been updated")

	err = repository.UpdateStatusByCodebaseId(*txn, *id, be.Status, be.Tenant)
	if err != nil {
		log.Printf("Error has occurred during the update of codebase: %v", err)
		_ = txn.Rollback()
		return errors.New(fmt.Sprintf("cannot create codebase with name %v", be.Name))
	}

	err = txn.Commit()
	if err != nil {
		log.Printf("An error has occurred while ending transaction: %s", err)
		return err
	}

	log.Println("Business entity has been saved successfully")

	return nil
}

func getBeIdOrCreate(txn *sql.Tx, be codebase.Codebase, schemaName string) (*int, error) {
	log.Printf("Start retrieving BE by name, tenant and type: %v", be)
	id, err := repository.GetCodebaseId(*txn, be.Name, schemaName)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for BE %v has not been found", be)
		return createBE(txn, be, schemaName)
	}
	return id, nil
}

func createBE(txn *sql.Tx, c codebase.Codebase, schemaName string) (*int, error) {
	log.Println("Start insertion in the repository business entity...")

	serverId, err := getGitServerId(txn, c.GitServer, schemaName)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot get git server: %v", c.GitServer))
	}
	log.Printf("GitServer is fetched: %v", serverId)
	if serverId == nil {
		return nil, fmt.Errorf("git server has not been found for %v", c.GitServer)
	}
	c.GitServerId = serverId

	id, err := getJiraServerId(txn, c.JiraServer, schemaName)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't get Jira server id by % name", *c.JiraServer)
	}
	if id != nil {
		c.JiraServerId = id
	}

	if c.JenkinsSlave != "" {
		jsId, err := jenkins_slave.SelectJenkinsSlave(*txn, c.JenkinsSlave, schemaName)
		if err != nil || jsId == nil {
			return nil, errors.New(fmt.Sprintf("couldn't get jenkins slave id: %v", c.JenkinsSlave))
		}
		log.Printf("Jenkins Slave Id for %v codebase is %v", c.Name, *jsId)

		c.JenkinsSlaveId = jsId
	}

	if c.JobProvisioning != "" {
		jpId, err := jp.SelectJobProvision(*txn, c.JobProvisioning, schemaName)
		if err != nil || jpId == nil {
			return nil, errors.New(fmt.Sprintf("couldn't get job provisioning id: %v", c.JobProvisioning))
		}

		log.Printf("Job Probisioning Id for %v codebase is %v", c.Name, *jpId)

		c.JobProvisioningId = jpId
	}

	id, err = repository.CreateCodebase(*txn, c, schemaName)
	if err != nil {
		log.Printf("Error has occurred during business entity creation: %v", err)
		return nil, errors.New(fmt.Sprintf("cannot create business entity %v", c))
	}
	log.Printf("Id of the newly created business entity is %v", *id)
	return id, nil
}

func getGitServerId(txn *sql.Tx, gitServerName string, schemaName string) (*int, error) {
	log.Println("Fetching GitServer Id to set relation into codebase...")

	serverId, err := repository.SelectGitServer(*txn, gitServerName, schemaName)
	if err != nil {
		return nil, err
	}

	return serverId, nil
}

func getJiraServerId(txn *sql.Tx, name *string, schemaName string) (*int, error) {
	if name == nil {
		return nil, nil
	}
	log.Printf("Fetching JiraServer Id by %v name to set relation into codebase...", name)

	id, err := jiraserver.SelectJiraServer(*txn, *name, schemaName)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func (s BEService) Delete(name, schema string) error {
	log.Printf("start deleting %v codebase", name)
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.Wrapf(err, "couldn't open transaction while deleting codebase %v", name)
	}

	if err := repository.Delete(*txn, name, schema); err != nil {
		return errors.Wrapf(err, "couldn't delete codebase %v", name)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrapf(err, "couldn't close transaction while deleting codebase %v", name)
	}
	log.Printf("end deleting %v codebase", name)
	return nil
}
