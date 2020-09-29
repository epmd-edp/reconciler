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

package service

import (
	"github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
)

type ServiceDto struct {
	Name        string
	Version     string
	Description string
	Url         string
	Icon        string
	SchemaName  string
}

func ConvertToServiceDto(service v1alpha1.Service, edpName string) ServiceDto {
	return ServiceDto{
		Name:        service.Name,
		Version:     service.Spec.Version,
		Description: service.Spec.Description,
		Url:         service.Spec.Url,
		Icon:        service.Spec.Icon,
		SchemaName:  edpName,
	}

}
