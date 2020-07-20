# Reconciler Operator

Get acquainted with the Reconciler Operator and the installation process as well as the local development.

## Overview

Reconciler Operator is an EDP operator that is responsible for saving state of CR's in EDP database. 
Operator installation can be applied on two container orchestration platforms: OpenShift and Kubernetes.
                                                                                                     
_**NOTE:** Operator is platform-independent, that is why there is a unified instruction for deploying._

## Prerequisites
* Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
* Cluster admin access to the cluster;
* EDP project/namespace is deployed by following one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/release-2.4/documentation/openshift_install_edp.md#edp-project) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/release-2.4/documentation/kubernetes_install_edp.md#edp-namespace).

## Installation
In order to install the EDP Reconciler Operator, follow the steps below:

1. To add the Helm EPAMEDP Charts for local client, run "helm repo add":
     ```bash
     helm repo add epamedp https://chartmuseum.demo.edp-epam.com/
     ```
2. Choose available Helm chart version:
     ```bash
     helm search repo epamedp/reconciler
     ```
   Example response:   
     ```bash
     NAME                    CHART VERSION   APP VERSION     DESCRIPTION
     epamedp/reconciler      v2.4.0                          Helm chart for Golang application/service deplo...
     ```

    _**NOTE:** It is highly recommended to use the latest released version._
    
3. Deploy operator:

    Full available chart parameters list:
    ```
        - <chart_version>                               # Helm chart version;
        - global.edpName                                # a namespace or a project name (in case of OpenShift);
        - global.platform                               # a platform type that can be "kubernetes" or "openshift";
        - global.database.host                          # database host;
        - global.database.name                          # database name;
        - global.database.port                          # database port;
        - name                                          # component name;
        - image.name                                    # EDP reconciler Docker image name. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/reconciler);
        - image.version                                 # EDP reconciler Docker image tag. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/reconciler/tags);
    ```
    
4. Install operator in the <edp_cicd_project> namespace with the helm command; find below the installation command example:
    ```bash
    helm install reconciler epamedp/reconciler --namespace <edp_cicd_project> --version <chart_version> --set name=reconciler --set global.edpName=<edp_cicd_project> --set global.platform=<platform_type> --set global.database.name=<db-name> --set global.database.host=<db-name>.<namespace_name> --set global.database.port=<port> 
    ```
5. Check the <edp_cicd_project> namespace that should contain operator deployment with your operator in a running status

## Local Development
In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](documentation/local-development.md) page.