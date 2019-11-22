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

// Package helper implements simple methods to work with k8s API
package helper

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EDPConfigCM = "edp-config"
	EDPNameKey  = "edp_name"
)

// GetEDPName tries to find edp name parameter from edp-config CM using
// provided client and namespace to search
func GetEDPName(client client.Client, namespace string) (*string, error) {
	cm := &v1.ConfigMap{}
	err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: namespace,
		Name:      EDPConfigCM,
	}, cm)
	if err != nil {
		return nil, err
	}
	r := cm.Data[EDPNameKey]
	if len(r) == 0 {
		return nil, fmt.Errorf("there is not key %v in cm %v", EDPNameKey, EDPConfigCM)
	}
	return &r, nil
}
