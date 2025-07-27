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
	@echo "  help      - Show this help"