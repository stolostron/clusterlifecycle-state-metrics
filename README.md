# clusterlifecycle-state-metrics

This project generates a number of metrics used for business analysis.

## Available Metrics

- clc_clusterdeployment_created

## testing

1. `make build`
2. `./clusterlifecycle-state-metrics --port=8080 --telemetry-port=8081 --kubeconfig=$KUBECONFIG`
3. `curl http://localhost:8080/metrics`

## Generated metrics:

```
curl http://localhost:8080/metrics
# HELP clc_clusterdeployment_created Unix creation timestamp
# TYPE clc_clusterdeployment_created gauge
clc_clusterdeployment_created{namespace="itdove-aws-ss5t",managedcluster="itdove-aws-ss5t"} 1.604609472e+09
# HELP clc_managedcluster_info Managed cluster information
# TYPE clc_managedcluster_info gauge
clc_managedcluster_info{cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",managedcluster_name="local-cluster",vendor="OpenShift",cloud="Amazon",version="v1.16.2"} 1
clc_managedcluster_info{cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",managedcluster_name="itdove-test",vendor="Other",cloud="Other",version="v1.19.1"} 1
```

## Deploy on RHACM

1. `oc login` your hub RHACM cluster.
2. Set the image you want to deploy in [deployment.yaml](overlays/template/deployment.yaml)
3. run `make deploy`
4. Open Prometheus console and check for metrics "clc_managedcluster_info"

## promql examples:

1. Retrieve the number of imported clusters per hub:

```
sum by (hub_cluster_id) (
   clc_managedcluster_info 
) 
```

and per vendor:

```
sum by (hub_cluster_id, vendor) (
   clc_managedcluster_info 
) 
```

and per cloud:

```
sum by (hub_cluster_id, cloud) (
   clc_managedcluster_info 
) 
```

and per version:

```
sum by (hub_cluster_id, version) (
   clc_managedcluster_info 
) 
```

2. Retrieve the number of cluster created by hive per hub:

```
sum by (hub_cluster_id) (
   clc_clusterdeployment_created 
) 
```

3. Retrieve the number of cluster created by hive vs imported in the hub:

```
sum(clc_managedcluster_info * on(hub_cluster_id) group_left(name) clc_clusterdeployment_created{}) by (hub_cluster_id)
```