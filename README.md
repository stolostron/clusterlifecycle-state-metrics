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

## Deploy on RHACM

This method is for test only as it deploys some parameters are hard-coded such as the `openshift-monitoring` and `open-cluster-management` namespaces. You can use the rcm-chart to have more control.

It also creates an ingress which allows to retrieve the infomration from outside of the cluster.

1. `oc login` your hub RHACM cluster.
2. Set the image you want to deploy in [deployment.yaml](overlays/deploy/deployment.yaml)
3. run `make deploy`
4. Open Prometheus console and check for metrics "acm_managed_cluster_info"

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

