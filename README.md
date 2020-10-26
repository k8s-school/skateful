Manage Qserv local storage

- Install local storage class
- Create data directories on k8s nodes
- Create yaml for PVs/PVCs

#  Use Kubernetes local storage class with Qserv

## Pre-requisites

Clone repository

```shell
git clone https://github.com/lsst-dm/qserv-storage.git
cd qserv-storage
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
