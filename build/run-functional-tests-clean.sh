#!/bin/bash

# Copyright (c) 2020 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

set -e
# set -x

CLUSTER_NAME=$PROJECT_NAME-functional-test

echo "delete clusters"
kind delete cluster --name ${CLUSTER_NAME}