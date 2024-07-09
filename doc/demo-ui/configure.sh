#!/bin/bash

CURRENT_DIR="$(dirname "$(readlink -f "$0")")"
HACK_DIR="$(realpath "${CURRENT_DIR}/../../hack")"

ROSA_CLUSTER=${ROSA_CLUSTER:-""}
DEV_SSO_NS="dev-sso"

# Install KubeSaw and Workspaces-Konflux
(cd "${HACK_DIR}/.." && "./hack/demo.sh")

# Update toolchainconfig
printf "patching toolchainconfig"
oc patch ToolchainConfig/config -n toolchain-host-operator --type=merge --patch-file=/dev/stdin << EOF
spec:
  host:
    automaticApproval:
      enabled: true
    environment: appstudio
    registrationService:
      environment: appstudio
      verification:
        enabled: false
        excludedEmailDomains: redhat.com,acme.com,user.us
    tiers:
      defaultSpaceTier: appstudio
      defaultUserTier: nodeactivation
      durationBeforeChangeTierRequestDeletion: 5s
EOF

# Setup the development SSO
"${CURRENT_DIR}/setup-dev-sso.sh" -sn "${DEV_SSO_NS}" -r "${ROSA_CLUSTER}"

# add JWKS Into traefik 
BASE_URL=$(oc get ingresses.config.openshift.io/cluster -o jsonpath='{.spec.domain}')
RHSSO_URL="https://keycloak-${DEV_SSO_NS}.$BASE_URL"

cfg=$(oc get configmap -n workspaces-system workspaces-traefik-sidecar-dynamic-config --output=jsonpath="{.data['config\.yaml']}" | \
  yq -e '.http.middlewares.jwt-authorizer.plugin.jwt.keys[0]="'"${RHSSO_URL}/auth/realms/sandbox-dev/protocol/openid-connect/certs"'"')

oc get configmap -n workspaces-system workspaces-traefik-sidecar-dynamic-config --output=json | \
  jq --arg cfg "$( yq -P <<< "$cfg" )" '.data."config.yaml" = $cfg' | \
  oc apply -f -


oc apply -f - << EOF
apiVersion: toolchain.dev.openshift.com/v1alpha1
kind: UserSignup
metadata:
  labels:
    toolchain.dev.openshift.com/email-hash: 98ea4248e45ba9df4b37dc992f022092
    toolchain.dev.openshift.com/state: approved
  name: alice
  namespace: toolchain-host-operator
spec:
  identityClaims:
    email: alice@user.us
    givenName: alice
    preferredUsername: alice
    sub: alice
EOF

oc apply -f - << EOF
apiVersion: toolchain.dev.openshift.com/v1alpha1
kind: UserSignup
metadata:
  labels:
    toolchain.dev.openshift.com/email-hash: a471fd6a49bba5d5ac74b1eaf08ea17a
    toolchain.dev.openshift.com/state: approved
  name: bob
  namespace: toolchain-host-operator
spec:
  identityClaims:
    email: bob@user.us
    givenName: bob
    preferredUsername: bob
    sub: bob
EOF

# oc wait clusteroperators.config.openshift.io authentication --for condition=Progressing=False --timeout 10m
# oc wait clusteroperators.config.openshift.io authentication --for condition=Available=True --timeout 10m

# TODO: Create UserSignups

