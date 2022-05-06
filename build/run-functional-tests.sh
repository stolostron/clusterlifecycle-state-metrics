#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

set -e
# set -x

CURR_FOLDER_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
KIND_KUBECONFIG="${CURR_FOLDER_PATH}/../kind_kubeconfig.yaml"
CLUSTER_NAME=${PROJECT_NAME}-functional-test
export KUBECONFIG=${KIND_KUBECONFIG}
export DOCKER_IMAGE_AND_TAG=${1}

export FUNCT_TEST_TMPDIR="${CURR_FOLDER_PATH}/../test/functional/tmp"
export FUNCT_TEST_COVERAGE="${CURR_FOLDER_PATH}/../test/functional/coverage"

if ! which kubectl > /dev/null; then
    echo "installing kubectl"
    curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/$(uname | awk '{print tolower($0)}')/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
fi
if ! which kind > /dev/null; then
    echo "installing kind"
    curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.12.0/kind-$(uname)-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
fi
if ! which ginkgo > /dev/null; then
    echo "Installing ginkgo ..."
    pushd $(mktemp -d)
    GOSUMDB=off go get github.com/onsi/ginkgo/ginkgo
    GOSUMDB=off go get github.com/onsi/gomega/...
    popd
fi
if ! which gocovmerge > /dev/null; then
    echo "Installing gocovmerge..."
    pushd $(mktemp -d)
    GOSUMDB=off go get -u github.com/wadey/gocovmerge
    popd
fi

echo "setting up test tmp folder"
[ -d "$FUNCT_TEST_TMPDIR" ] && rm -r "$FUNCT_TEST_TMPDIR"
mkdir -p "$FUNCT_TEST_TMPDIR"
# mkdir -p "$FUNCT_TEST_TMPDIR/output"
mkdir -p "$FUNCT_TEST_TMPDIR/kind-config"

echo "setting up test coverage folder"
[ -d "$FUNCT_TEST_COVERAGE" ] && rm -r "$FUNCT_TEST_COVERAGE"
mkdir -p "${FUNCT_TEST_COVERAGE}"

echo "generating kind configfile"

cat << EOF > "${FUNCT_TEST_TMPDIR}/kind-config/kind-config.yaml"
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraMounts:
  - hostPath: "${FUNCT_TEST_COVERAGE}"
    containerPath: /tmp/coverage
  extraPortMappings:
  - containerPort: 8080
    hostPort: 8080
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
networking:
  apiServerPort: 6443
EOF

echo "cluster configuration:"
cat ${FUNCT_TEST_TMPDIR}/kind-config/kind-config.yaml

echo "creating cluster"
kind create cluster --image=kindest/node:v1.21.10 --name $CLUSTER_NAME --config "${FUNCT_TEST_TMPDIR}/kind-config/kind-config.yaml"

# setup kubeconfig
kind get kubeconfig --name $CLUSTER_NAME > ${KIND_KUBECONFIG}

# load image if possible
kind load docker-image ${DOCKER_IMAGE_AND_TAG} --name=$CLUSTER_NAME -v 99 || echo "failed to load image locally, will use imagePullSecret"

echo "install cluster"
# setup cluster
make kind-cluster-setup
for dir in overlays/test/* ; do
  echo ">>>>>>>>>>>>>>>Executing test: $dir"

  # install clusterlifecycle-state-metrics
  echo "install managedcluster-import-controller"
  kubectl apply -k "$dir" --dry-run=client -o yaml | sed "s|REPLACE_IMAGE|${DOCKER_IMAGE_AND_TAG}|g" | kubectl apply -f -

  echo "Create ingress for functional test"
  kubectl apply -f test/functional/resources/ingress.yaml
  
  # patch image
  echo "Wait rollout"
  kubectl rollout status -n open-cluster-management deployment clusterlifecycle-state-metrics --timeout=180s
  
  # exit 1

  echo "run functional test..."
  set +e
  make functional-test
  ERR=$?
  if [ $ERR != 0 ]; then
    POD_NAMES=`kubectl get pods -n open-cluster-management -oname | grep clusterlifecycle-state-metrics`
    for p in $POD_NAMES; do
      echo "-----------------------${p}------------------------------"
      echo "$p" | xargs -L 1 kubectl logs -n open-cluster-management
    done;
    echo "Error: $ERR"
    exit $ERR
  fi
  set -e

  echo "remove deployment"
  kubectl delete --wait=true -k "$dir"
done;

echo "Wait 20 sec for copy to coverage files to external storage if setup"
sleep 20

echo "delete cluster"
kind delete cluster --name $CLUSTER_NAME

echo "Wait 20 sec for copy to coverage files from kind cluster to host"
sleep 20

ls $FUNCT_TEST_COVERAGE

if [ `find $FUNCT_TEST_COVERAGE -prune -empty 2>/dev/null` ]; then
  echo "no coverage files found. skipping"
else
  echo "merging coverage files"

  gocovmerge "${FUNCT_TEST_COVERAGE}/"* >> "${FUNCT_TEST_COVERAGE}/cover-functional.out"
  COVERAGE=$(go tool cover -func="${FUNCT_TEST_COVERAGE}/cover-functional.out" | grep "total:" | awk '{ print $3 }' | sed 's/[][()><%]/ /g')
  echo "-------------------------------------------------------------------------"
  echo "TOTAL COVERAGE IS ${COVERAGE}%"
  echo "-------------------------------------------------------------------------"

  go tool cover -html "${FUNCT_TEST_COVERAGE}/cover-functional.out" -o ${PROJECT_DIR}/test/functional/coverage/cover-functional.html
fi
