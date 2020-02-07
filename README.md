# How to Install Operator

EDP installation can be applied on two container orchestration platforms: OpenShift and Kubernetes.

_**NOTE:** Installation of operators is platform-independent, that is why there is a unified instruction for deploying._

### Prerequisites
1. Linux machine or Windows Subsystem for Linux instance with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed with an authorized access to the cluster;
2. Admin space is deployed by following one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/master/documentation/openshift_install.md#admin-space) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/master/documentation/kubernetes_install.md#admin-space).

### Installation
* Go to the [releases](https://github.com/epmd-edp/reconciler/releases) page of this repository, choose a version, download an archive and unzip it;

_**NOTE:** It is highly recommended to use the latest released version._

* Go to the unzipped directory and apply all files with the Custom Resource Definitions resource:
```bash
for file in $(ls deploy/crds/*_crd.yaml); do kubectl apply -f $file; done
```

* Deploy operator:
```bash
kubectl patch -n <edp_cicd_project> -f deploy/operator.yaml --local=true --patch='{"spec":{"template":{"spec":{"containers":[{"image":"epamedp/reconciler:<operator_version>", "name":"reconciler-v2", "env": [{"name":"WATCH_NAMESPACE", "value":"<edp_cicd_project>"}, {"name":"PLATFORM_TYPE","value":"kubernetes"}, {"name":"DB_HOST","value":"<db_host>"}]}]}}}}' -o yaml | kubectl -n <edp_cicd_project> apply -f -
```

- _<operator_version> - a selected release version;_

- _<edp_cicd_project> - a namespace or a project name (in case of OpenSift) that is created by one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/master/documentation/openshift_install.md#install-edp) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/master/documentation/kubernetes_install.md#install-edp);_

- _<platform_type> - a platform type that can be "kubernetes" or "openshift"_.

- _<db_host> - a hostname of database. If it is deployed in the edp-deploy project then it will be as follows: "edp-install-wizard-db.edp-deploy"._

* Check the <edp_deploy_project> namespace that should be in a pending state of creating a secret by indicating the following message: "Error: secrets "db-admin-console" not found". Such notification is a normal flow and it will be fixed during the EDP installation.

### Local Development
To setup local environment for developing operator use [Local Development](documentation/local-development.md) user guide.