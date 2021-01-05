#!/bin/bash -e
###############################################################################
# Copyright (c) 2020 Red Hat, Inc.
###############################################################################
# PARAMETERS
# $1 - Final image name and tag to be produced
export DOCKER_IMAGE_AND_TAG=${1}
# $2 - Final image name and tag for coverage image
export DOCKER_IMAGE_COVERAGE_AND_TAG=${2}

#TO BE removed to activate script till END
echo "Please specify a COMPONENT_TYPE when running component/build-coverage to select the type of component you're building, or implement your own build-coverage.sh and set COMPONENT_BUILD_COMMAND to that file's path.";
exit 1;
#END 

#Example
docker build . \
$DOCKER_BUILD_OPTS \
--build-arg DOCKER_BASE_IMAGE=$DOCKER_IMAGE_AND_TAG \
-t $DOCKER_IMAGE_COVERAGE:$DOCKER_BUILD_TAG \
-f build/Dockerfile-coverage

if [ ! -z "$DOCKER_IMAGE_COVERAGE_AND_TAG" ]; then
    echo "Retagging image as $DOCKER_IMAGE_COVERAGE_AND_TAG"
    docker tag $DOCKER_IMAGE_COVERAGE:$DOCKER_BUILD_TAG "$DOCKER_IMAGE_COVERAGE_AND_TAG"
fi