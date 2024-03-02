#!/usr/bin/env bash

set -e

# build frpc
docker build -f ./frpc/Dockerfile -t opennaslab/frpc:latest ./frpc

# build frpc4-manager
docker build -f ./frpc4-manager/Dockerfile -t opennaslab/frpc4-manager:latest ./frpc4-manager
