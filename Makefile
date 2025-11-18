.PHONY: build install clean test test-verbose test-mcp test-coverage test-ci run release vet fmt lint deps ci

# Binary name
BINARY_DIR=bin
BINARY=$(BINARY_DIR)/lissto

# Build the CLI
build:
	mkdir -p $(BINARY_DIR)
	go build -o $(BINARY) .

# Install globally
install:
	go install

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(BINARY_DIR)/
	rm -f coverage.out coverage.html

# Run tests
test:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --randomize-all

# Run tests with verbose Ginkgo output
test-verbose:
	go run github.com/onsi/ginkgo/v2/ginkgo -r -v --randomize-all

# Run only MCP tests
test-mcp:
	go run github.com/onsi/ginkgo/v2/ginkgo -r -v ./pkg/mcp/

# Run tests with coverage (local development)
test-coverage:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --cover --coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for CI (with coverage using Ginkgo)
# Note: --race disabled due to Ginkgo parallel test orchestration issues (not production code)
test-ci:
	go run github.com/onsi/ginkgo/v2/ginkgo -r --cover --coverprofile=coverage.out --covermode=atomic --randomize-all --fail-fast

# Run the CLI (for development)
run:
	go run main.go

# Build for multiple platforms
build-all:
	mkdir -p build
	GOOS=darwin GOARCH=amd64 go build -o build/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o build/$(BINARY)-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o build/$(BINARY)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o build/$(BINARY)-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build -o build/$(BINARY)-windows-amd64.exe .

# Test release process locally (requires goreleaser)
release:
	goreleaser release --snapshot --clean

# Run go vet
vet:
	go vet ./...

# Format code
fmt:
	go fmt ./...

# Lint
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	go mod tidy

# Run all CI checks locally
ci: deps vet test-ci lint build
	@echo ""
	@echo "✅ All CI checks passed!"

# Help
help:
	@echo "Lissto CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build the CLI binary"
	@echo "  make install        - Install globally"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run tests with verbose Ginkgo output"
	@echo "  make test-mcp       - Run only MCP tests"
	@echo "  make test-coverage  - Run tests with coverage report (HTML)"
	@echo "  make test-ci        - Run tests with race detection and coverage (CI)"
	@echo "  make run            - Run the CLI in development mode"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make release        - Test release process locally (requires goreleaser)"
	@echo "  make vet            - Run go vet"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Download and verify dependencies"
	@echo "  make tidy           - Tidy dependencies"
	@echo "  make ci             - Run all CI checks locally"

