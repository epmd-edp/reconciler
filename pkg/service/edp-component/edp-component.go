package edp_component

import (
	"database/sql"
	"fmt"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	ec "github.com/epmd-edp/reconciler/v2/pkg/repository/edp-component"
	"github.com/pkg/errors"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
)

var log = logf.Log.WithName("edp-component-service")

type EDPComponentService struct {
	DB *sql.DB
}

func (s EDPComponentService) PutEDPComponent(component model.EDPComponent, schemaName string) error {
	log.Info("Start executing PutEDPComponent method...", "type", component.Type)

	t, err := s.DB.Begin()
	if err != nil {
		return err
	}

	id, err := ec.SelectEDPComponent(*t, component.Type, schemaName)
	if err != nil {
		_ = t.Rollback()
		return errors.Wrap(err, "rollback while executing SelectEDPComponent method")
	}

	if id != nil {
		log.Info("Component already exists in DB. Skip insert", "type", component.Type)
		_ = t.Rollback()
		return nil
	}

	tryToModifyUrl(&component)

	err = ec.CreateEDPComponent(*t, component, schemaName)
	if err != nil {
		_ = t.Rollback()
		return errors.Wrapf(err, "an error has occurred while creating edp component with type %v", component.Type)
	}
	log.Info("EDP component is added", "type", component.Type, "url", component.Url)

	err = t.Commit()
	if err != nil {
		return err
	}

	log.Info("End executing PutEDPComponent method... ", "type", component.Type)

	return nil
}

func tryToModifyUrl(c *model.EDPComponent) {
	if !strings.HasPrefix(c.Url, "https://") {
		c.Url = fmt.Sprintf("https://%v", c.Url)
	}
}
