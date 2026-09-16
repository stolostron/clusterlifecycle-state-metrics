# Architecture

`clusterlifecycle-state-metrics` is a Go service that watches Kubernetes and
Open Cluster Management resources and exposes lifecycle state as Prometheus
metrics. It is deployed to an RHACM hub through the manifests under
`overlays/deploy`.

## Runtime flow

1. `cmd/clusterlifecycle-state-metrics/main.go` parses options, builds the
   Kubernetes and controller-runtime clients, and starts the HTTP metrics and
   telemetry endpoints.
2. The configured collector set in `pkg/options` selects managed-cluster,
   managed-cluster-addon, and ManifestWork collectors.
3. Controllers and resource informers observe hub resources. Resource-specific
   generators under `pkg/generators` translate object state, labels, counts,
   timestamps, worker capacity, and conditions into Prometheus samples.
4. Cache implementations under `pkg/collectors` retain identifiers,
   timestamps, hibernation state, and scrape counters between observations.
   `composedMetricsCollector` writes the generated metrics to the HTTP handler.
5. Prometheus discovers the service through the ServiceMonitor in the selected
   Kustomize overlay. PrometheusRule resources provide alerting rules where
   configured.

## Main modules

- `pkg/options`: command-line defaults and collector registration. The init
  hook adds the custom managed-cluster collector to kube-state-metrics' valid
  collector set.
- `pkg/controllers`: reconciliation for resources whose lifecycle transitions
  require controller-runtime behavior, including ManifestWork state.
- `pkg/collectors`: shared collection orchestration and stateful caches.
- `pkg/generators`: resource-to-metric conversion, kept separate from API
  watching so metric semantics can be unit tested independently.
- `pkg/common`: shared constants and helpers used by generators and collectors.

## Test and deployment paths

Unit tests live beside their packages and use the checked-in vendor tree.
Environment-backed tests use controller-runtime envtest assets. Functional
tests under `test/functional` require a Kubernetes or kind environment and
exercise rendered deployment behavior, TLS, and resource events. The deploy
overlay renders the service, RBAC, ServiceMonitor, and PrometheusRule; the
test overlays provide alternate HTTP-oriented resources.

When changing a metric label or lifecycle interpretation, update the matching
generator tests first and then consider functional coverage for the affected
resource watch or deployment path.
