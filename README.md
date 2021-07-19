# alibaba-cloud-csi-driver-operator
An operator to deploy the Alibaba Cloud CSI driver in ACK.

This operator is installed by the cluster-storage-operator.

## Quick start
Before running the operator manually, you must remove the operator installed by CSO/CVO

### Scale down CVO and CSO
> oc scale --replicas=0 deploy/cluster-version-operator -n openshift-cluster-version
> 
> oc scale --replicas=0 deploy/cluster-storage-operator -n openshift-cluster-storage-operator

### Delete operator resources (daemonset, deployments)
> oc -n openshift-cluster-csi-drivers delete deployment/csi-provisioner ds/csi-plugin

To build and run the operator locally:
### Create only the resources the operator needs to run via CLI
> oc apply -f https://raw.githubusercontent.com/openshift/cluster-storage-operator/master/assets/csidriveroperators/alibaba-cloud-csi/cr.yaml

### Build the operator
> ./build/build.sh

### Set the environment variables
> export PLUGIN_IMAGE=quay.io/openshift/origin-alibaba-cloud-csi-driver:latest
> 
> export PROVISIONER_IMAGE=quay.io/openshift/origin-csi-external-provisioner:latest
> 
> export ATTACHER_IMAGE=quay.io/openshift/origin-csi-external-attacher:latest
> 
> export RESIZER_IMAGE=quay.io/openshift/origin-csi-external-resizer:latest
> 
> export SNAPSHOTTER_IMAGE=quay.io/openshift/origin-csi-external-snapshotter:latest
> 
> export NODE_DRIVER_REGISTRAR_IMAGE=quay.io/openshift/origin-csi-node-driver-registrar:latest
> 

### Run the operator via CLI
> ./bin/alibaba-cloud-csi-driver-operator start --kubeconfig $MY_KUBECONFIG --namespace openshift-cluster-csi-drivers
