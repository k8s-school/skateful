# Manage PVs/PVCs for a stateful application

# Backup PVCs/PVs for CSI storage

```
# Install
go install github.com/k8s-school/skateful@latest
# Watch inline doc to access user manual
skateful -h
```

# Create yaml for PVs/PVCs, for localStorageClass

This feature will move into `skateful` binary ASAP.

For local storage
- Install local storage class
- Create data directories on k8s nodes

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
./mkdirs.sh
# ingest/ directory need to be manually created on Qserv master node
```

## 2 - Create StorageClass, PersistentVolumes and PersistentVolumesClaims

```shell
./create-manifests.sh
kubectl apply -n qserv-dev -f out/
kubectl get pvc, pv
```
