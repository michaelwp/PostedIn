# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod

# Binary info
BINARY_NAME=linkedin-scheduler
WEB_API_BINARY_NAME=web-api-server
BINARY_PATH=bin/$(BINARY_NAME)
WEB_API_BINARY_PATH=bin/$(WEB_API_BINARY_NAME)
MAIN_PATH=cmd/scheduler/main.go
WEB_API_MAIN_PATH=cmd/web-api/main.go

# Default target
.DEFAULT_GOAL := help

# Build the main application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Binary created at $(BINARY_PATH)"

# Build the web API server
build-web-api:
	@echo "Building $(WEB_API_BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(WEB_API_BINARY_PATH) $(WEB_API_MAIN_PATH)
	@echo "Binary created at $(WEB_API_BINARY_PATH)"

# Build all binaries
build-all: build build-web-api
	@echo "All binaries built successfully"

# Run the main application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Run the web API server
run-web-api:
	@echo "Running $(WEB_API_BINARY_NAME)..."
	$(GOCMD) run $(WEB_API_MAIN_PATH)

# Run the built main binary
run-bin: build
	@echo "Running built binary..."
	./$(BINARY_PATH)

# Run the built web API server binary
run-web-api-bin: build-web-api
	@echo "Running built web API server..."
	./$(WEB_API_BINARY_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_PATH) $(WEB_API_BINARY_PATH)
	@echo "Clean completed"

# Format Go code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Run golangci-lint
lint:
	@echo "Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run

# Run golangci-lint with auto-fix
lint-fix:
	@echo "Running golangci-lint with auto-fix..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not found. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --fix

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Tidy modules
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy

# Development workflow
dev: fmt vet lint build
	@echo "Development build completed"

# Development workflow with auto-fix
dev-fix: fmt vet lint-fix build
	@echo "Development build with auto-fix completed"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download

# Run pre-commit checks manually
pre-commit:
	@echo "Running pre-commit checks..."
	@.git/hooks/pre-commit

# Start scheduler daemon (automatically manages scheduled posts)
start-daemon: build
	@echo "Starting LinkedIn scheduler daemon..."
	@echo "Auto-scheduling is enabled - posts will be published automatically"
	@echo "Use Ctrl+C to stop the daemon"
	./$(BINARY_PATH)

# Show help
help:
	@echo "Available targets:"
	@echo "  build             - Build the main application"
	@echo "  build-web-api     - Build the web API server (includes OAuth callback)"
	@echo "  build-all         - Build all applications"
	@echo "  run               - Run the main application directly"
	@echo "  run-web-api       - Run the web API server directly"
	@echo "  run-bin           - Build and run the main binary"
	@echo "  run-web-api-bin   - Build and run the web API server binary"
	@echo "  clean             - Clean build artifacts"
	@echo "  fmt               - Format Go code"
	@echo "  vet               - Run go vet"
	@echo "  lint              - Run golangci-lint"
	@echo "  lint-fix          - Run golangci-lint with auto-fix"
	@echo "  test              - Run tests"
	@echo "  tidy              - Tidy Go modules"
	@echo "  dev               - Format, vet, lint, and build (development workflow)"
	@echo "  dev-fix           - Format, vet, lint with auto-fix, and build"
	@echo "  deps              - Install dependencies"
	@echo "  pre-commit        - Run pre-commit checks manually"
	@echo "  start-daemon      - Start scheduler daemon with auto-publishing"
	@echo "  help              - Show this help message"

.PHONY: build build-web-api build-all run run-web-api run-bin run-web-api-bin clean fmt vet lint lint-fix test tidy dev dev-fix deps pre-commit start-daemon help