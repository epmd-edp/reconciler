package model

import (
	edpComponentV1Api "github.com/epmd-edp/edp-component-operator/pkg/apis/v1/v1alpha1"
	"github.com/pkg/errors"
)

type EDPComponent struct {
	Type string
	Url  string
	Icon string
}

func ConvertToEDPComponent(k8sObj edpComponentV1Api.EDPComponent) (*EDPComponent, error) {
	if &k8sObj == nil {
		return nil, errors.New("k8s EDP component object should not be nil")
	}
	s := k8sObj.Spec

	return &EDPComponent{
		Type: s.Type,
		Url:  s.Url,
		Icon: s.Icon,
	}, nil
}
