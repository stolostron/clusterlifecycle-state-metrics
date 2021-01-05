#!/bin/bash -e
###############################################################################
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################

# PARAMETERS
# $1 - Final image name and tag to be produced

#TO BE removed to activate script till END
echo "Please specify a COMPONENT_TYPE when running component/build-e2e to select the type of component you're building, or implement your own build-e2e.sh and set COMPONENT_BUILD_COMMAND to that file's path.";
exit 1;
#END 

#Example
export DOCKER_IMAGE_AND_TAG=${1}
export DOCKER_BUILD_TAG=e2e-test

docker build . \
$DOCKER_BUILD_OPTS \
-t $DOCKER_IMAGE:$DOCKER_BUILD_TAG \
-f build/Dockerfile-e2e

if [ ! -z "$DOCKER_IMAGE_AND_TAG" ]; then
    echo "Retagging image as $DOCKER_IMAGE_AND_TAG"
    docker tag $DOCKER_IMAGE:$DOCKER_BUILD_TAG "$DOCKER_IMAGE_AND_TAG"
fi
