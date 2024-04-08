#!/bin/bash

docker build -t ${IMG} .
docker login
docker push ${IMG}