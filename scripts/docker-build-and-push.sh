#!/bin/bash

version="${VER:-0.0.1}"
image="videlov/api-gateway-webhook-service:${version}"

docker build -t ${image} .
docker login
docker push ${image}