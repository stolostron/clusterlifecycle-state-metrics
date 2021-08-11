#!/bin/bash

# Copyright (c) 2021 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

set -o nounset
set -o pipefail

rm -rf multicloud-operators-foundation

echo "############  Cloning multicloud-operators-foundation"
git clone https://github.com/open-cluster-management/multicloud-operators-foundation.git

cd multicloud-operators-foundation || {
  printf "cd failed, multicloud-operators-foundation does not exist"
  return 1
}

# Deploying e2e env
make deploy-hub
make deploy-klusterlet
make deploy-foundation-hub
make deploy-foundation-webhook
make deploy-foundation-agent

echo "############  Cleanup"
cd ../ || exist
rm -rf multicloud-operators-foundation

echo "############  Finished installation!!!"
