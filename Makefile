# Copyright Contributors to the Open Cluster Management project

SHELL := /bin/bash

export ARCH       ?= $(shell uname -m)
export ARCH_TYPE   = $(if $(patsubst x86_64,,$(ARCH)),$(ARCH),amd64)

export CGO_ENABLED  = 0
export GO111MODULE := on
export GOOS         = $(shell go env GOOS)
export GOARCH       = $(ARCH_TYPE)
export GOPACKAGES   = $(shell go list ./... | grep -v /vendor | grep -v /internal | grep -v /build | grep -v /test)

export PROJECT_DIR            = $(shell 'pwd')
export PROJECT_NAME            = $(shell basename ${PROJECT_DIR})

export BUILD_DIR              = $(PROJECT_DIR)/build

export DOCKER_FILE        = $(BUILD_DIR)/Dockerfile.prow
export DOCKER_FILE_COVERAGE = $(BUILD_DIR)/Dockerfile.coverage.prow
export DOCKER_IMAGE      ?= clusterlifecycle-state-metrics
export DOCKER_IMAGE_COVERAGE_POSTFIX ?= -coverage
export DOCKER_IMAGE_COVERAGE      ?= $(DOCKER_IMAGE)$(DOCKER_IMAGE_COVERAGE_POSTFIX)
export DOCKER_BUILDER    ?= docker

export KUBECONFIG ?= ${HOME}/.kube/config

export KUBEBUILDER_HOME := /usr/local/kubebuilder

export PATH := ${PATH}:${KUBEBUILDER_HOME}/bin

BEFORE_SCRIPT := $(shell build/before-make.sh)

all: build-image

.PHONY: clean
clean:
	kind delete cluster --name ${PROJECT_NAME}-functional-test
	
.PHONY: dependencies
dependencies:
	@build/install-dependencies.sh

.PHONY: check
## Runs a set of required checks
check: dependencies check-copyright

.PHONY: check-copyright
check-copyright:
	@build/check-copyright.sh

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

.PHONY: build-coverage
build-coverage:
	$(SELF) component/build-coverage

.PHONY: copyright-check
copyright-check:
	./build/copyright-check.sh $(TRAVIS_BRANCH)

.PHONY: run
## Run the operator against the kubeconfig targeted cluster
run:
	go run cmd/clusterlifecycle-state-metrics/main.go --http-port=8082 --http-telemetry-port=8083 --csm-kubeconfig=${KUBECONFIG} -v=4; \

.PHONY: run-coverage
## Run the operator against the kubeconfig targeted cluster
run-coverage:
	#go test -v -covermode=atomic -coverpkg=github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/... -c -tags testrunmain ./cmd/clusterlifecycle-state-metrics -o clusterlifecycle-state-metrics-coverage
	go test -v -covermode=atomic -coverpkg=github.com/open-cluster-management/clusterlifecycle-state-metrics/pkg/... -tags testrunmain ./cmd/clusterlifecycle-state-metrics -args --http-port=8080 --http-telemetry-port=8081 --csm-kubeconfig=${KUBECONFIG}
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
	kubectl apply -f test/functional/resources/crds/certificates_crd.yaml
	kubectl apply -f test/functional/resources/crds/issuers_crd.yaml
	kubectl apply -f test/functional/resources/crds/managedclusterinfos_crd.yaml
	kubectl apply -f test/functional/resources/crds/managedclusters_crd.yaml
	kubectl apply -f test/functional/resources/crds/clusterversions_crd.yaml
	kubectl apply -f test/functional/resources/crds/servicemonitor_crd.yaml
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
functional-test-full: build-image-coverage
	@build/run-functional-tests.sh ${DOCKER_IMAGE_COVERAGE}:latest
