.PHONY: build install clean test test-verbose test-mcp test-coverage run release

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

# Run tests
test:
	go test -v ./...

# Run tests with verbose Ginkgo output
test-verbose:
	go test -v -ginkgo.v ./...

# Run only MCP tests
test-mcp:
	go test -v ./pkg/mcp/...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

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

# Format code
fmt:
	go fmt ./...

# Lint
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

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
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make run            - Run the CLI in development mode"
	@echo "  make build-all      - Build for multiple platforms"
	@echo "  make release        - Test release process locally (requires goreleaser)"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make deps           - Download and tidy dependencies"

