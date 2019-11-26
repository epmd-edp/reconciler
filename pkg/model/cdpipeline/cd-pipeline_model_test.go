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

package cdpipeline

import (
	"fmt"
	edpv1alpha1 "github.com/epmd-edp/cd-pipeline-operator/v2/pkg/apis/edp/v1alpha1"
	"github.com/epmd-edp/reconciler/v2/pkg/model"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

const (
	name                  = "fake-name"
	username              = "fake-user"
	detailedMessage       = "fake-detailed-message"
	inputDockerStream     = "fake-docker-stream-verified"
	thirdPartyServices    = "rabbit-mq"
	applicationsToPromote = "fake-application"
	result                = "success"
	cdPipelineAction      = "setup_initial_structure"
	event                 = "created"
	edpName               = "foobar"
)

func TestConvertMethodToCDPipeline(t *testing.T) {
	timeNow := time.Now()

	k8sObj := edpv1alpha1.CDPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-namespace",
			Name:      name,
		},
		Spec: edpv1alpha1.CDPipelineSpec{
			Name:                  name,
			InputDockerStreams:    []string{inputDockerStream},
			ThirdPartyServices:    []string{thirdPartyServices},
			ApplicationsToPromote: []string{applicationsToPromote},
		},
		Status: edpv1alpha1.CDPipelineStatus{
			Username:        username,
			DetailedMessage: detailedMessage,
			Value:           "active",
			Action:          cdPipelineAction,
			Result:          result,
			Available:       true,
			LastTimeUpdated: timeNow,
			Status:          event,
		},
	}

	cdPipeline, err := ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}

	if cdPipeline.Name != name {
		t.Fatal(fmt.Sprintf("name is not %v", name))
	}

	checkSpecField(t, cdPipeline.InputDockerStreams, inputDockerStream, "input docker stream")

	checkSpecField(t, cdPipeline.ThirdPartyServices, thirdPartyServices, "third party services")

	checkSpecField(t, cdPipeline.ApplicationsToPromote, applicationsToPromote, "applications to promote")

	if cdPipeline.ActionLog.Event != model.FormatStatus(event) {
		t.Fatal(fmt.Sprintf("event has incorrect status %v", event))
	}

	if cdPipeline.ActionLog.DetailedMessage != detailedMessage {
		t.Fatal(fmt.Sprintf("detailed message is incorrect %v", detailedMessage))
	}

	if cdPipeline.ActionLog.Username != username {
		t.Fatal(fmt.Sprintf("username is incorrect %v", username))
	}

	if !cdPipeline.ActionLog.UpdatedAt.Equal(timeNow) {
		t.Fatal(fmt.Sprintf("'updated at' is incorrect %v", username))
	}

	if cdPipeline.ActionLog.Action != cdPipelineAction {
		t.Fatal(fmt.Sprintf("action is incorrect %v", cdPipelineAction))
	}

	if cdPipeline.ActionLog.Result != result {
		t.Fatal(fmt.Sprintf("result is incorrect %v", result))
	}

	actionMessage := fmt.Sprintf(cdPipelineActionMessageMap[cdPipelineAction], name)
	if cdPipeline.ActionLog.ActionMessage != actionMessage {
		t.Fatal(fmt.Sprintf("action message is incorrect %v", actionMessage))
	}

}

func checkSpecField(t *testing.T, src []string, toCheck string, entityName string) {
	if len(src) != 1 {
		t.Fatal(fmt.Sprintf("%v has incorrect size", entityName))
	}

	if src[0] != toCheck {
		t.Fatal(fmt.Sprintf("%v name is not %v", entityName, toCheck))
	}
}

func TestCDPipelineActionMessages(t *testing.T) {

	var (
		acceptCdPipelineRegistrationMsg = "Accept CD Pipeline %v registration"
		jenkinsConfigurationMsg         = "CI Jenkins pipelines %v provisioning"
		setupInitialStructureMsg        = "Initial structure for CD Pipeline %v is created"
	)

	k8sObj := edpv1alpha1.CDPipeline{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fake-namespace",
			Name:      name,
		},
		Spec: edpv1alpha1.CDPipelineSpec{
			Name:                  name,
			InputDockerStreams:    []string{inputDockerStream},
			ThirdPartyServices:    []string{thirdPartyServices},
			ApplicationsToPromote: []string{applicationsToPromote},
		},
		Status: edpv1alpha1.CDPipelineStatus{
			Username:        username,
			DetailedMessage: detailedMessage,
			Value:           "active",
			Action:          edpv1alpha1.AcceptCDPipelineRegistration,
			Result:          result,
			Available:       true,
			LastTimeUpdated: time.Now(),
			Status:          event,
		},
	}

	cdPipeline, err := ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(acceptCdPipelineRegistrationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = edpv1alpha1.JenkinsConfiguration
	cdPipeline, err = ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(jenkinsConfigurationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = edpv1alpha1.SetupInitialStructureForCDPipeline
	cdPipeline, err = ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(setupInitialStructureMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = edpv1alpha1.AcceptCDPipelineRegistration
	cdPipeline, err = ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(acceptCdPipelineRegistrationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status = edpv1alpha1.CDPipelineStatus{}
	cdPipeline, err = ConvertToCDPipeline(k8sObj, edpName)
	if err != nil {
		t.Fatal(err)
	}
}
