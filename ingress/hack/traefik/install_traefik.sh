#!/bin/bash

set -e -x -o pipefail

# add the traefik repo
helm repo add traefik https://traefik.github.io/charts
helm repo update

# install traefik
helm install traefik traefik/traefik \
  --namespace workspaces-system --create-namespace \
  -f values.yaml
