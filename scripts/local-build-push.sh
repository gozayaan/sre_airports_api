#!/bin/bash

# This script locally builds image and pushes to docker hub

CURRENT_TAG=v1.7
PREVIOUS_TAG=v1.6
DOCKERHUB_USERNAME=YourName

# build local image for codebase test
DOCKER_BUILDKIT=1 docker build -f Dockerfile -t bd-airports ./src

# update tag
docker tag bd-airports ${DOCKERHUB_USERNAME}/bd-airports:${CURRENT_TAG}

# push
docker push ${DOCKERHUB_USERNAME}/bd-airports:${CURRENT_TAG}
