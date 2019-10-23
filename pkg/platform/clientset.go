package platform

import (
	edpv1alpha1Codebase "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"
	appsV1client "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	projectV1Client "github.com/openshift/client-go/project/clientset/versioned/typed/project/v1"
	routeV1Client "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
	securityTypedClient "github.com/openshift/client-go/security/clientset/versioned/typed/security/v1"
	templateV1Client "github.com/openshift/client-go/template/clientset/versioned/typed/template/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var SchemeGroupVersion = schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}

type ClientSet struct {
	CoreClient     *coreV1Client.CoreV1Client
	TemplateClient *templateV1Client.TemplateV1Client
	ProjectClient  *projectV1Client.ProjectV1Client
	SecurityClient *securityTypedClient.SecurityV1Client
	AppClient      *appsV1client.AppsV1Client
	RouteClient    *routeV1Client.RouteV1Client
	EDPRestClient  *rest.RESTClient
}

func CreateOpenshiftClients() (*ClientSet, error) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	restConfig, err := config.ClientConfig()
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	coreClient, err := coreV1Client.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	templateClient, err := templateV1Client.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	projectClient, err := projectV1Client.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	securityClient, err := securityTypedClient.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	appClient, err := appsV1client.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	routeClient, err := routeV1Client.NewForConfig(restConfig)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}
	edpRestClient, err := getApplicationClient(config)
	if err != nil {
		log.Printf("[ERROR] %s", err)
		return nil, err
	}

	return &ClientSet{
		CoreClient:     coreClient,
		TemplateClient: templateClient,
		ProjectClient:  projectClient,
		SecurityClient: securityClient,
		AppClient:      appClient,
		RouteClient:    routeClient,
		EDPRestClient:  edpRestClient,
	}, nil
}

func getApplicationClient(k8sConfig clientcmd.ClientConfig) (*rest.RESTClient, error) {
	var config *rest.Config
	var err error

	config, err = k8sConfig.ClientConfig()

	if err != nil {
		return nil, err
	}

	clientset, err := createCrdClient(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func createCrdClient(cfg *rest.Config) (*rest.RESTClient, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.GroupVersion = &SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&edpv1alpha1Codebase.CodebaseBranch{},
		&edpv1alpha1Codebase.CodebaseBranchList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
