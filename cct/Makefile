# Makefile for CCT (Crawler Control Tool) application

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
GOLINT=golangci-lint

# Binary name
BINARY_NAME=cct
BINARY_UNIX=$(BINARY_NAME)_unix

# Main package path
MAIN_PACKAGE=.

# Build directory
BUILD_DIR=./bin

# Source files
SRC=$(shell find . -name "*.go" -type f)

# Default target
.PHONY: all
all: lint test build

# Build the application
.PHONY: build
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
.PHONY: run
run:
	@echo "Running..."
	$(GORUN) $(MAIN_PACKAGE)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@which $(GOLINT) > /dev/null || (echo "Installing golangci-lint..." && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin)
	$(GOLINT) run ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Update dependencies
.PHONY: deps
deps:
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

# Build for multiple platforms
.PHONY: build-all
build-all: build-linux build-windows build-macos

# Build for Linux
.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_linux $(MAIN_PACKAGE)

# Build for Windows
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PACKAGE)

# Build for macOS
.PHONY: build-macos
build-macos:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_macos $(MAIN_PACKAGE)

# Create database schema
.PHONY: db-schema
db-schema:
	@echo "Creating database schema..."
	@if [ -z "$(DB_NAME)" ]; then \
		echo "Error: DB_NAME environment variable is not set"; \
		exit 1; \
	fi
	psql -d $(DB_NAME) -f crawler_schema.sql

# Help target
.PHONY: help
help:
	@echo "CCT (Crawler Control Tool) Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make              Build the application after running tests and linting"
	@echo "  make build        Build the application"
	@echo "  make run          Run the application"
	@echo "  make clean        Clean build artifacts"
	@echo "  make test         Run tests"
	@echo "  make test-coverage Run tests with coverage report"
	@echo "  make lint         Run linter"
	@echo "  make fmt          Format code"
	@echo "  make deps         Update dependencies"
	@echo "  make build-all    Build for multiple platforms"
	@echo "  make db-schema    Create database schema (requires DB_NAME env var)"
	@echo "  make help         Show this help message"
