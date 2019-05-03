# Qdrouterd Operator

A Kubernetes operator to manage Qdrouterd interior and deployments, automating creation and administration

## Introduction

Qdrouterd Operator proves a `Qdrouterd` [Custom Resource Definition](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions/) (CRD) that models a Qdrouterd deployment. This CRD allows for specifying the number of messaging routers, the deployment topology as well as other options for the interconnect operation:

## Usage

Deploy the Qdrouterd Operator into the Kubernetes cluster where it will manage requests for the `Qdrouterd` resource. The Qdrouterd Operator will watch for create, update and delete resource requests and perform the necessary steps to ensure the present cluster state matches the desired state.

### Deploy Qdrouterd Operator

The `deploy` directory contains the manifests needed to properly install the
Operator.

Create the service account for the operator.

```
$ kubectl create -f deploy/service_account.yaml
```

Create the RBAC role and role-binding that grants the permissions
necessary for the operator to function.

```
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
```

Deploy the CRD to the cluster that defines the Qdrouterd resource.

```
$ kubectl create -f deploy/crds/interconnectedcloud_v1alpha1_qdrouterd_crd.yaml
```

Next, deploy the operator into the cluster.

```
$ kubectl create -f deploy/operator.yaml
```

This step will create a pod on the Kubernetes cluster for the Qdrouterd Operator.
Observe the Qdrouterd Operator pod and verify it is in the running state.

```
$ kubectl get pods -l name=qdrouterd-operator
```

If for some reason, the pod does not get to the running state, look at the
pod details to review any event that prohibited the pod from starting.

```
$ kubectl describe pod -l name=qdrouterd-operator
```

You will be able to confirm that the new CRD has been registered in the cluster and you can review its details.

```
$ kubectl get crd
$ kubectl describe crd qdrouterds.interconnectedcloud.github.io
```

To create a Qdrouterd deployment, you must create a `Qdrouterd` resource representing the desired specification of the deployment. For example, to create a 3-node Qdrouterd mesh deployment you may run:

```console
$ cat <<EOF | kubectl create -f -
apiVersion: interconnectedcloud.github.io/v1alpha1
kind: Qdrouterd
metadata:
  name: example-interconnect
spec:
  # Add fields here
  deploymentPlan:
    image: quay.io/interconnectedcloud/qdrouterd:1.6.0
    role: interior
    size: 3
    placement: Any
EOF
```

The Qdrouterd Operator will create a deployment of three router instances, all connected together with default address semantics. It will also create a service through which the interior router mesh can be accessed. It will configure a default set of listeners and connectors as described below. You will be able to confirm that the instance has been created in the cluster and you can review its details. To view the Qdrtouterd instance, the deployment it manages and the associated pods that are deployed:

```
$ kubectl describe qdr example-interconnect
$ kubectl describe deploy example-interconnect
$ kubectl describe svc example-interconnect
$ kubectl get pods -o yaml
```

### Deployment Plan

The Qdrouterd Operator *Deployment Plan* defines the attributes for a custom resource instance.

#### Role and Placement

The *Deployment Plan* **Role** defines the mode of operation for the routers in a topology.

  * interior - This mode creates an interconnect of auto-meshed routers for concurrent connection capacity and resiliency.
    Connectivity between the interior routers will be defined by *InterRouterListeners* and *InterRouterConnectors*. Downlink
    connectivity with *edge* routers will be via the *EdgeListeners*.

  * edge -  This mode creates a set of stand-alone edge routers. The connectivity from the edge to interior routers
    will be via the *EdgeConnectors*.

The *Deployment Plan* **Placement** defines the deployment resource and the associated scheduling of the pods in the cluster.

  * Any - There is no constraint on pod placement. The operator will manage a *Deployment* resource where the number of
    pods scheduled will be up to the *Deployment Plan Size* defined.

  * Every - The pod placement is for each node in the cluster. The operator will manage a "DaemonSet" resource where the
    number of pods scheduled will correspond to the number of nodes in the cluster. The *Deployment Plan Size* is
    disregarded with this placement declaration.

  * Anti-Affinity - This constrains scheduling and prevents multiple router pods from running on the same node in the
    cluster. The operator will manage a *Deployment* resource with a number of pods up to the *Deployment Plan Size*.
    If *Deployment Plan Size* is greater than the number of nodes in the cluster, the excess pods that cannot be
    schedule will remain in the *pending* state.

### Connectivity

The connectivity between routers in a deployment is via the declared *listeners* and *connectors*. There are three types of *listeners* supported by the operator.

  * Listeners - A normal listener for accepting messaging client connections. The operator supports this listener for
    both *interior* and *edge* routers.

  * InterRouterListeners - An inter router listener for accepting connections from peer *interior* routers. The operator
    support this listener for *interior* routers **only**.

  * EdgeListeners - An edge listener for accepting connections from downlink *edge* routers. The operator supports this
    listener for *interior* routers **only**.

There are three types of *connectors* supported by the operator.

  * Connectors - A normal connector for connecting to an external messaging intermediary. The operator supports this connector
    for both *interior* and *edge* routers.

  * InterRouterConnectors - An inter router connector for establishing connectivity to peer *interior* routers. The operator
    supports this listener for *interior* routers **only**.

  * EdgeConnector - An edge connector for establishing up-link connectivity from *edge* to *interior* routers. The operator
    supports this listener for *edge* routers **only**.

## Development

This Operator is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk). Follow the [Quick Start](https://github.com/operator-framework/operator-sdk) instructions to checkout and install the operator-sdk CLI.

Local development may be done with [minikube](https://github.com/kubernetes/minikube) or [minishift](https://www.okd.io/minishift/).

#### Source Code

Clone this repository to a location on your workstation such as `$GOPATH/src/github.com/ORG/REPO`. Navigate to the repository and install the dependencies.

```
$ cd $GOPATH/src/github.com/ORG/REPO/qdrouterd-operator
$ dep ensure && dep status
```

#### Run Operator Locally

Ensure the service account, role, role bindings and CRD are added to  the local cluster.

```
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/crds/interconnectedcloud_v1alpha1_qdrouterd_crd.yaml
```

Start the operator locally for development.

```
$ operator-sdk up local
```

Create a minimal Qdrouterd resource to observe and test your changes.

```console
$ cat <<EOF | kubectl create -f -
apiVersion: interconnectedcloud.github.io/v1alpha1
kind: Qdrouterd
metadata:
  name: example-interconnect
spec:
  deploymentPlan:
    image: quay.io/interconnectedcloud/qdrouterd:1.6.0
    role: interior
    size: 3
    placement: Any
EOF
```

As you make local changes to the code, restart the operator to enact the changes.

#### Build

The Makefile will do the dependency check, operator-sdk generate k8s, run local test, and finally the operator-sdk build. Please ensure any local docker server is running.

```
make
```

#### Test

Before submitting PR, please test your code. 

File or local validation.
```
$ make test
```

Cluster-based test. 
Ensure there is a cluster running before running the test.

```
$ make cluster-test
```

## Manage the operator using the Operator Lifecycle Manager

Ensure the Operator Lifecycle Manager is installed in the local cluster.  By default, the `catalog-source.sh` will install the operator catalog resources in `operator-lifecycle-manager` namespace.  You may also specify different namespace where you have the Operator Lifecycle Manager installed.

```
$ ./hack/catalog-source.sh [namespace]
$ oc apply -f deploy/olm-catalog/qdrouterd-operator/0.1.0/catalog-source.yaml
```
