VERSION ?= none
REGISTRY_BASE ?= public.ecr.aws/f5j9d0q5/seaman

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
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


##@ Build

.PHONY: build-image
build-image: ## build Docker image
	docker build . \
		--build-arg APP_VERSION=$(VERSION) \
		--build-arg APP_COMMIT=$(shell git rev-parse --short HEAD) \
		-t $(REGISTRY_BASE)

.PHONY: push-image
push-image: ## push Docker image
	docker push $(REGISTRY_BASE)


##@ Development

.PHONY: generate
generate: ## Generate code
	go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint against code.
	$(GOLANGCI_LINT) run -c .golangci.yml

.PHONY: test
test: fmt vet ## Run some test against code.
	go test ./... -cover -v


##@ Install Tools

GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/cmd/golangci-lint
