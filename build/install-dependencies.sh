#!/bin/bash -e

# Copyright (c) 2020 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

# Go tools
_OS=$(go env GOOS)
_ARCH=$(go env GOARCH)
KubeBuilderVersion="2.3.0"

if ! which patter > /dev/null; then      echo "Installing patter ..."; go install github.com/apg/patter@latest; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go install github.com/alexfalkowski/gocovmerge/v2@v2.14.0; fi

# Build tools
if ! which kubebuilder > /dev/null; then
   # Install kubebuilder for unit test
   echo "Install Kubebuilder components for test framework usage!"

   # download kubebuilder and extract it to tmp
   curl -L https://go.kubebuilder.io/dl/"$KubeBuilderVersion"/"${_OS}"/"${_ARCH}" | tar -xz -C /tmp/

   # move to a long-term location and put it on your path
   # (you'll need to set the KUBEBUILDER_ASSETS env var if you put it somewhere else)
   sudo mv /tmp/kubebuilder_"$KubeBuilderVersion"_"${_OS}"_"${_ARCH}" $KUBEBUILDER_HOME
fi
# Image tools

# Check tools
