module github.com/epmd-edp/reconciler/v2

go 1.12

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37

require (
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/epmd-edp/cd-pipeline-operator/v2 v2.3.0-58.0.20200424095303-90cf9714dc73
	github.com/epmd-edp/codebase-operator/v2 v2.3.0-95.0.20200416135035-a8dc7f65f401
	github.com/epmd-edp/edp-component-operator v0.0.1-2
	github.com/epmd-edp/jenkins-operator/v2 v2.3.0-130.0.20200416062406-16c330e09a19
	github.com/lib/pq v1.0.0
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.0.0-20190530173525-d6f9cdf2f52e
	github.com/pkg/errors v0.8.1
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
	sigs.k8s.io/controller-runtime v0.1.12
)
