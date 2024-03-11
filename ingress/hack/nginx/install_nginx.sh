#!/bin/bash

# add nginx repository in helm
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# installing the ingress
helm install ingress-nginx \
  ingress-nginx/ingress-nginx \
  -n workspaces-system \
  --create-namespace \
  -f values.yaml
