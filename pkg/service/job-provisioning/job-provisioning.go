package job_provisioning

import (
	"database/sql"
	jenkinsV2Api "github.com/epmd-edp/jenkins-operator/v2/pkg/apis/v2/v1alpha1"
	jp "github.com/epmd-edp/reconciler/v2/pkg/repository/job-provisioning"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("job-provisioning-service")

type JobProvisionService struct {
	DB *sql.DB
}

func (s JobProvisionService) PutJobProvisions(provisions []jenkinsV2Api.JobProvision, schemaName string) error {
	log.Info("Start executing PutJobProvisions method... ")

	txn, err := s.DB.Begin()
	if err != nil {
		return err
	}

	for _, p := range provisions {
		id, err := jp.SelectJobProvision(*txn, p.Name, schemaName)
		if err != nil {
			_ = txn.Rollback()
			return errors.Wrapf(err, "an error has occurred while selecting job provision %v", p.Name)
		}

		if id != nil {
			log.Info("Job Provision already exists. Skip adding into db", "name", p)
			continue
		}

		err = jp.CreateJobProvision(*txn, p.Name, schemaName)
		if err != nil {
			_ = txn.Rollback()
			return errors.Wrapf(err, "an error has occurred while creating job provision %v", p.Name)
		}
	}

	err = txn.Commit()
	if err != nil {
		return err
	}

	log.Info("End executing PutJobProvisions method... ")

	return err
}
