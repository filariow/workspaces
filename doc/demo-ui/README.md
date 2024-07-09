# Demo UI

## Prerequisites

Infra:
* ROSA Classic Cluster
* KubeSaw installed 
* Konflux-Workspaces installed
* Keycloak installed and configured

Resources:
* Keycloak
    * Users Alice and Bob
* KubeSaw
    * UserSignup Alice and Bob

**As SRE**
```bash
kubectl get usersignups.toolchain.dev.openshift.com -n toolchain-host-operator
```
```bash
kubectl get spaces.toolchain.dev.openshift.com -n toolchain-host-operator
```
```bash
kubectl get internalworkspaces.workspaces.konflux.io -n workspaces-system
```
```bash
kubectl get spacebinding -n toolchain-host-operator
```

## Workflows

### Users has private workspaces

```gherkin
Given Users Alice and Bob exist
And   Alice has a private workspace
And   Bob has a private workspace
Then  Alice and Bob can have details and access their own private workspaces
And   Alice and Bob can NOT have details NOR access others private workspaces
```

**As Alice**
```bash
cat $KUBECONFIG | yq '.users[0].user.token' | jq -R 'split(".") | .[1] | @base64d | fromjson'
kubectl get workspaces.workspaces.konflux.io -A
kubectl get workspaces.workspaces.konflux.io -n alice default -o yaml | yq
```

**As Bob**
```bash
cat $KUBECONFIG | yq '.users[0].user.token' | jq -R 'split(".") | .[1] | @base64d | fromjson'
kubectl get workspaces.workspaces.konflux.io -A
kubectl get workspaces.workspaces.konflux.io -n bob default -o yaml | yq
```

### Community Workspace

```gherkin
Given Users Alice and Bob exist
And   Alice has a private workspace
When  Alice changes visibility to community
Then  Bob can have details of Alice's private workspace
And   Bob can access Alice's private workspace
```

**As Alice**
```bash
TKN=$(cat $KUBECONFIG | yq '.users[0].user.token')
URL=$(cat $KUBECONFIG | yq '.clusters[0].cluster.server')
WS=$(kubectl get workspaces.workspaces.konflux.io -n alice -o json default | \
        jq '.spec.visibility="community"' | jq 'del(.status)')
curl -s -X PUT \
    -H "Authorization: Bearer ${TKN}" \
    -d "${WS}" \
    "${URL}/apis/workspaces.konflux.io/v1alpha1/namespaces/alice/workspaces/default" | jq 
```

```bash
kubectl get workspaces.workspaces.konflux.io -n alice -o yaml default | yq
```

**As SRE**
```bash
kubectl get spacebinding -n toolchain-host-operator
```

**As Bob**
```bash
kubectl get workspaces.workspaces.konflux.io -A
```

```bash
kubectl get workspaces.toolchain.dev.openshift.com alice -o yaml | yq
```

```bash
server="$(cat $KUBECONFIG | yq '.clusters[0].cluster.server')/workspaces/alice"
kubectl get pods -n alice-tenant --server "$server"
```
