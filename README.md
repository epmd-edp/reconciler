# EDP Reconciler

## Overview

Reconciler is en EDP operator that is responsible for a work with the EDP tenant database.

### Prerequisites
* Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
* Cluster admin access to the cluster;
* EDP project/namespace is deployed by following one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/release-2.3/documentation/openshift_install_edp.md#edp-project) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/release-2.3/documentation/kubernetes_install_edp.md#edp-namespace).

### Installation
In order to install the EDP Reconciler, follow the steps below:

1. Go to the [releases](https://github.com/epmd-edp/reconciler/releases) page of this repository, choose a version, download an archive, and unzip it;

    _**NOTE:** It is highly recommended to use the latest released version._
2. Navigate to the unzipped directory and deploy an operator:
    ```bash
    helm install reconciler --namespace <edp_cicd_project> --set name=reconciler --set namespace=<edp_cicd_project> --set platform=<platform_type> --set image.name=epamedp/reconciler --set image.version=<operator_version> deploy-templates
    ```
    - _<edp_cicd_project> - a namespace or a project name (in case of OpenShift) that is created by one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/release-2.3/documentation/openshift_install_edp.md#edp-project) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/release-2.3/documentation/kubernetes_install_edp.md#edp-namespace);_ 
    - _<platform_type> - a platform type that can be "kubernetes" or "openshift";_
    - _<operator_version> - a selected release version tag for the operator from Docker Hub;_

3.  Check the <edp_cicd_project> namespace that should be in a pending state of creating a secret by indicating the following message: "Error: secrets "db-admin-console" not found". Such notification is a normal flow and it will be fixed during the EDP installation.

### Local Development
In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](documentation/local-development.md) page.