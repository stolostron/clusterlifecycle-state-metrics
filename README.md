# clusterlifecycle-state-metrics

This project generates a number of metrics used for business analysis.

## Available Metrics

- clc_clusterdeployment_created

## testing

1. `make run`
2. `curl http://localhost:8080/metrics`

## Generated metrics:

```
curl http://localhost:8080/metrics
# HELP clc_managedcluster_info Managed cluster information
# TYPE clc_managedcluster_info gauge
clc_managedcluster_info{hub_cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",vendor="OpenShift",cloud="Amazon",version="v1.16.2",created_via="Other"} 1
```

## Deploy on RHACM

This method is for test only as it deploys some parameters are hard-coded such as the `openshift-monitoring` and `open-cluster-management` namespaces. You can use the rcm-chart to have more control.

1. `oc login` your hub RHACM cluster.
2. Set the image you want to deploy in [deployment.yaml](overlays/deploy/deployment.yaml)
3. run `make deploy`
4. Open Prometheus console and check for metrics "clc_managedcluster_info"

The metrics then will appear on prometheus.

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

## Add a new metric

1. Update the [pkg/options/collector.go](pkg/options/collector.go) `DefaultCollectors` varaible with your new metric name.
2. Update the [pkg/collectors/builder.go](pkg/collectors/builder.go) `availableCollectors` variable with your new metric name and create similar methods than `buildManagedClusterInfoCollector`
3. Clone the [pkg/collectors/managedclusterinfo.go](pkg/collectors/managedclusterinfo.go) to implement your metric and adapt it for your new metric.
4. Create unit tests.
5. Create functianal tests.