package apis

import (
	"github.com/epmd-edp/reconciler/v2/pkg/apis/edp/v1alpha1"
	"github.com/openshift/api/template/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)
}
