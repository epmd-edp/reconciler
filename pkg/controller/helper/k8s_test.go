package helper

import (
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func TestGetEDPNameHappyPath(t *testing.T) {
	// given
	ns := "test-ns"
	edpN := "foobar"
	cm := &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      EDPConfigCM,
			Namespace: ns,
		},
		Data: map[string]string{
			EDPNameKey: edpN,
		},
	}
	cl := fake.NewFakeClient(cm)

	// when
	actN, err := GetEDPName(cl, ns)

	//then
	if err != nil {
		t.Errorf("GetEDPName() error = %v, wantErr %v", err, nil)
	}
	if actN == nil || *actN != edpN {
		t.Errorf("GetEDPName() expected = %v, actual = %v", edpN, actN)
	}
}

func TestGetEDPNameNoKey(t *testing.T) {
	// given
	ns := "test-ns"
	cm := &coreV1.ConfigMap{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      EDPConfigCM,
			Namespace: ns,
		},
		Data: map[string]string{},
	}
	cl := fake.NewFakeClient(cm)

	// when
	actN, err := GetEDPName(cl, ns)

	//then
	if err == nil {
		t.Errorf("GetEDPName() error = %v, wantErr %v", err, "not key error")
	}
	if actN != nil {
		t.Errorf("GetEDPName() expected = %v, actual = %v", nil, actN)
	}
}

func TestGetEDPNameNoCM(t *testing.T) {
	// given
	ns := "test-ns"
	cl := fake.NewFakeClient()

	// when
	actN, err := GetEDPName(cl, ns)

	//then
	if err == nil {
		t.Errorf("GetEDPName() error = %v, wantErr %v", err, "not found error")
	}
	if actN != nil {
		t.Errorf("GetEDPName() expected = %v, actual = %v", nil, actN)
	}
}
