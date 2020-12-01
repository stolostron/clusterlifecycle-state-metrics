# ocm-state-metrics

This project generates a number of metrics used for business analysis.

## Available Metrics

- ocm_clusterdeployment_created

## testing

1. `make build`
2. `./ocm-state-metrics --port=8080 --telemetry-port=8081 --kubeconfig=$KUBECONFIG`
3. `curl http://localhost:8080/metrics`

## Generated metrics:

```
curl http://localhost:8080/metrics
# HELP ocm_clusterdeployment_created Unix creation timestamp
# TYPE ocm_clusterdeployment_created gauge
ocm_clusterdeployment_created{namespace="itdove-aws-ss5t",managedcluster="itdove-aws-ss5t"} 1.604609472e+09
# HELP ocm_managedcluster_info Managed cluster information
# TYPE ocm_managedcluster_info gauge
ocm_managedcluster_info{cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",managedcluster_name="local-cluster",vendor="OpenShift",cloud="Amazon",version="v1.16.2"} 1
ocm_managedcluster_info{cluster_id="faddba46-201e-4d5d-bf52-9918517a9e6a",managedcluster_name="itdove-test",vendor="Other",cloud="Other",version="v1.19.1"} 1
```

## Deploy on RHACM

1. `oc login` your hub RHACM cluster.
2. Set the image you want to deploy in [deployment.yaml](overlays/template/deployment.yaml)
3. run `make deploy`
4. Open Prometheus console and check for metrics "ocm_managedcluster_info"
