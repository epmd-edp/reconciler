package model

import (
	"github.com/openshift/api/template/v1"
	"github.com/pkg/errors"
	"strings"
)

type ThirdPartyService struct {
	Name        string
	Description string
	Version     string
	Tenant      string
}

func ConvertToService(k8sObject v1.Template) (*ThirdPartyService, error) {
	if &k8sObject == nil {
		return nil, errors.New("k8s object should be not nil")
	}

	var serviceVersion string

	serviceParameters := k8sObject.Parameters
	for _, parameter := range serviceParameters {
		if parameter.Name == "SERVICE_VERSION" {
			serviceVersion = parameter.Value
			break
		}
	}

	return &ThirdPartyService{
		Name:        k8sObject.Name,
		Description: k8sObject.Annotations["description"],
		Version:     serviceVersion,
		Tenant:      strings.TrimSuffix(k8sObject.Namespace, "-edp-cicd"),
	}, nil

}
