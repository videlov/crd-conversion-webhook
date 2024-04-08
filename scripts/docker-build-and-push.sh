#!/bin/bash

IMG=videlov/api-gateway-webhook-service:0.0.17

docker build -t ${IMG} .
docker login
docker push ${IMG}