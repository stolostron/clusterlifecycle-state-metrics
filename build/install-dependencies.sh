#!/bin/bash -e

# Copyright (c) 2020 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

export GO111MODULE=off

# Go tools
_OS=$(go env GOOS)
_ARCH=$(go env GOARCH)
KubeBuilderVersion="2.3.0"

if ! which patter > /dev/null; then      echo "Installing patter ..."; go get -u github.com/apg/patter; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go get -u github.com/wadey/gocovmerge; fi
if ! which go-bindata > /dev/null; then
	echo "Installing go-bindata..."
	cd $(mktemp -d) && GOSUMDB=off go get -u github.com/go-bindata/go-bindata/...
fi
go-bindata --version

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
