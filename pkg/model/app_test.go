package model

import (
	"business-app-reconciler-controller/models"
	edpv1alpha1 "business-app-reconciler-controller/pkg/apis/edp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	k8sObject := edpv1alpha1.BusinessApplication{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fightclub",
			Name:      "fc-ui",
		},
		Spec: edpv1alpha1.BusinessApplicationSpec{
			Lang:      "java",
			Framework: "spring-boot",
			BuildTool: "maven",
			Strategy:  models.Create,
		},
		Status: edpv1alpha1.BusinessApplicationStatus{
			Available:       true,
			LastTimeUpdated: time.Now(),
			Status:          "created",
		},
	}

	app, err := Convert(k8sObject)
	if err != nil {
		t.Fatal(err)
	}

	if app.Name != "fc-ui" {
		t.Fatal("name is not fc-ui")
	}
}
