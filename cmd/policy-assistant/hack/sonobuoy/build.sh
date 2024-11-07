#!/usr/bin/env bash

set -xv
set -euo pipefail


IMAGE=mfenwick100/sonobuoy-policy-assistant:latest

docker build -t $IMAGE .
docker push $IMAGE
