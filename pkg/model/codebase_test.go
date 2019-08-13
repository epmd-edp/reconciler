package model

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	k8sObject := edpv1alpha1.Codebase{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "fightclub",
			Name:      "fc-ui",
		},
		Spec: edpv1alpha1.CodebaseSpec{
			Lang:      "java",
			Framework: "spring-boot",
			BuildTool: "maven",
			Strategy:  edpv1alpha1.Create,
		},
		Status: edpv1alpha1.CodebaseStatus{
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
