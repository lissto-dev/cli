.PHONY: build install clean test test-verbose test-mcp test-coverage run release vet fmt lint deps tidy ci check-goreleaser-version help

# Required GoReleaser version (pinned to avoid breaking changes)
GORELEASER_VERSION=2.13.0

# Binary name
BUILD_DIR=bin
BINARY=$(BUILD_DIR)/lissto

# Build variables (can be overridden)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

# Linker flags for version injection
LDFLAGS := -s -w \
	-X github.com/lissto-dev/cli/cmd.Version=$(VERSION) \
	-X github.com/lissto-dev/cli/cmd.Commit=$(COMMIT) \
	-X github.com/lissto-dev/cli/cmd.Date=$(BUILD_DATE)

# Docker variables
DOCKER_IMAGE ?= lissto/cli
PLATFORMS ?= linux/amd64,linux/arm64
CACHE_FROM ?=
CACHE_TO ?=

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test $$(go list ./...) -coverprofile cover.out

.PHONY: test-verbose
test-verbose: fmt vet ## Run tests with verbose output.
	go test -v $$(go list ./...)

.PHONY: test-mcp
test-mcp: fmt vet ## Run only MCP tests.
	go test -v ./pkg/mcp/...

.PHONY: test-coverage
test-coverage: fmt vet ## Run tests with coverage report (HTML).
	go test $$(go list ./...) -coverprofile cover.out
	go tool cover -html=cover.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint linter and perform fixes
	$(GOLANGCI_LINT) run --fix

##@ Build

.PHONY: build
build: fmt vet build-binary ## Build CLI binary (with fmt and vet).

.PHONY: build-binary
build-binary: ## Build CLI binary (without fmt/vet, used by Docker).
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

.PHONY: install
install: ## Install globally.
	go install -ldflags="$(LDFLAGS)"

.PHONY: run
run: fmt vet ## Run the CLI in development mode.
	go run main.go

.PHONY: build-all
build-all: fmt vet ## Build for multiple platforms.
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/lissto-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/lissto-darwin-arm64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/lissto-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/lissto-linux-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/lissto-windows-amd64.exe .

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker image for current platform.
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		-t $(DOCKER_IMAGE):$(VERSION) .

.PHONY: docker-push
docker-push: ## Build and push multi-arch Docker image.
	docker buildx build \
		--platform $(PLATFORMS) \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		$(if $(CACHE_FROM),--cache-from $(CACHE_FROM),) \
		$(if $(CACHE_TO),--cache-to $(CACHE_TO),) \
		-t $(DOCKER_IMAGE):$(VERSION) \
		--push .

##@ Release

.PHONY: release
release: check-goreleaser-version ## Test release process locally (requires goreleaser).
	goreleaser release --snapshot --clean

.PHONY: check-goreleaser-version
check-goreleaser-version: ## Check GoReleaser version.
	@echo "Checking GoReleaser version..."
	@INSTALLED_VERSION=$$(goreleaser --version 2>/dev/null | grep GitVersion | awk '{print $$2}'); \
	if [ -z "$$INSTALLED_VERSION" ]; then \
		echo "❌ GoReleaser not found. Install it with: brew install goreleaser"; \
		exit 1; \
	fi; \
	if [ "$$INSTALLED_VERSION" != "$(GORELEASER_VERSION)" ]; then \
		echo "⚠️  Warning: GoReleaser version mismatch"; \
		echo "   Expected: $(GORELEASER_VERSION)"; \
		echo "   Installed: $$INSTALLED_VERSION"; \
		echo "   Install correct version: brew install goreleaser@$(GORELEASER_VERSION) || brew upgrade goreleaser"; \
		echo "   Continuing anyway..."; \
	else \
		echo "✅ GoReleaser version $(GORELEASER_VERSION) is correct"; \
	fi

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts.
	go clean
	@rm -rf $(BUILD_DIR)/
	@rm -f cover.out coverage.html

##@ Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint

## Tool Versions
GOLANGCI_LINT_VERSION ?= v2.4.0

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(realpath $(1)-$(3)) $(1)
endef

.PHONY: deps
deps: ## Download and verify dependencies.
	go mod download
	go mod verify

.PHONY: tidy
tidy: ## Tidy dependencies.
	go mod tidy

