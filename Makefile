.PHONY: gosnappass
gosnappass: build-dir
	go build -ldflags="-s -w" -o ./build/gosnappass ./cmd/gosnappass/

build: gosnappass

.PHONY: clean
clean:
	rm -rf build/
	rm -rf bin/

.PHONY: run
run:
	go run ./cmd/gosnappass

.PHONY: build-dir
build-dir:
	test -d build || mkdir build

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter checks.
	$(GOLANGCI_LINT) run

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.50.0
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT):
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

GOFUMPT = $(shell pwd)/bin/gofumpt
gofumpt: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(GOFUMPT),mvdan.cc/gofumpt@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef

# Helpers for starting Redis locally
CONTAINER_ENGINE ?= podman
# The host port to listen on for the redis dev server
# Useful for tesing changes to REDIS_PORT
REDIS_PORT ?= 6379

.PHONY: redis
redis:
	$(CONTAINER_ENGINE) run --rm -d -p $(REDIS_PORT):6379 --name gosnappass-redis-server docker.io/redis/redis-stack

.PHONY: redis-teardown
redis-teardown:
	$(CONTAINER_ENGINE) stop gosnappass-redis-server

IMAGE_TAG ?= gosnappass:devbuild0

.PHONY: image
image:
	$(CONTAINER_ENGINE) build -f Dockerfile -t $(IMAGE_TAG)


PODNAME ?= gosnappass-dev-pod
.PHONY: pod
pod: image
	test $(CONTAINER_ENGINE) == podman
	$(CONTAINER_ENGINE) pod create -p 5000:5000 $(PODNAME)
	$(CONTAINER_ENGINE) run -d --pod $(PODNAME) --name gosnappass-redis-server-in-pod docker.io/redis/redis-stack
	$(CONTAINER_ENGINE) run -d --pod $(PODNAME) --name gosnappass-in-pod $(IMAGE_TAG)

.PHONY: pod-teardown
pod-teardown:
	test $(CONTAINER_ENGINE) == podman
	$(CONTAINER_ENGINE) pod stop $(PODNAME)
	sleep 5
	$(CONTAINER_ENGINE) pod rm $(PODNAME)
