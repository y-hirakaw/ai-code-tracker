.PHONY: build test install clean fmt vet

# Variables
BINARY_NAME=aict
BINARY_PATH=./cmd/aict
BUILD_DIR=./bin
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(BINARY_PATH)

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run integration tests (including E2E)
test-integration:
	@echo "Running integration tests..."
	go test -v ./test/integration/...

# Run performance benchmarks
bench-performance:
	@echo "Running performance benchmarks..."
	go run cmd/aict-bench/main.go --all --events=1000

# Run security scan
security-scan:
	@echo "Running security scan..."
	bash scripts/security-scan.sh

# Run all tests and checks
test-all: test test-integration bench-performance security-scan

# Install the binary to $GOPATH/bin
install: build
	@echo "Installing $(BINARY_NAME) to $$GOPATH/bin..."
	go install $(BINARY_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all checks
check: fmt vet test

# Development build with race detection
dev:
	@echo "Building development version with race detection..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(BINARY_PATH)

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  install   - Install binary to $$GOPATH/bin"
	@echo "  clean     - Clean build artifacts"
	@echo "  fmt       - Format code"
	@echo "  vet       - Vet code"
	@echo "  check     - Run fmt, vet, and test"
	@echo "  dev       - Build development version with race detection"
	@echo "  test-integration - Run integration tests"
	@echo "  bench-performance - Run performance benchmarks"
	@echo "  security-scan - Run security scan"
	@echo "  test-all  - Run all tests and checks"
	@echo "  help      - Show this help"