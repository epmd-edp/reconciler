package platform

import (
	"os"
)

const (
	openshift       = "openshift"
	platformTypeKey = "PLATFORM_TYPE"
)

func IsOpenshift() bool {
	return os.Getenv(platformTypeKey) == openshift
}
