.PHONY: build test test-unit test-integration coverage clean fmt lint install

BINARY_NAME = aict
BUILD_DIR = bin
COVERAGE_FILE = coverage.out

# Build
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/aict

# Test targets
test: test-unit test-integration

test-unit:
	go test ./... -coverprofile=$(COVERAGE_FILE)

test-integration: build
	./test_functional.sh
	./test_since_option.sh

# Coverage
coverage: test-unit
	go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report: coverage.html"

# Code quality
fmt:
	go fmt ./...

lint:
	go vet ./...

# Install
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

# Clean
clean:
	rm -rf $(BUILD_DIR)/ $(COVERAGE_FILE) coverage.html
