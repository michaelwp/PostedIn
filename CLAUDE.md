# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based LinkedIn post scheduling application called "PostedIn". The application allows users to schedule LinkedIn posts with specific dates and times, manage their scheduled posts, and track posting status through a CLI interface.

## Common Commands

### Building and Running
```bash
# Using Makefile (recommended)
make build                          # Build the binary
make run                            # Run the application directly
make run-bin                        # Build and run the binary
make help                           # Show all available targets

# Manual commands
go run cmd/scheduler/main.go        # Run the application directly
go build -o bin/linkedin-scheduler cmd/scheduler/main.go  # Build the binary
./bin/linkedin-scheduler            # Run the built binary
```

### Development
```bash
# Using Makefile (recommended)
make dev                            # Format, vet, and build
make fmt                            # Format Go code
make vet                            # Run go vet
make test                           # Run tests
make tidy                           # Tidy modules
make clean                          # Clean build artifacts

# Manual commands
go mod tidy                         # Clean up module dependencies
go fmt ./...                        # Format all Go files
go vet ./...                        # Run Go vet for static analysis
```

## Project Structure

```
PostedIn/
├── cmd/scheduler/          # Application entry point
├── internal/
│   ├── models/            # Data models (Post)
│   ├── scheduler/         # Core scheduling logic
│   └── cli/              # Command-line interface
├── pkg/storage/          # Storage implementations
├── go.mod
└── posts.json           # Data storage (created automatically)
```

## Architecture Notes

The application follows Go best practices with modular design and clear separation of concerns:

- **Models** (`internal/models/`) - Data structures and business entities
- **Scheduler** (`internal/scheduler/`) - Core business logic for post management
- **CLI** (`internal/cli/`) - User interface and interaction handling
- **Storage** (`pkg/storage/`) - Data persistence layer (JSON file storage)
- **CMD** (`cmd/scheduler/`) - Application entry point and dependency wiring

### Core Features

1. **Schedule Posts** - Add new posts with future dates/times
2. **List Posts** - View all scheduled posts with status indicators
3. **Check Due Posts** - Review posts ready for publishing
4. **Delete Posts** - Remove scheduled posts
5. **Persistent Storage** - JSON-based data persistence

The modular architecture makes the codebase maintainable and allows for easy extension (e.g., adding different storage backends or interfaces).