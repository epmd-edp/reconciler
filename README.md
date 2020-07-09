# EDP Reconciler

## Overview

Reconciler is an EDP operator that is responsible for a work with the EDP tenant database.

### Prerequisites
* Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
* Cluster admin access to the cluster;
* EDP project/namespace is deployed by following one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/release-2.4/documentation/openshift_install_edp.md#edp-project) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/release-2.4/documentation/kubernetes_install_edp.md#edp-namespace).

### Installation
In order to install the EDP Reconciler, follow the steps below:

1. To add the Helm EPAMEDP Charts for local client, run "helm repo add":
     ```bash
     helm repo add epamedp https://chartmuseum.demo.edp-epam.com/
     ```
2. Choose available Helm chart version:
     ```bash
     helm search repo epamedp/reconciler
     NAME                    CHART VERSION   APP VERSION     DESCRIPTION
     epamedp/reconciler      v2.4.0                          Helm chart for Golang application/service deplo...
     ```

    _**NOTE:** It is highly recommended to use the latest released version._
3. Deploy operator:
Parameters:
    - <chart_version>                               # a version of Helm chart;
    - global.edpName                                # a namespace or a project name (in case of OpenShift);
    - global.platform                               # a platform type that can be "kubernetes" or "openshift";
    - name                                          # component name;
    - database.required                             # database deployment request can be "true" or "false";
    - database.host                                 # database host;
    - database.name                                 # database name;
    - database.port                                 # database port;
    - image.name                                    # EDP image in [Dockerhub](https://hub.docker.com/u/epamedp);
    - image.version                                 # EDP tag. The released image can be found on [Dockerhub](https://hub.docker.com/repository/docker/epamedp/reconciler/tags);

Inspect the sample of launching a Helm chart for Reconciler installation:
```bash
helm install reconciler epamedp/reconciler --namespace <edp_cicd_project> --version <chart_version> --set name=reconciler --set global.edpName=<edp_cicd_project> --set global.platform=<platform_type> --set image.name=epamedp/reconciler --set image.version=<operator_version> 
```

4.  Check the <edp_cicd_project> namespace that should be in a pending state of creating a secret by indicating the following message: "Error: secrets "db-admin-console" not found". Such notification is a normal flow and it will be fixed during the EDP installation.

### Local Development
In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](documentation/local-development.md) page.