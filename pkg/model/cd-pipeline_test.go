package model

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
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

	cdPipeline, err := ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	if cdPipeline.Name != name {
		t.Fatal(fmt.Sprintf("name is not %v", name))
	}

	checkSpecField(t, cdPipeline.InputDockerStreams, inputDockerStream, "input docker stream")

	checkSpecField(t, cdPipeline.ThirdPartyServices, thirdPartyServices, "third party services")

	checkSpecField(t, cdPipeline.ApplicationsToPromote, applicationsToPromote, "applications to promote")

	if cdPipeline.ActionLog.Event != formatStatus(event) {
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
		acceptCdPipelineRegistration = "accept_cd_pipeline_registration"
		jenkinsConfiguration         = "jenkins_configuration"
		setupInitialStructure        = "setup_initial_structure"
		cdPipelineRegistration       = "cd_pipeline_registration"
		nonExistedAction             = "fake-action"

		acceptCdPipelineRegistrationMsg = "Accept CD Pipeline %v registration"
		jenkinsConfigurationMsg         = "CI Jenkins pipelines codebase %v provisioning"
		setupInitialStructureMsg        = "Initial structure for CD Pipeline %v is created"
		cdPipelineRegistrationMsg       = "CD Pipeline %v registration"
		nonExistedActionMsg             = "fake message"
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
			Action:          acceptCdPipelineRegistration,
			Result:          result,
			Available:       true,
			LastTimeUpdated: time.Now(),
			Status:          event,
		},
	}

	cdPipeline, err := ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(acceptCdPipelineRegistrationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = jenkinsConfiguration
	cdPipeline, err = ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(jenkinsConfigurationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = setupInitialStructure
	cdPipeline, err = ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(setupInitialStructureMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = cdPipelineRegistration
	cdPipeline, err = ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, fmt.Sprintf(cdPipelineRegistrationMsg, name), cdPipeline.ActionLog.ActionMessage,
		fmt.Sprintf("converted action is incorrect %v", cdPipeline.ActionLog.ActionMessage))

	k8sObj.Status.Action = nonExistedAction
	cdPipeline, err = ConvertToCDPipeline(k8sObj)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotEqual(t, nonExistedActionMsg, cdPipeline.ActionLog.ActionMessage)
}
