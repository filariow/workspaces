#!/usr/bin/env bash

set -e -o pipefail

user_help() {
    echo "Deploy in cluster keycloak and configure registration service to use it."
    echo "options:"
    echo "-sn, --sso-ns  namespace where the SSO provider will be installed"
    echo "-r,  --rosa-cluster      The name of the target ROSA cluster (optional). Use this flag to properly configure OAuth on ROSA Classic clusters."
    echo "-h,  --help              To show this help text"
    echo ""
}

read_arguments() {
    while test $# -gt 0; do
           case "$1" in
                -h|--help)
                    user_help
                    exit 0
                    ;;
                -sn|--sso-ns)
                    shift
                    DEV_SSO_NS=$1
                    shift
                    ;;
                -r|--rosa-cluster)
                    shift
                    CLUSTER_NAME="$1"
                    shift
                    ;;
                *)
                   echo "$1 is not a recognized flag!" >> /dev/stderr
                   user_help
                   exit 1
                   ;;
          esac
    done

    [[ -n "${DEV_SSO_NS}" ]] || { printf "SSO namespace is required\n"; user_help; exit 1; }
}

check_commands()
{
    for cmd in "$@"
    do
        check_command "${cmd}"
    done
}

check_command()
{
    command -v "$1" > /dev/null && return 0

    printf "please install '%s' before running this script\n" "$1"
    exit 1
}

setup_oauth_internal()
{
  set -e -o pipefail

  RHSSO_URL="$1"

  printf "configuring OAuth authentication for keycloak '%s'\n" "${RHSSO_URL}"
  KEYCLOAK_SECRET="$2" envsubst < "dev-sso/openid-secret.yaml" | oc apply -f - 
  
  printf "applying patch for oauths configuration\n"
  oc patch oauths.config.openshift.io/cluster --type=json --patch='
  [
    {
      "op": "add",
      "path": "/spec/identityProviders/-",
      "value": {
        "mappingMethod": "lookup",
        "name": "rhd",
        "openID": {
          "ca": {
            "name": ""
          },
          "claims": {
            "preferredUsername": [ "username" ],
            "email": [ "email" ],
            "name": [ "username" ]
          },
          "clientID": "sandbox",
          "clientSecret": {
            "name": "openid-client-secret-sandbox"
          },
          "issuer": "'"${RHSSO_URL}"'/auth/realms/sandbox-dev"
        },
        "type": "OpenID"
      }
    }
  ]
  '
}

setup_oauth_rosa()
{
  ISSUER_URL="$1/auth/realms/sandbox-dev"
  KEYCLOAK_SECRET="$2"
  CLUSTER_NAME="$3"

  printf "Setting up OAuth in ROSA cluster '%s' for issuer '%s'\n" "${CLUSTER_NAME}" "${ISSUER_URL}"
  rosa create idp \
    --cluster="${CLUSTER_NAME}" \
    --type='openid' \
    --client-id='sandbox' \
    --client-secret="${KEYCLOAK_SECRET}" \
    --mapping-method='lookup' \
    --issuer-url="${ISSUER_URL}" \
    --email-claims='email' \
    --name-claims='username' \
    --username-claims='username' \
    --name='rhd' \
    --yes
}

setup_oauth()
{
  set -e

  if [[ -n "$3" ]]; then 
    setup_oauth_rosa "$1" "$2" "$3"
  else
    setup_oauth_internal "$1" "$2"
  fi
}

read_arguments "$@"

if [[ -n "${CI}" ]]; then
    set -ex
else
    set -e
fi


check_commands yq oc base64 openssl

parent_path=$( cd "$(dirname "${BASH_SOURCE[0]}")" ; pwd -P )
cd "${parent_path}"

printf "creating %s namespace\n" "${DEV_SSO_NS}"
DEV_SSO_NS=${DEV_SSO_NS} envsubst < "dev-sso/namespace.yaml" | oc apply -f -

# Install rhsso operator
SUBSCRIPTION_NAME=${DEV_SSO_NS}
printf "installing RH SSO operator\n"
DEV_SSO_NS=${DEV_SSO_NS} SUBSCRIPTION_NAME=${SUBSCRIPTION_NAME} envsubst < "dev-sso/rhsso-operator.yaml" | oc apply -f -

source ./wait-until-is-installed.sh "-crd keycloak.org -cs '' -n ${DEV_SSO_NS} -s ${SUBSCRIPTION_NAME}"

printf "installing dev Keycloak in namespace %s\n" "${DEV_SSO_NS}"
KEYCLOAK_SECRET=$(openssl rand -base64 32)
export KEYCLOAK_SECRET
DEV_SSO_NS=${DEV_SSO_NS} KEYCLOAK_SECRET=${KEYCLOAK_SECRET} envsubst < "dev-sso/keycloak.yaml" | oc apply -f -

while ! oc get statefulset -n "${DEV_SSO_NS}" keycloak &> /dev/null ; do
    printf "waiting for keycloak statefulset in %s to exist...\n" "${DEV_SSO_NS}"
    sleep 10
done

printf "waiting for keycloak in %s to be ready...\n" "${DEV_SSO_NS}"
TIMEOUT=200s
oc wait --for=jsonpath='{.status.ready}'=true keycloak/sandbox-dev -n "${DEV_SSO_NS}" --timeout "${TIMEOUT}"  || \
  { oc get keycloak sandbox-dev -n "${DEV_SSO_NS}" -o yaml && exit 1; }

oc rollout status statefulset -n "${DEV_SSO_NS}" keycloak --timeout 20m
  
BASE_URL=$(oc get ingresses.config.openshift.io/cluster -o jsonpath='{.spec.domain}')
RHSSO_URL="https://keycloak-${DEV_SSO_NS}.${BASE_URL}"

printf "Setup OAuth\n"
setup_oauth "${RHSSO_URL}" "${KEYCLOAK_SECRET}" "${CLUSTER_NAME}"

## Configure toolchain to use the internal keycloak
printf "patching toolchainconfig\n"
oc patch ToolchainConfig/config -n toolchain-host-operator --type=merge --patch-file=/dev/stdin << EOF
spec:
  host:
    registrationService:
      auth:
        authClientConfigRaw: '{
                  "realm": "sandbox-dev",
                  "auth-server-url": "${RHSSO_URL}/auth",
                  "ssl-required": "none",
                  "resource": "sandbox-public",
                  "clientId": "sandbox-public",
                  "public-client": true,
                  "confidential-port": 0
                }'
        authClientLibraryURL: ${RHSSO_URL}/auth/js/keycloak.js
        authClientPublicKeysURL: ${RHSSO_URL}/auth/realms/sandbox-dev/protocol/openid-connect/certs
EOF

# Restart the registration-service to ensure the new configuration is used
oc delete pods -n toolchain-host-operator --selector=name=registration-service

KEYCLOAK_ADMIN_PASSWORD=$(oc get secrets -n "${DEV_SSO_NS}" credential-sandbox-dev -o jsonpath='{.data.ADMIN_PASSWORD}' | base64 -d)
printf "================================================= DEV SSO ACCESS ==============================================================================================\n"
printf "to login into keycloak use user 'admin' and password '%s' at '%s/auth'\n" "${KEYCLOAK_ADMIN_PASSWORD}" "${RHSSO_URL}"
printf "use user 'alice@user.us' with password 'alice' to login at 'https://registration-service-toolchain-host-operator.%s'\n" "${BASE_URL}"
printf "use user 'bob@user.us' with password 'bob' to login at 'https://registration-service-toolchain-host-operator.%s'\n" "${BASE_URL}"
printf "================================================================================================================================================================\n"
