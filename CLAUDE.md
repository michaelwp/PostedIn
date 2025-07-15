# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based LinkedIn post scheduling application called "PostedIn". The application allows users to schedule LinkedIn posts with specific dates and times, manage their scheduled posts, and automatically publish them at exact scheduled times. Features include timezone-aware scheduling, automatic publishing with Go timers, multiple post deletion, and comprehensive LinkedIn API integration.

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
make dev                            # Format, vet, lint, and build
make fmt                            # Format Go code
make vet                            # Run go vet
make lint                           # Run golangci-lint
make test                           # Run tests
make tidy                           # Tidy modules
make clean                          # Clean build artifacts
make pre-commit                     # Run pre-commit checks manually
make start-daemon                   # Start scheduler daemon with auto-publishing

# Manual commands
go mod tidy                         # Clean up module dependencies
go fmt ./...                        # Format all Go files
go vet ./...                        # Run Go vet for static analysis
```

## Project Structure

```
PostedIn/
├── cmd/
│   ├── scheduler/          # Main application entry point
│   └── callback-server/    # OAuth callback server
├── internal/
│   ├── models/            # Data models (Post with CronEntryID)
│   ├── scheduler/         # Core scheduling logic
│   ├── cli/              # Command-line interface (11 menu options)
│   ├── cron/             # Automatic scheduling system (timer-based)
│   ├── config/           # Configuration and timezone management
│   ├── timezone/         # Timezone handling utilities
│   ├── auth/             # LinkedIn OAuth authentication
│   ├── debug/            # Authentication debugging utilities
│   └── api/              # API server for OAuth callbacks
├── pkg/
│   ├── storage/          # JSON storage implementation
│   └── linkedin/         # LinkedIn API client
├── go.mod
├── config.json          # Configuration file (auto-created)
├── posts.json           # Data storage (auto-created)
└── linkedin_token.json  # OAuth token storage (auto-created)
```

## Architecture Notes

The application follows Go best practices with modular design and clear separation of concerns:

- **Models** (`internal/models/`) - Data structures and business entities (Post with CronEntryID for timer tracking)
- **Scheduler** (`internal/scheduler/`) - Core business logic for post management (single + multiple deletion)
- **CLI** (`internal/cli/`) - User interface and interaction handling (11 menu options with timezone display)
- **Cron** (`internal/cron/`) - Automatic scheduling and timer management (timezone-aware, timer-based)
- **Config** (`internal/config/`) - Configuration and timezone management (supports IANA timezones)
- **Storage** (`pkg/storage/`) - Data persistence layer (JSON file storage)
- **LinkedIn** (`pkg/linkedin/`) - LinkedIn API client and OAuth authentication
- **CMD** (`cmd/scheduler/`) - Application entry point and dependency wiring

### Core Features

1. **Schedule Posts** - Add new posts with future dates/times in your configured timezone
2. **List Posts** - View all scheduled posts with status indicators and countdown timers
3. **Check Due Posts** - Review posts ready for publishing
4. **Delete Posts** - Remove single or multiple scheduled posts (supports 1,3,5 or 1 3 5 format)
5. **Automatic Publishing** - Timer-based automatic posting at exact scheduled times
6. **Timezone Management** - Configure and manage timezone settings (auto-detection + manual)
7. **LinkedIn Integration** - OAuth authentication and direct posting to LinkedIn
8. **Auto-Scheduler Status** - Real-time status display with active timers and next execution times
9. **Persistent Storage** - JSON-based data persistence with timer state tracking

### Key Technical Details

- **Timer-Based Scheduling**: Uses Go's `time.AfterFunc()` for precise one-time execution instead of cron expressions
- **Timezone Awareness**: All scheduling respects user's configured timezone using `cron.WithLocation()`
- **Multiple Post Deletion**: Supports both comma-separated and space-separated ID formats with confirmation
- **Self-Cleaning**: Automatically removes completed timers and cleans up orphaned resources
- **Status Tracking**: Real-time countdown timers and comprehensive status reporting

The modular architecture makes the codebase maintainable and allows for easy extension (e.g., adding different storage backends, notification systems, or social media platforms).