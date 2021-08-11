#!/bin/bash

# Copyright (c) 2021 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

set -o nounset
set -o pipefail

rm -rf multicloud-operators-foundation

echo "############  Cloning multicloud-operators-foundation"
git clone https://github.com/open-cluster-management/multicloud-operators-foundation.git

cd multicloud-operators-foundation
make clean-deploy
cd ../ || exist
rm -rf multicloud-operators-foundation

echo "############  Cloning registration-operator"
git clone https://github.com/open-cluster-management/registration-operator.git

cd registration-operator
make clean-deploy
cd ../ || exist
rm -rf registration-operator

echo "############  Finished clean!!!"