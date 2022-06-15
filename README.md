[comment]: # ( Copyright Contributors to the Open Cluster Management project )

# clusterlifecycle-state-metrics

This project generates a number of metrics used for business analysis.

## Available Metrics

- acm_managed_cluster_info

## testing

1. `make run`
2. `curl http://localhost:8080/metrics`

## Generated metrics:

```
curl http://localhost:8080/metrics
# HELP acm_managed_cluster_info Managed cluster information
# TYPE acm_managed_cluster_info gauge
acm_managed_cluster_info{hub_cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",managed_cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",vendor="OpenShift",cloud="Amazon",version="v1.16.2",created_via="Other",vcpu="4"} 1
```

## Build/Push/Deploy on RHACM

Each steps can be run separatly:

### Prereqs

1. `oc login` your hub RHACM cluster.
2. Set in the [kustomization.yaml](./deploy/kustomization.yaml#L6) the namespace where you want the collector to be deployed.
3. Set the following `IMG` environment variable, this is the name of the image and where it will be pushed.

```bash
export QUAY_USER=<your_user>
export IMG_TAG=<tag_you_want_to_use>
export IMG=quay.io/${QUAY_USER}/clusterlifecycle-state-metrics:${IMG_TAG}
make docker-build docker-push deploy
```
### build

```bash
make docker-build
```

### push

```bash
make docker-push
```

### Deploy

```bash
make deploy
```

It also creates an ingress which allows to retrieve the infomration from outside of the cluster.

The metrics then will appear on prometheus.

## promql examples:

1. Retrieve the number of imported clusters per hub:

```
sum by (hub_cluster_id) (
   acm_managed_cluster_info 
) 
```

and per vendor:

```
sum by (hub_cluster_id, vendor) (
   acm_managed_cluster_info 
) 
```

and per cloud:

```
sum by (hub_cluster_id, cloud) (
   acm_managed_cluster_info 
) 
```

and per version:

```
sum by (hub_cluster_id, version) (
   acm_managed_cluster_info 
) 
```

