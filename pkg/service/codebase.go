package service

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/service/codebaseperfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfdatasource"
	"github.com/epmd-edp/reconciler/v2/pkg/service/perfserver"
	"log"

	"github.com/epmd-edp/reconciler/v2/pkg/model/codebase"
	"github.com/epmd-edp/reconciler/v2/pkg/repository"
	"github.com/epmd-edp/reconciler/v2/pkg/repository/jenkins-slave"
	jiraserver "github.com/epmd-edp/reconciler/v2/pkg/repository/jira-server"
	jp "github.com/epmd-edp/reconciler/v2/pkg/repository/job-provisioning"
	"github.com/pkg/errors"
)

type CodebaseService struct {
	DB                *sql.DB
	DataSourceService perfdatasource.PerfDataSourceService
	PerfService       perfserver.PerfServerService
	CodebaseDsService codebaseperfdatasource.CodebasePerfDataSourceService
}

func (s CodebaseService) PutCodebase(c codebase.Codebase) error {
	log.Printf("Start creation of business entity %v...", c)
	log.Println("Start transaction...")
	txn, err := s.DB.Begin()
	if err != nil {
		return errors.Wrapf(err, "an error has occurred during opening transaction: %v", c.Name)
	}

	id, err := s.putCodebase(txn, c, c.Tenant)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during get Codebase id or create: %v", c.Name)
	}
	log.Printf("Id of BE to be updated: %v", *id)

	log.Println("Start update status of codebase...")
	codebaseActionId, err := repository.CreateActionLog(*txn, c.ActionLog, c.Tenant)
	if err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during status creation: %v", c.Name)
	}
	log.Println("ActionLog has been saved into the repository")

	log.Println("Start update codebase_action status of codebase...")
	if err := repository.CreateCodebaseAction(*txn, *id, *codebaseActionId, c.Tenant); err != nil {
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during codebase_action creation: %v", c.Name)
	}
	log.Println("codebase_action has been updated")

	if err := repository.UpdateStatusByCodebaseId(*txn, *id, c.Status, c.Tenant); err != nil {
		log.Printf("Error has occurred during the update of codebase: %v", err)
		_ = txn.Rollback()
		return errors.Wrapf(err, "an error has occurred during the update of codebase: %v", c.Name)
	}

	if err := txn.Commit(); err != nil {
		return errors.Wrapf(err, "An error has occurred while ending transaction: %v", c.Name)
	}
	log.Printf("Codebase %v has been saved successfully", c.Name)

	if err := s.DataSourceService.InsertPerfDataSources(c.Perf, c.Tenant); err != nil {
		return errors.Wrap(err, "an error has occurred during filling perf data source table")
	}

	if err := s.CodebaseDsService.InsertCodebasePerfDataSources(*id, c.Perf, c.Tenant); err != nil {
		return errors.Wrapf(err, "couldn't create CodebasePerfDataSource record. codebase id %v", id)
	}
	return nil
}

func (s CodebaseService) putCodebase(txn *sql.Tx, c codebase.Codebase, schema string) (*int, error) {
	log.Printf("Start retrieving Codebase by name, tenant and type: %v", c)
	id, err := repository.GetCodebaseId(*txn, c.Name, schema)
	if err != nil {
		return nil, err
	}
	if id == nil {
		log.Printf("Record for Codebase %v has not been found", c)
		return s.createBE(txn, c, schema)
	}
	return id, updateCodebase(txn, c, schema)
}

func updateCodebase(txn *sql.Tx, c codebase.Codebase, schema string) error {
	log.Printf("start updating codebase %v", c.Name)
	if err := repository.Update(*txn, c, schema); err != nil {
		return errors.Wrapf(err, "couldn't update codebase %v", c.Name)
	}
	log.Printf("codebase %v has been updated", c.Name)
	return nil
}

func (s CodebaseService) createBE(txn *sql.Tx, c codebase.Codebase, schemaName string) (*int, error) {
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
		return nil, errors.Wrapf(err, "couldn't get Jira server id by %v name", *c.JiraServer)
	}
	if id != nil {
		c.JiraServerId = id
	}

	if c.JenkinsSlave != nil && *c.JenkinsSlave != "" {
		jsId, err := jenkins_slave.SelectJenkinsSlave(*txn, *c.JenkinsSlave, schemaName)
		if err != nil || jsId == nil {
			return nil, errors.New(fmt.Sprintf("couldn't get jenkins slave id: %v", c.JenkinsSlave))
		}
		log.Printf("Jenkins Slave Id for %v codebase is %v", c.Name, *jsId)

		c.JenkinsSlaveId = jsId
	}

	if c.JobProvisioning != nil && *c.JobProvisioning != "" {
		jpId, err := jp.SelectJobProvision(*txn, *c.JobProvisioning, "ci", schemaName)
		if err != nil || jpId == nil {
			return nil, errors.New(fmt.Sprintf("couldn't get job provisioning id: %v", c.JobProvisioning))
		}

		log.Printf("Job Probisioning Id for %v codebase is %v", c.Name, *jpId)

		c.JobProvisioningId = jpId
	}

	if err := s.setPerfServerIdToCodebaseDto(c.Perf, schemaName); err != nil {
		return nil, errors.Wrapf(err, "couldn't set %v perf server id", c.Perf.Name)
	}

	id, err = repository.CreateCodebase(*txn, c, schemaName)
	if err != nil {
		log.Printf("Error has occurred during business entity creation: %v", err)
		return nil, errors.New(fmt.Sprintf("cannot create business entity %v", c))
	}
	log.Printf("Id of the newly created business entity is %v", *id)
	return id, nil
}

func (s CodebaseService) setPerfServerIdToCodebaseDto(perf *codebase.Perf, tenant string) error {
	if perf == nil {
		return nil
	}

	id, err := s.PerfService.GetPerfServerId(perf.Name, tenant)
	if err != nil {
		return err
	}

	if id == nil {
		return fmt.Errorf("%v perf server record doesn't exist", perf.Name)
	}
	perf.Id = id
	return nil
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

func (s CodebaseService) Delete(name, schema string) error {
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
