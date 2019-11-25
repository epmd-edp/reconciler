/*
 * Copyright 2019 EPAM Systems.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package thirdpartyservice

import (
	"github.com/openshift/api/template/v1"
	"github.com/pkg/errors"
)

type ThirdPartyService struct {
	Name        string
	Description string
	Version     string
	Tenant      string
}

func ConvertToService(k8sObject v1.Template, edpName string) (*ThirdPartyService, error) {
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
		Tenant:      edpName,
	}, nil

}
