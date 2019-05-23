package service

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"reconciler/pkg/db"
	"reconciler/pkg/model"
	"testing"
	"time"
)

func TestBEService_CreateBE(t *testing.T) {
	service := BEService{
		DB: db.Instance,
	}
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
			Status:          "INITIALIZED",
		},
	}
	be, err := model.Convert(k8sObject)
	be.Type = "application"
	fmt.Println(err)

	err = service.PutBE(*be)

}
