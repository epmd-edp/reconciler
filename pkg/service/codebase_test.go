package service

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reconciler/models"
	edpv1alpha1 "reconciler/pkg/apis/edp/v1alpha1"
	"reconciler/pkg/db"
	"reconciler/pkg/model"
	"testing"
	"time"
)

func TestBEService_CreateBE(t *testing.T) {
	database, err := db.InitConnection()
	if err != nil {
		t.Fatal(err)
	}

	service := BEService{
		DB: *database,
	}
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
			Status:          "INITIALIZED",
		},
	}
	be, err := model.Convert(k8sObject)
	be.Type = model.App
	fmt.Println(err)

	err = service.PutBE(*be)

}
