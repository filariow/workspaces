ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_FOLDER := $(ROOT_DIR)/bin
$(BIN_FOLDER):
	mkdir $(BIN_FOLDER)

GO ?= go
LD_FLAGS ?= -s -w

IMG ?= workspaces/rest-api:latest
NAMESPACE ?= workspaces-system

KUBECLI ?= kubectl
KUSTOMIZE ?= $(BIN_FOLDER)/kustomize
MOCKGEN ?= $(BIN_FOLDER)/mockgen
YQ ?= $(BIN_FOLDER)/yq
GOLANG_CI ?= $(GO) run -modfile $(shell dirname $(ROOT_DIR))/hack/tools/golang-ci/go.mod github.com/golangci/golangci-lint/cmd/golangci-lint
CONTROLLER_GEN ?= $(BIN_FOLDER)/controller-gen
CONTROLLER_TOOLS_VERSION ?= v0.14.0

.PHONY: kustomize
kustomize: $(KUSTOMIZE)
$(KUSTOMIZE):
	test -s $(BIN_FOLDER)/kustomize || \
		$(GO) build \
			-modfile $(shell dirname $(ROOT_DIR))/hack/tools/kustomize/go.mod \
			-o $(BIN_FOLDER)/kustomize \
			sigs.k8s.io/kustomize/kustomize/v4

.PHONY: mockgen
mockgen: $(MOCKGEN)
$(MOCKGEN): $(BIN_FOLDER)
	@cp hack/tools/mockgen $(BIN_FOLDER)/

.PHONY: yq
yq: $(YQ)
$(YQ): $(BIN_FOLDER)
	$(GO) build \
		-modfile $(shell dirname $(ROOT_DIR))/hack/tools/yq/go.mod \
		-o $(BIN_FOLDER)/yq \
		github.com/mikefarah/yq/v4

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(BIN_FOLDER)
	test -s $(BIN_FOLDER)/controller-gen && $(BIN_FOLDER)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(BIN_FOLDER) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: all
all: clean
	@:

.PHONY: clean
clean:
	@-rm -r $(BIN_FOLDER)

.PHONY: build
build:
	@$(GO) build \
		-ldflags '$(LD_FLAGS)' \
		-trimpath \
		-o $(BIN_FOLDER)/server \
		main.go

.PHONY: generate-mock
generate-mock: mockgen
	@echo $(GO) generate ./...
	@PATH=$(ROOT_DIR)/bin:${PATH} $(GO) generate ./...

.PHONY: run
run:
	$(GO) run main.go

.PHONY: lint
lint:
	$(GOLANG_CI) run ./...

.PHONY: vet
vet-moc-mockk:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

E2E_KUBECONFIG ?= /tmp/workspaces-e2e
E2E_CLUSTERNAME ?= workspaces-e2e
E2E_CLUSTER_PROVIDER ?= kwokctl

.PHONY: test-e2e-create-cluster
test-e2e-create-cluster:
	KUBECONFIG=$(E2E_KUBECONFIG) $(E2E_CLUSTER_PROVIDER) create cluster --name workspaces-e2e

.PHONY: test-e2e-delete-cluster
test-e2e-delete-cluster:
	KUBECONFIG=$(E2E_KUBECONFIG) $(E2E_CLUSTER_PROVIDER) delete cluster --name workspaces-e2e

.PHONY: test-e2e-prepare
test-e2e-prepare:
	$(MAKE) -C ../operator manifests
	KUBECONFIG=$(E2E_KUBECONFIG) $(MAKE) install
	KUBECONFIG=$(E2E_KUBECONFIG) kubectl apply\
		-f ../operator/config/crd/bases/workspaces.io_workspaces.yaml\
		-f https://raw.githubusercontent.com/codeready-toolchain/host-operator/master/config/crd/bases/toolchain.dev.openshift.com_spacebindings.yaml

.PHONY: test-e2e
test-e2e: test-e2e-prepare
	KUBECONFIG=$(E2E_KUBECONFIG) $(GO) test ./test/e2e/...

.PHONY: test
test: generate-mock
	$(GO) test ./...

.PHONY: test-with-coverage
test-with-coverage: generate-mock
	$(GO) test ./... -covermode=atomic -coverprofile cover.out

.PHONY: docker-build
docker-build:
	docker build -t ${IMG} -f Dockerfile ..

.PHONY: install
install: kustomize
	$(KUSTOMIZE) build ./config/crd | $(KUBECLI) apply -f -

.PHONY: deploy
deploy: kustomize yq
	YQ="$(YQ)" KUSTOMIZE="$(KUSTOMIZE)" ./hack/deploy.sh "$(NAMESPACE)" "$(IMG)"

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

