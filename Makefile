# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

SHELL := bash -eu -o pipefail

# VERSION defines the project version for the bundle.
# Update this value when you upgrade the version of your project.
# To re-generate a bundle for another specific version without changing the standard setup, you can:
# - use the VERSION as arg of the bundle target (e.g make bundle VERSION=0.0.2)
# - use environment variables to overwrite this value (e.g export VERSION=0.0.2)
VERSION            ?= $(shell cat VERSION)
GIT_HASH_SHORT     := $(shell git rev-parse --short=8 HEAD)
VERSION_DEV_SUFFIX := ${GIT_HASH_SHORT}

FUZZTIME ?= 60s

# Add an identifying suffix for `-dev` builds only.
# Release build versions are verified as unique by the CI build process.
ifeq ($(findstring -dev,$(VERSION)), -dev)
    VERSION := $(VERSION)-$(VERSION_DEV_SUFFIX)
endif

HELM_VERSION ?= ${VERSION}

REGISTRY         ?= 080137407410.dkr.ecr.us-west-2.amazonaws.com
REGISTRY_NO_AUTH ?= edge-orch
REPOSITORY       ?= cluster

DOCKER_TAG              ?= ${VERSION}
DOCKER_IMAGE_MANAGER    ?= ${REGISTRY}/${REGISTRY_NO_AUTH}/${REPOSITORY}/capi-provider-intel-manager:${DOCKER_TAG}
DOCKER_IMAGE_SOUTHBOUND ?= ${REGISTRY}/${REGISTRY_NO_AUTH}/${REPOSITORY}/capi-provider-intel-southbound:${DOCKER_TAG}

## Labels to add Docker/Helm/Service CI meta-data.
LABEL_SOURCE       ?= $(shell git remote get-url $(shell git remote))
LABEL_REVISION     = $(shell git rev-parse HEAD)
LABEL_CREATED      ?= $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")

DOCKER_LABEL_ARGS  ?= \
	--build-arg org_oci_version="${VERSION}" \
	--build-arg org_oci_source="${LABEL_SOURCE}" \
	--build-arg org_oci_revision="${LABEL_REVISION}" \
	--build-arg org_oci_created="${LABEL_CREATED}"

# Docker Build arguments
DOCKER_BUILD_ARGS ?= \
	--build-arg http_proxy="$(http_proxy)" --build-arg https_proxy="$(https_proxy)" \
	--build-arg no_proxy="$(no_proxy)" --build-arg HTTP_PROXY="$(http_proxy)" \
	--build-arg HTTPS_PROXY="$(https_proxy)" --build-arg NO_PROXY="$(no_proxy)"

# Image URLs to use all building/pushing image targets
IMG_MANAGER := ${DOCKER_IMAGE_MANAGER}
IMG_SOUTHBOUND := ${DOCKER_IMAGE_SOUTHBOUND}

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.31.0

## Virtual environment name
VENV_NAME = venv-env

# GoCov versions
GOLANG_GOCOV_VERSION := latest
GOLANG_GOCOV_XML_VERSION := latest

TEST_PATHS := ./internal/... ./pkg/...
GO_TESTABLES = $(shell go list ${TEST_PATHS}|grep -vE pkg/api)

# This controls if the Inventory Stub should be enabled or not while deploying Intel Provider Manager. The value true will enable the Inventory Stub.
USE_INV_STUB ?= false

# This controls if gRPC Middleware Stub should be enabled or not while deploying Intel Southbound Handler. The value true will enable the stub middleware.
# The stub middleware bypasses RBAC and inject a dummy projectId. This is to be used only for tests.
USE_GRPC_MIDDLEWARE_STUB ?= false

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOARCH       := $(shell go env GOARCH)
GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre=all -N -l" -asmflags="all=-spectre=all" -ldflags="all=-s -w"
ifeq ($(GOARCH),arm64)
  GOEXTRAFLAGS := -trimpath -gcflags="all=-spectre= -N -l" -asmflags="all=-spectre=" -ldflags="all=-s -w"
endif
ifeq ($(GO_VENDOR),true)
	GOEXTRAFLAGS := -mod=vendor $(GOEXTRAFLAGS)
endif
# For protobuf generation
GOPATH=$(shell go env GOPATH)
GO_BUILD_CMD = go
PROTOC_GEN_VALIDATE_VERSION := v1.1.0

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest gocov helm-test ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ${GO_TESTABLES} -race -gcflags -l -coverprofile cover.out -covermode atomic -short
	${GOBIN}/gocov convert cover.out | ${GOBIN}/gocov-xml > coverage.xml
	go tool cover -html=cover.out -o coverage.html

# TODO(user): To use a different vendor for e2e tests, modify the setup under 'tests/e2e'.
# The default setup assumes Kind is pre-installed and builds/loads the Manager Docker image locally.
# Prometheus and CertManager are installed by default; skip with:
# - PROMETHEUS_INSTALL_SKIP=true
# - CERT_MANAGER_INSTALL_SKIP=true
.PHONY: test-e2e
test-e2e: manifests generate fmt vet ## Run the e2e tests. Expected an isolated environment using Kind.
	@command -v kind >/dev/null 2>&1 || { \
		echo "Kind is not installed. Please install Kind manually."; \
		exit 1; \
	}
	@kind get clusters | grep -q 'kind' || { \
		echo "No Kind cluster is running. Please start a Kind cluster before running the e2e tests."; \
		exit 1; \
	}
	go test ./test/e2e/ -v -ginkgo.v

.PHONY: fuzz
fuzz: ## Run Fuzz tests
	hack/fuzz_all.sh ${FUZZTIME}

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run --timeout 15m

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix --timeout 15m

.PHONY: dependency-check-ci
dependency-check: ## Empty for now

##@ Build

.PHONY: build
build: build-manager build-southbound ## Build manager and southbound handler binaries.

.PHONY: build-manager
build-manager: ## Build manager binary.
	go build -o bin/manager ${GOEXTRAFLAGS} cmd/manager/main.go

.PHONY: build-southbound
build-southbound: ## Build southbound handler binary.
	go build -o bin/southbound_handler ${GOEXTRAFLAGS} cmd/southbound/main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./cmd/manager/main.go

.PHONY: vendor
vendor:  ## Build vendor directory of module dependencies.
	GOPRIVATE=github.com/open-edge-platform/* go mod vendor

# If you wish to build the manager image targeting other platforms you can use the --platform flag.
# (i.e. docker build --platform linux/arm64). However, you must enable docker buildKit for it.
# More info: https://docs.docker.com/develop/develop-images/build_enhancements/
.PHONY: docker-build
docker-build: vendor ## Build docker images.
	$(CONTAINER_TOOL) build -t ${IMG_MANAGER} -f Dockerfile.manager . ${DOCKER_BUILD_ARGS} ${DOCKER_LABEL_ARGS}
	$(CONTAINER_TOOL) build -t ${IMG_SOUTHBOUND} -f Dockerfile.southbound . ${DOCKER_BUILD_ARGS} ${DOCKER_LABEL_ARGS}

.PHONY: docker-push
docker-push: ## Push docker images.
	$(CONTAINER_TOOL) push ${IMG_MANAGER}
	$(CONTAINER_TOOL) push ${IMG_SOUTHBOUND}

.PHONY: docker-list
docker-list: ## Print name of docker container images
	@echo "images:"
	@echo "  capi-provider-intel-manager:"
	@echo "    name: '$(IMG_MANAGER)'"
	@echo "    version: '$(VERSION)'"
	@echo "    gitTagPrefix: 'v'"
	@echo "    buildTarget: 'docker-build'"
	@echo "  capi-provider-intel-southbound:"
	@echo "    name: '$(IMG_SOUTHBOUND)'"
	@echo "    version: '$(VERSION)'"
	@echo "    gitTagPrefix: 'v'"
	@echo "    buildTarget: 'docker-build'"

# PLATFORMS defines the target platforms for the manager image be built to provide support to multiple
# architectures. (i.e. make docker-buildx IMG=myregistry/mypoperator:0.0.1). To use this option you need to:
# - be able to use docker buildx. More info: https://docs.docker.com/build/buildx/
# - have enabled BuildKit. More info: https://docs.docker.com/develop/develop-images/build_enhancements/
# - be able to push the image to your registry (i.e. if you do not set a valid value via IMG=<myregistry/image:<tag>> then the export will fail)
# To adequately provide solutions that are compatible with multiple platforms, you should consider using this option.
PLATFORMS ?= linux/arm64,linux/amd64,linux/s390x,linux/ppc64le
.PHONY: docker-buildx
docker-buildx: ## Build and push docker image for the manager for cross-platform support
	# copy existing Dockerfile and insert --platform=${BUILDPLATFORM} into Dockerfile.cross, and preserve the original Dockerfile
	sed -e '1 s/\(^FROM\)/FROM --platform=\$$\{BUILDPLATFORM\}/; t' -e ' 1,// s//FROM --platform=\$$\{BUILDPLATFORM\}/' Dockerfile > Dockerfile.cross
	- $(CONTAINER_TOOL) buildx create --name cluster-api-provider-intel-builder
	$(CONTAINER_TOOL) buildx use cluster-api-provider-intel-builder
	- $(CONTAINER_TOOL) buildx build --push --platform=$(PLATFORMS) --tag ${IMG} -f Dockerfile.cross .
	- $(CONTAINER_TOOL) buildx rm cluster-api-provider-intel-builder
	rm Dockerfile.cross

.PHONY: build-installer
build-installer: manifests generate kustomize ## Generate a consolidated YAML with CRDs and deployment.
	mkdir -p dist
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG_MANAGER}
	$(KUSTOMIZE) build config/default > dist/install.yaml

HELM_DIRS = $(shell find ./deployment/charts -maxdepth 1 -mindepth 1 -type d -print )
HELM_PKGS = $(shell find . -name "*.tgz" -maxdepth 1 -mindepth 1 -type f -print )

.PHONY: helm-clean
helm-clean: ## Clean helm chart build annotations.
	for d in $(HELM_DIRS); do \
		yq eval -i '.version = "0.0.0"' $$d/Chart.yaml; \
		yq eval -i 'del(.appVersion)' $$d/Chart.yaml; \
		yq eval -i 'del(.annotations.revision)' $$d/Chart.yaml; \
		yq eval -i 'del(.annotations.created)' $$d/Chart.yaml; \
	done
	rm -f $(HELM_PKGS)

.PHONY: helm-test
helm-test: ## Template the charts.
	for d in $(HELM_DIRS); do \
		helm template intel $$d > /dev/null; \
	done

.PHONY: helm-build
helm-build: ## Package helm charts.
	for d in $(HELM_DIRS); do \
		yq eval -i '.version = "${HELM_VERSION}"' $$d/Chart.yaml; \
		yq eval -i '.appVersion = "${VERSION}"' $$d/Chart.yaml; \
		yq eval -i '.annotations.revision = "${LABEL_REVISION}"' $$d/Chart.yaml; \
		yq eval -i '.annotations.created = "${LABEL_CREATED}"' $$d/Chart.yaml; \
		helm package --app-version=${VERSION} --version=${HELM_VERSION} --debug -u $$d; \
	done

.PHONY: helm-push
helm-push: ## Push helm charts.
	for c in $(HELM_PKGS); do helm push $$c oci://${REGISTRY}/${REGISTRY_NO_AUTH}/${REPOSITORY}/charts; done

.PHONY: helm-list
helm-list:
	@echo "charts:"
	@for d in $(HELM_DIRS); do \
    cname=$$(grep "^name:" "$$d/Chart.yaml" | cut -d " " -f 2) ;\
    echo "  $$cname:" ;\
    echo -n "    "; grep "^version" "$$d/Chart.yaml"  ;\
    echo "    gitTagPrefix: 'v'" ;\
    echo "    outDir: '.'" ;\
  done

.PHONY: helm-install
helm-install: helm-build ## Install Helm charts in the local cluster
	kubectl apply -f config/crd/deps/cluster.edge-orchestrator.intel.com_clusterconnects.yaml
	helm upgrade --install intel-infra-provider-crds intel-infra-provider-crds-${HELM_VERSION}.tgz
	helm upgrade --install intel-infra-provider intel-infra-provider-${HELM_VERSION}.tgz --set metrics.serviceMonitor.enabled=false \
           --set manager.extraArgs.use-inv-stub=${USE_INV_STUB} --set=southboundApi.extraArgs.useGrpcStubMiddleware=${USE_GRPC_MIDDLEWARE_STUB}

.PHONY: helm-uninstall
helm-uninstall: ## Uninstall Helm charts from the local cluster
	helm uninstall intel-infra-provider || true
	helm uninstall intel-infra-provider-crds || true

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG_MANAGER}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f -

.PHONY: undeploy
undeploy: kustomize ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy-chart
deploy-chart:
	helm install intel-infra-provider-crds deployment/charts/intel-infra-provider-crds
	helm install intel-infra-provider deployment/charts/intel-infra-provider

.PHONY: undeploy-chart
undeploy-chart:
	helm uninstall intel-infra-provider

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v5.5.0
CONTROLLER_TOOLS_VERSION ?= v0.16.4
ENVTEST_VERSION ?= release-0.19
GOLANGCI_LINT_VERSION ?= v1.64.6

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

##@ Standard targets

.PHONY: cobertura
cobertura:
	go install github.com/boumenot/gocover-cobertura@latest

.PHONY: gocov
gocov:
	go install github.com/axw/gocov/gocov@${GOLANG_GOCOV_VERSION}
	go install github.com/AlekSi/gocov-xml@${GOLANG_GOCOV_XML_VERSION}

$(VENV_NAME): requirements.txt
	echo "Creating virtualenv $@"
	python3 -m venv $@;\
	. ./$@/bin/activate; set -u;\
	python3 -m pip install --upgrade pip;\
	python3 -m pip install -r requirements.txt

.PHONY: license
license: $(VENV_NAME) ## Check licensing with the reuse tool.
	## Check licensing with the reuse tool.
	. ./$</bin/activate; set -u;\
	reuse --version;\
	reuse --root . lint

.PHONY: golint
golint: lint ## Lint Go files.

.PHONY: helmlint
helmlint: ## Lint Helm charts.
	helm lint ./deployment/charts/*

YAML_FILES := $(shell find . -path './venv-env' -prune -o -type f \( -name '*.yaml' -o -name '*.yml' \) -print )

.PHONY: yamllint
yamllint: $(VENV_NAME) ## Lint YAML files.
	. ./$</bin/activate; set -u;\
	yamllint --version;\
	yamllint -c .yamllint -s $(YAML_FILES)

.PHONY: mocks
mocks: ## Generate mock files for unit test using mockery
	mockery --version || go install github.com/vektra/mockery/v2@latest
	mockery

##@ Southbound Handler targets

.PHONY: proto-generate
proto-generate: ## Generate protobuf code.
	cd pkg/api && protoc \
		-I proto/.  -I ${GOPATH}/pkg/mod/github.com/envoyproxy/protoc-gen-validate@${PROTOC_GEN_VALIDATE_VERSION} \
		--go_out=./proto --go_opt=paths=source_relative --go-grpc_out=./proto --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=allow_delete_body=true:. --validate_out=lang=go:./proto --grpc-gateway_opt=paths=source_relative \
		proto/cluster_orchestrator_southbound.proto

.PHONY: protogen-deps
protogen-deps: ## Install dependencies for protobuf generation.
	$(GO_BUILD_CMD) install github.com/google/gnostic@latest
	$(GO_BUILD_CMD) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO_BUILD_CMD) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO_BUILD_CMD) install github.com/googleapis/gnostic-grpc@latest
	$(GO_BUILD_CMD) install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
	$(GO_BUILD_CMD) install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest
	$(GO_BUILD_CMD) install github.com/envoyproxy/protoc-gen-validate@${PROTOC_GEN_VALIDATE_VERSION}

.PHONY: sb-run
sb-run: fmt vet ## Run the southbound handler.
	go run cmd/southbound/main.go

.PHONY: clean
clean: ## Clean build artifacts.
	rm -rf bin/southbound_handler bin/manager coverage.*

##@ Setting up a simple demo.

KIND_CLUSTER ?= kind

.PHONY: kind-create
kind-create: ## Create a Kind cluster.
	kind create cluster --config test/demo/kind-cluster-with-extramounts.yaml

.PHONY: kind-delete
kind-delete: ## Delete the Kind cluster.
	kind delete cluster -n ${KIND_CLUSTER} || true

RS_REGISTRY ?= registry-rs.edgeorchestration.intel.com
RS_IMG_MANAGER    ?= ${RS_REGISTRY}/${REGISTRY_NO_AUTH}/${REPOSITORY}/capi-provider-intel-manager:${DOCKER_TAG}
RS_IMG_SOUTHBOUND ?= ${RS_REGISTRY}/${REGISTRY_NO_AUTH}/${REPOSITORY}/capi-provider-intel-southbound:${DOCKER_TAG}

.PHONY: kind-load
kind-load: docker-build
	docker tag ${IMG_MANAGER} ${RS_IMG_MANAGER}
	docker tag ${IMG_SOUTHBOUND} ${RS_IMG_SOUTHBOUND}
	kind load docker-image ${RS_IMG_MANAGER} -n ${KIND_CLUSTER}
	kind load docker-image ${RS_IMG_SOUTHBOUND} -n ${KIND_CLUSTER}

.PHONY: clusterctl-init
clusterctl-init:
	CLUSTER_TOPOLOGY=true clusterctl init --core cluster-api:v1.8.5 --bootstrap rke2:v0.9.0 --control-plane rke2:v0.9.0
	kubectl wait pod --all --for=condition=Ready --namespace=capi-system --timeout=300s
	kubectl wait pod --all --for=condition=Ready --namespace=rke2-bootstrap-system --timeout=300s
	kubectl wait pod --all --for=condition=Ready --namespace=rke2-control-plane-system --timeout=300s

# The namespace (i.e., Project ID) where the demo resources are installed
PROJECTID ?= "53cd37b9-66b2-4cc8-b080-3722ed7af64a"
NAMESPACE := ${PROJECTID}

# The node GUID of the RKE2 node where the demo resources are installed
NODEGUID ?= 12345678-1234-1234-1234-123456789012

.PHONY: demo-setup
demo-setup: kind-delete kind-create kind-load helm-install clusterctl-init ## Setup a demo environment in Kind.
	kubectl apply -f config/crd/deps/cluster.edge-orchestrator.intel.com_clusterconnects.yaml || true

.PHONY: demo
demo: ## Run the cluster creation demo.
	kubectl create ns connect-gateway-secrets || true
	NAMESPACE=${NAMESPACE} NODEGUID=${NODEGUID} envsubst < test/demo/rke2-intel-example.yaml | kubectl apply -f -
	# Wait for ClusterConnection to exist
	until kubectl -n ${NAMESPACE} get clusterconnect ${NAMESPACE}-intel-rke2-test; do echo "Waiting for ClusterConnect"; sleep 1; done
	kubectl get pods -A | grep "edge-connect-gateway" || kubectl patch -n ${NAMESPACE} ClusterConnect ${NAMESPACE}-intel-rke2-test --type=merge --subresource status --patch 'status: {controlPlaneEndpoint: {host: "foo.com", port: 12345}, ready: true}'
	# Wait for IntelMachine to exist
	until (( `kubectl -n ${NAMESPACE} get intelmachine -o yaml | yq '.items | length'` > 0 )); do echo "Waiting for IntelMachine"; sleep 1; done
	#kubectl annotate -n ${NAMESPACE} intelmachine `kubectl get intelmachine -n ${NAMESPACE} -o=jsonpath='{.items..metadata.name}'` intelmachine.infrastructure.cluster.x-k8s.io/agent-status="active"
	sleep 5
	clusterctl describe cluster intel-rke2-test -n ${NAMESPACE}

.PHONY: demo-clusterclass
demo-clusterclass: ## Run the cluster creation demo using clusterclass.
	kubectl create ns connect-gateway-secrets || true
	NAMESPACE=${NAMESPACE} NODEGUID=${NODEGUID} envsubst < test/demo/rke2-intel-clusterclass-example.yaml | kubectl apply -f -

.PHONY: demo-cleanup
demo-cleanup: ## Remove the demo resources.
	kubectl delete clusters.cluster.x-k8s.io intel-rke2-test -n ${NAMESPACE} --wait=false || true
	sleep 1
	kubectl patch -n ${NAMESPACE} intelcluster intel-rke2-test  --type=merge --patch '{"metadata":{"finalizers":null}}' || true
	kubectl patch -n ${NAMESPACE} intelmachine `kubectl get intelmachine -n ${NAMESPACE} -o=jsonpath='{.items..metadata.name}'` --type=merge --patch '{"metadata":{"finalizers":null}}' || true
	kubectl delete -n ${NAMESPACE} intelmachinebindings --all --ignore-not-found=true || true
	kubectl delete clusterconnects --all --ignore-not-found=true || true

##@ Developing in Coder env.

CODER_DIR ?= ~/orch-deploy
CHART_NS  := orch-cluster

.PHONY: coder-redeploy
coder-redeploy: helm-build kind-load ## Redeploy local charts in the Coder KinD cluster
	kubectl config use-context kind-kind
	kubectl patch application -n dev root-app --type=merge -p '{"spec":{"syncPolicy":{"automated":{"selfHeal":false}}}}'
	kubectl delete application -n dev intel-infra-provider --ignore-not-found=true
	kubectl delete application -n dev intel-infra-provider-crd --ignore-not-found=true
	kubectl delete crd intelmachines.infrastructure.cluster.x-k8s.io --ignore-not-found=true
	kubectl delete crd intelmachinetemplates.infrastructure.cluster.x-k8s.io --ignore-not-found=true
	kubectl delete crd intelmachinebindings.infrastructure.cluster.x-k8s.io --ignore-not-found=true
	kubectl delete crd intelclusters.infrastructure.cluster.x-k8s.io --ignore-not-found=true
	kubectl delete crd intelclustertemplates.infrastructure.cluster.x-k8s.io --ignore-not-found=true
	helm upgrade --install -n ${CHART_NS} intel-infra-provider-crds intel-infra-provider-crds-${HELM_VERSION}.tgz
	helm upgrade --install -n ${CHART_NS} intel-infra-provider intel-infra-provider-${HELM_VERSION}.tgz \
		--set metrics.serviceMonitor.enabled=true \
		--set manager.extraArgs.use-inv-stub=${USE_INV_STUB} \
		--set manager.extraEnv[0].name=TENANT_ID \
		--set manager.extraEnv[0].value=${PROJECTID} \
		--set manager.extraEnv[1].name=HOST_ID \
		--set manager.extraEnv[1].value=${NODEGUID} \
		-f test/coder/traefik-values.yaml
	helm -n ${CHART_NS} ls
