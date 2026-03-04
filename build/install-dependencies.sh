#!/bin/bash -e

# Copyright (c) 2020 Red Hat, Inc.
# Copyright Contributors to the Open Cluster Management project

# Go tools
if ! which patter > /dev/null; then      echo "Installing patter ..."; go install github.com/apg/patter@latest; fi
if ! which gocovmerge > /dev/null; then  echo "Installing gocovmerge..."; go install github.com/alexfalkowski/gocovmerge/v2@v2.14.0; fi

# Image tools

# Check tools
