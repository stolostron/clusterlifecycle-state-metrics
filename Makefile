
SHELL := /bin/bash

export BINDATA_TEMP_DIR := $(shell mktemp -d)

export GIT_COMMIT      = $(shell git rev-parse --short HEAD)
export GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
export GITHUB_USER    := $(shell echo $(GITHUB_USER) | sed 's/@/%40/g')
export GITHUB_TOKEN   ?=

export ARCH       ?= $(shell uname -m)
export ARCH_TYPE   = $(if $(patsubst x86_64,,$(ARCH)),$(ARCH),amd64)
export BUILD_DATE  = $(shell date +%m/%d@%H:%M:%S)
export VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))

export CGO_ENABLED  = 0
export GO111MODULE := on
export GOOS         = $(shell go env GOOS)
export GOARCH       = $(ARCH_TYPE)
export GOPACKAGES   = $(shell go list ./... | grep -v /vendor | grep -v /internal | grep -v /build | grep -v /test)

export PROJECT_DIR            = $(shell 'pwd')
export PROJECT_NAME            = $(shell basename ${PROJECT_DIR})

export BUILD_DIR              = $(PROJECT_DIR)/build
export COMPONENT_SCRIPTS_PATH = $(BUILD_DIR)
export KLUSTERLET_CRD_FILE      = $(PROJECT_DIR)/build/resources/agent.open-cluster-management.io_v1beta1_klusterlet_crd.yaml

export COMPONENT_NAME ?= $(shell cat ./COMPONENT_NAME 2> /dev/null)
export COMPONENT_VERSION ?= $(shell cat ./COMPONENT_VERSION 2> /dev/null)
export SECURITYSCANS_IMAGE_NAME ?= $(shell cat ./COMPONENT_NAME 2> /dev/null)
export SECURITYSCANS_IMAGE_VERSION ?= $(shell cat ./COMPONENT_VERSION 2> /dev/null)

export RELEASE_MAIN_BRANCH ?= main

## WARNING: OPERATOR-SDK - IMAGE_DESCRIPTION & DOCKER_BUILD_OPTS MUST NOT CONTAIN ANY SPACES
export IMAGE_DESCRIPTION ?= RCM_Controller
export DOCKER_FILE        = $(BUILD_DIR)/Dockerfile.prow
export DOCKER_FILE_COVERAGE = $(BUILD_DIR)/Dockerfile.coverage.prow
export DOCKER_REGISTRY   ?= quay.io
export DOCKER_NAMESPACE  ?= stolostron
export DOCKER_IMAGE      ?= $(COMPONENT_NAME)
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_BUILD_TAG  ?= latest
export DOCKER_TAG        ?= $(shell whoami)
export DOCKER_BUILDER    ?= docker

export KUBECONFIG ?= ${HOME}/.kube/config

BEFORE_SCRIPT := $(shell build/before-make.sh)

export DOCKER_BUILD_OPTS  = --build-arg VCS_REF=$(VCS_REF) \
	--build-arg VCS_URL=$(GIT_REMOTE_URL) \
	--build-arg IMAGE_NAME=$(DOCKER_IMAGE) \
	--build-arg IMAGE_DESCRIPTION=$(IMAGE_DESCRIPTION) \
	--build-arg ARCH_TYPE=$(ARCH_TYPE) \
	--build-arg REMOTE_SOURCE=. \
	--build-arg REMOTE_SOURCE_DIR=/remote-source \
	--build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)

# Only use git commands if it exists
ifdef GIT
GIT_COMMIT      = $(shell git rev-parse --short HEAD)
GIT_REMOTE_URL  = $(shell git config --get remote.origin.url)
VCS_REF     = $(if $(shell git status --porcelain),$(GIT_COMMIT)-$(BUILD_DATE),$(GIT_COMMIT))
endif

.PHONY: dependencies
dependencies:
	@build/install-dependencies.sh

.PHONY: check
## Runs a set of required checks
check: copyright-check

.PHONY: test
## Runs go unit tests
test: dependencies
	@build/run-unit-tests.sh

.PHONY: build-image
## Builds controller binary inside of an image
build-image: 
	$(DOCKER_BUILDER) build -f $(DOCKER_FILE) . -t $(DOCKER_IMAGE)

.PHONY: build-image-coverage
## Builds controller binary inside of an image
build-image-coverage: build-image
	$(DOCKER_BUILDER) build -f $(DOCKER_FILE_COVERAGE) . -t $(DOCKER_IMAGE_COVERAGE)

.PHONY: copyright-check
copyright-check:
	./build/copyright-check.sh $(TRAVIS_BRANCH)

.PHONY: clean
## Clean build-harness and remove Go generated build and test files
clean::
	@rm -rf $(BUILD_DIR)/_output
	kind delete cluster --name ${PROJECT_NAME}-functional-test

.PHONY: run
## Run the operator against the kubeconfig targeted cluster
run:
	go run cmd/clusterlifecycle-state-metrics/main.go --http-port=8080 --http-telemetry-port=8081 --csm-kubeconfig=${KUBECONFIG} -v=4; \

.PHONY: run-coverage
## Run the operator against the kubeconfig targeted cluster
run-coverage:
	#go test -v -covermode=atomic -coverpkg=github.com/stolostron/clusterlifecycle-state-metrics/pkg/... -c -tags testrunmain ./cmd/clusterlifecycle-state-metrics -o clusterlifecycle-state-metrics-coverage
	go test -v -covermode=atomic -coverpkg=github.com/stolostron/clusterlifecycle-state-metrics/pkg/... -tags testrunmain ./cmd/clusterlifecycle-state-metrics -args --http-port=8080 --http-telemetry-port=8081 --kubeconfig=${KUBECONFIG}
	# -args -port 8080 -telemetry-port 8081 -kubeconfig ${KUBECONFIG}
.PHONY: lint
## Runs linter against go files
lint:
	@echo "Running linting tool ..."
	@GOGC=25 golangci-lint run --timeout 5m

############################################################
# deploy section
############################################################

.PHONY: deploy
deploy:
	cd overlays/deploy
	kustomize build overlays/deploy | kubectl apply -f -

.PHONY: undeploy
undeploy:
	cd overlays/deploy
	kubectl delete --wait=true -k overlays/deploy

############################################################
# functional test section
############################################################

.PHONY: install-fake-crds
install-fake-crds:
	@echo installing crds
	kubectl apply -f test/functional/resources/certificates_crd.yaml
	kubectl apply -f test/functional/resources/issuers_crd.yaml
	kubectl apply -f test/functional/resources/managedclusterinfos_crd.yaml
	kubectl apply -f test/functional/resources/hive_v1_clusterdeployment_crd.yaml
	kubectl apply -f test/functional/resources/clusterversions_crd.yaml
	kubectl apply -f test/functional/resources/servicemonitor_crd.yaml
	@sleep 10 

.PHONY: kind-cluster-setup
kind-cluster-setup: install-fake-crds
	@echo installing fake resources
	kubectl apply -f test/functional/resources/namespace_osm.yaml
	kubectl apply -f test/functional/resources/namespace_ocm.yaml
	kubectl apply -f test/functional/resources/clusterversions_cr.yaml
	@echo "Install ingress NGNIX"
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml
	@echo "Wait ingress NGNIX ready"
	kubectl wait --namespace ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=180s
	kubectl delete ValidatingWebhookCOnfiguration ingress-nginx-admission

.PHONY: functional-test
functional-test:
	@echo executing test
	ginkgo -tags functional -v --slowSpecThreshold=30 test/functional -- -v=5

.PHONY: functional-test-full
# functional-test-full: 
functional-test-full: component/build-coverage
	$(SELF) component/test/functional