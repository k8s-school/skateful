Manage PV/PVC for a stateful application

- Create yaml for PVs/PVCs, for local storage or GKE

For local storage
- Install local storage class
- Create data directories on k8s nodes

#  Use Kubernetes local storage class

## Pre-requisites

Clone repository

```shell
git clone https://github.com/k8s-school/skateful.git
cd skateful
cp env.example.in2p3.sh env.sh
```

Customize `env.sh` file.

## 1 - Create data directories

```shell
./mkdir.sh
# ingest/ directory need to be manually created on Qserv master node
```

## 2 - Create StorageClass, PersistentVolumes and PersistentVolumesClaims

```shell
./create-manifests.sh
kubectl apply -n qserv-dev -f out/
kubectl get pvc, pv
```
