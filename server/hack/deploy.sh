#!/bin/bash

set -e -o pipefail

_namespace="$1"
_image="$2"

LOCATION=$(readlink -f "$0")
DIR=$(dirname "${LOCATION}")
ROOT_DIR="$(realpath "${DIR}"/../..)"
KUBECLI=${KUBECLI:-kubectl}
KUSTOMIZE=${KUSTOMIZE:-kustomize}

# retrieving toolchain-host namespace
_toolchain_host_ns=$(${KUBECLI} get namespaces -o name | grep toolchain-host | cut -d'/' -f2 | head -n 1)
if [[ -z "${_toolchain_host_ns}" ]]; then
    _toolchain_host_ns="toolchain-host_operator"
fi

# retrieve OCP cluster domain
_domain=$(oc get dns cluster -o jsonpath='{ .spec.baseDomain }')

# prepare temporary folder
_f=$(mktemp --directory /tmp/workspaces-rest.XXXXX)
cp -r "${ROOT_DIR}/server/manifests" "${_f}/manifests"

./prepare_overlay.sh \
  "${ROOT_DIR}/manifests" \
  "local" \
  "${_namespace}" \
  "${_image}" \
  "${_toolchain_host_ns}" \
  "${_domain}"

# apply manifests
${KUSTOMIZE} build . | ${KUBECLI} apply -f -

# cleanup
rm -r "${_f}"
