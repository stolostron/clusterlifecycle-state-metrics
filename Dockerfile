FROM docker.io/openshift/origin-release:golang-1.14 AS builder
WORKDIR /go/src/github.com/open-cluster-management/ocm-state-metrics
COPY . .
ENV GO_PACKAGE github.com/open-cluster-management/ocm-state-metrics
RUN make build --warn-undefined-variables

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /go/src/github.com/open-cluster-management/ocm-state-metrics/ocm-state-metrics /
RUN microdnf update && microdnf clean all
