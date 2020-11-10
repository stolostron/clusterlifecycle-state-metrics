FROM openshift/origin-release:golang-1.13 as builder

ARG REMOTE_SOURCE
ARG REMOTE_SOURCE_DIR
ARG GITHUB_TOKEN
ARG COMPONENT_VERSION

ENV USE_VENDORIZED_BUILD_HARNESS=true
ENV COMPONENT_NAME=ocm-state-metrics
ENV COMPONENT_VERSION="${COMPONENT_VERSION}"
ENV COMPONENT_TAG_EXTENSION=" "
#RUN yum install -y tar gzip make which

COPY $REMOTE_SOURCE $REMOTE_SOURCE_DIR/app/
WORKDIR $REMOTE_SOURCE_DIR/app
RUN GOFLAGS="" go build ./cmd/ocm-state-metrics
RUN GOFLAGS="" go test -covermode=atomic -coverpkg=github.com/open-cluster-management/ocm-state-metrics/pkg/... -c -tags testrunmain ./cmd/ocm-state-metrics -o ocm-state-metrics-coverage

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
RUN microdnf update

ARG VCS_REF
ARG VCS_URL
ARG IMAGE_NAME
ARG IMAGE_DESCRIPTION
ARG ARCH_TYPE
ARG REMOTE_SOURCE_DIR
ARG BUILD_HARNESS_EXTENSIONS_PROJECT

ENV OPERATOR=/usr/local/bin/ocm-state-metrics \
    USER_UID=1001 \
    USER_NAME=ocm-state-metrics

# install operator binary
COPY --from=builder $REMOTE_SOURCE_DIR/app/ocm-state-metrics ${OPERATOR}
COPY --from=builder $REMOTE_SOURCE_DIR/app/ocm-state-metrics-coverage ${OPERATOR}-coverage
COPY --from=builder \
    $REMOTE_SOURCE_DIR/app/$BUILD_HARNESS_EXTENSIONS_PROJECT/modules/component/bin/component/coverage-entrypoint-func.sh \
    /usr/local/bin/coverage-entrypoint-func.sh
COPY --from=builder $REMOTE_SOURCE_DIR/app/build/bin /usr/local/bin
# COPY --from=builder $REMOTE_SOURCE_DIR/app/build/resources /usr/local/resources
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}

LABEL com.redhat.component="ocm-state-metrics" \
      name="ocm-state-metrics-rhel8" \
      version="${CI_CONTAINER_VERSION}" \
      release="${CI_CONTAINER_RELEASE}" \
      summary="rocm-state-metrics" \
      io.openshift.expose-services="" \
      io.openshift.tags="data,images" \
      io.k8s.display-name="ocm-state-metrics" \
      maintainer="['eisraeli@redhat.com', 'jbieren@redhat.com']" \
      description="ocm-state-metrics"