# ocm-state-metrics

This project generates a number of metrics used for business analysis.

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
ocm_managedcluster_info{cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",cluster_domain="itdove-2.dev02.red-chesterfield.com",managedcluster_name="local-cluster",vendor="OpenShift",cloud="Amazon",version="v1.16.2"} 1
ocm_managedcluster_info{cluster_id="787e5a35-c911-4341-a2e7-65c415147aeb",cluster_domain="itdove-2.dev02.red-chesterfield.com",managedcluster_name="itdove-aws-ss5t",vendor="OpenShift",cloud="aws",version=""} 1
```