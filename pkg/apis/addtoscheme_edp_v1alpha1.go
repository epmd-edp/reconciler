package apis

import (
	"github.com/openshift/api/template/v1"
	"reconciler/pkg/apis/edp/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)
}
