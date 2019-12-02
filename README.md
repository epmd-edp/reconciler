# How to Install operator

_EDP installation can be applied on two container orchestration platforms: OpenShift and Kubernetes._

_**NOTE:** Installation of operators is platform independent, accordingly we have unified instruction for deploying._


### Prerequisites
1. Machine with [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) installed with authorized access to the cluster.
2. Admin space is deployed using instruction of [edp-install](https://github.com/epmd-edp/edp-install#admin-space) repository

### Installation
* Go to the [releases](https://github.com/epmd-edp/reconciler/releases) page of this repository, choose a version, download an archive and unzip it.

_**NOTE:** It is highly recommended to use the latest released version._

* Go to the unzipped directory and apply all files with Custom Resource Definitions

`for file in $(ls crds/*_crd.yaml); do kubectl apply -f $file; done`

* Deploy operator

`kubectl patch -n <edp_cicd_project> -f deploy/operator.yaml --local=true --patch='{"spec":{"template":{"spec":{"containers":[{"image":"epamedp/reconciler:<operator_version>", "name":"reconciler-v2", "env": [{"name":"WATCH_NAMESPACE", "value":"<edp_cicd_project>"}, {"name":"PLATFORM_TYPE","value":"kubernetes"}, {"name":"DB_HOST","value":"<db_host>"}]}]}}}}' -o yaml | kubectl -n <edp_cicd_project> apply -f -`

_** <operator_version> - release version you've chosen_

_** <edp_cicd_project> - a namespace or project(in Opensift case) name which you created following [edp-install instructions](https://github.com/epmd-edp/edp-install#install-edp)_

_** <platform_type> - Can be "kubernetes" or "openshift"_

_** <db_host> - hostname of database. If deployed in edp-deploy project then "edp-install-wizard-db.edp-deploy"_

* Check <edp_deploy_project> namespace. It should be pending for creating secret with the following message: "Error: secrets "db-admin-console" not found". It's ok, will be fixed during EDP installation.