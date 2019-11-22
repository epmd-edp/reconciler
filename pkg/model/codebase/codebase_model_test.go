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

package codebase

import (
	edpv1alpha1 "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	frw := "spring-boot"
	k8sObject := edpv1alpha1.Codebase{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fightclub",
			Name:      "fc-ui",
		},
		Spec: edpv1alpha1.CodebaseSpec{
			Lang:      "java",
			Framework: &frw,
			BuildTool: "maven",
			Strategy:  edpv1alpha1.Create,
		},
		Status: edpv1alpha1.CodebaseStatus{
			Available:       true,
			LastTimeUpdated: time.Now(),
			Status:          "created",
		},
	}

	app, err := Convert(k8sObject, "foobar")
	if err != nil {
		t.Fatal(err)
	}

	if app.Name != "fc-ui" {
		t.Fatal("name is not fc-ui")
	}
}
