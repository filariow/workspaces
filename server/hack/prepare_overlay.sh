#!/bin/bash

set -e -o pipefail

_manifests_folder_relative="$1"
_overlay="$2"
_namespace="$3"
_image="$4"
_toolchain_host_ns="$5"
_domain="$6"

KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}

overlay_folder="$(realpath "${_manifests_folder_relative}/overlays/${_overlay}")"

# adding local overlay
mkdir "${overlay_folder}"
cd "${overlay_folder}"

${KUSTOMIZE} create
${KUSTOMIZE} edit add base "../../default"
${KUSTOMIZE} edit set namespace "${_namespace}"
${KUSTOMIZE} edit set image workspaces/rest-api="${_image}"
${KUSTOMIZE} edit add configmap rest-api-server-config \
    --behavior=replace \
    --from-literal=kubesaw.namespace="${_toolchain_host_ns}"

## patch ingress
cat << EOF > 'patch-ingress.yaml'
- op: replace
  path: /spec/rules/0/host
  value: api-workspaces.${_domain}
EOF
${KUSTOMIZE} edit add patch 'patch-ingress.yaml'
