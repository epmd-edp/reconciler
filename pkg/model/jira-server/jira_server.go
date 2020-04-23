package jira_server

import "github.com/epmd-edp/codebase-operator/v2/pkg/apis/edp/v1alpha1"

type JiraServer struct {
	Name      string
	Available bool
	Tenant    string
}

func ConvertSpecToJira(jira v1alpha1.JiraServer, tenant string) JiraServer {
	return JiraServer{
		Name:      jira.Name,
		Available: jira.Status.Available,
		Tenant:    tenant,
	}
}
