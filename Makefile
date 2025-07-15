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
CALLBACK_BINARY_NAME=callback-server
BINARY_PATH=bin/$(BINARY_NAME)
CALLBACK_BINARY_PATH=bin/$(CALLBACK_BINARY_NAME)
MAIN_PATH=cmd/scheduler/main.go
CALLBACK_MAIN_PATH=cmd/callback-server/main.go

# Default target
.DEFAULT_GOAL := help

# Build the main application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Binary created at $(BINARY_PATH)"

# Build the callback server
build-callback:
	@echo "Building $(CALLBACK_BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) -o $(CALLBACK_BINARY_PATH) $(CALLBACK_MAIN_PATH)
	@echo "Binary created at $(CALLBACK_BINARY_PATH)"

# Build all binaries
build-all: build build-callback
	@echo "All binaries built successfully"

# Run the main application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# Run the callback server
run-callback:
	@echo "Running $(CALLBACK_BINARY_NAME)..."
	$(GOCMD) run $(CALLBACK_MAIN_PATH)

# Run the built main binary
run-bin: build
	@echo "Running built binary..."
	./$(BINARY_PATH)

# Run the built callback server binary
run-callback-bin: build-callback
	@echo "Running built callback server..."
	./$(CALLBACK_BINARY_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -f $(BINARY_PATH) $(CALLBACK_BINARY_PATH)
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
	@echo "  build-callback    - Build the callback server"
	@echo "  build-all         - Build both applications"
	@echo "  run               - Run the main application directly"
	@echo "  run-callback      - Run the callback server directly"
	@echo "  run-bin           - Build and run the main binary"
	@echo "  run-callback-bin  - Build and run the callback server binary"
	@echo "  clean             - Clean build artifacts"
	@echo "  fmt               - Format Go code"
	@echo "  vet               - Run go vet"
	@echo "  lint              - Run golangci-lint"
	@echo "  test              - Run tests"
	@echo "  tidy              - Tidy Go modules"
	@echo "  dev               - Format, vet, lint, and build (development workflow)"
	@echo "  deps              - Install dependencies"
	@echo "  pre-commit        - Run pre-commit checks manually"
	@echo "  start-daemon      - Start scheduler daemon with auto-publishing"
	@echo "  help              - Show this help message"

.PHONY: build build-callback build-all run run-callback run-bin run-callback-bin clean fmt vet lint test tidy dev deps pre-commit start-daemon help