package cluster

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("cluster-util")

func GetOwnerReference(ownerKind string, ors []metav1.OwnerReference) *metav1.OwnerReference {
	log.Info("finding owner", "kind", ownerKind)
	if len(ors) == 0 {
		return nil
	}
	for _, o := range ors {
		if o.Kind == ownerKind {
			return &o
		}
	}
	return nil
}
