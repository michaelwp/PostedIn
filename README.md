# PostedIn - LinkedIn Post Scheduler

A command-line application for scheduling LinkedIn posts with Go, featuring automatic publishing and timezone-aware scheduling.

## Features

- **Smart Scheduling** - Schedule posts with specific dates and times in your timezone
- **Automatic Publishing** - Timer-based automatic posting at exact scheduled times
- **Multiple Post Management** - Delete single or multiple posts at once
- **Timezone Support** - Configure your local timezone for accurate scheduling
- **LinkedIn API Integration** - Automatically publish to LinkedIn
- **OAuth2 Authentication** - Secure LinkedIn login
- **Auto-publish** - Bulk publish all due posts
- **Real-time Status** - Live status display with countdown timers
- **Persistent JSON storage** - Reliable data storage
- **Clean modular architecture** - Well-organized codebase

## Project Structure

```
PostedIn/
├── bin/                    # Binary files (gitignored)
├── cmd/
│   ├── scheduler/          # Main application entry point
│   │   └── main.go
│   └── callback-server/    # OAuth callback server
│       └── main.go
├── internal/
│   ├── models/            # Data models (Post)
│   │   └── post.go
│   ├── scheduler/         # Core scheduling logic
│   │   └── scheduler.go
│   ├── cli/              # Command-line interface
│   │   └── cli.go
│   ├── cron/             # Automatic scheduling system
│   │   └── cron.go
│   ├── config/           # Configuration management
│   │   └── config.go
│   ├── timezone/         # Timezone handling
│   │   └── timezone.go
│   ├── auth/             # LinkedIn OAuth authentication
│   │   └── auth.go
│   ├── debug/            # Debugging utilities
│   │   └── auth.go
│   └── api/              # API server
│       └── server.go
├── pkg/
│   ├── storage/          # Storage implementations
│   │   └── json.go
│   └── linkedin/         # LinkedIn API client
│       └── client.go
├── go.mod
├── config.json          # Configuration file (created automatically)
├── posts.json           # Data storage (created automatically)
└── linkedin_token.json  # OAuth token storage (created automatically)
```

## Building and Running

### Using Makefile (Recommended):
```bash
make help              # Show all available targets
make build             # Build the binary
make run               # Run the application directly
make run-bin           # Build and run the binary
make start-daemon      # Start scheduler daemon with auto-publishing
make clean             # Clean build artifacts
```

### Manual Commands:
```bash
# Run directly
go run cmd/scheduler/main.go

# Build binary
go build -o bin/linkedin-scheduler cmd/scheduler/main.go
./bin/linkedin-scheduler
```

### Development:
```bash
make dev       # Format, vet, lint, and build
make fmt       # Format code
make vet       # Run go vet
make lint      # Run golangci-lint
make test      # Run tests
make tidy      # Tidy modules
make pre-commit # Run pre-commit checks manually
```

## LinkedIn API Setup

To use LinkedIn posting features, you need to set up a LinkedIn app:

1. See [LINKEDIN_SETUP.md](LINKEDIN_SETUP.md) for detailed setup instructions
2. Run the app and it will create a `config.json` template
3. Fill in your LinkedIn app credentials
4. Use option 5 to authenticate with LinkedIn
5. Start posting!

## Quick Start

1. **Build and run**: `make run` or `go run cmd/scheduler/main.go`
2. **Configure timezone**: Choose option 9 to set your local timezone
3. **Setup LinkedIn**: See [LINKEDIN_SETUP.md](LINKEDIN_SETUP.md) and use option 5 to authenticate
4. **Schedule posts**: Use option 1 to schedule posts
5. **Watch them publish automatically**: The app will publish at exact scheduled times

## Usage

The application provides an interactive menu with the following options:

1. **Schedule a new post** - Enter content and target date/time in your timezone
2. **List scheduled posts** - View all posts with their status and countdown timers
3. **Check due posts** - Review posts ready for publishing
4. **Delete posts** - Remove single or multiple posts (supports: `5` or `1,3,5` or `1 3 5`)
5. **Authenticate with LinkedIn** - Set up LinkedIn API access (one-time setup)
6. **Publish specific post to LinkedIn** - Manually publish a post
7. **Auto-publish all due posts** - Bulk publish ready posts
8. **Debug LinkedIn authentication** - Troubleshoot authentication issues
9. **Configure timezone** - Set your local timezone (shows current timezone in menu)
10. **Check auto-scheduler status** - View detailed status of automatic scheduling
11. **Exit** - Close the application

## Automatic Scheduling

PostedIn features a sophisticated automatic scheduling system:

- **Timer-Based**: Uses precise Go timers instead of periodic checking
- **Timezone-Aware**: Respects your configured timezone settings
- **Real-Time Status**: Shows countdown timers and next scheduled publication
- **Auto-Start**: Automatically starts when you schedule your first post
- **Self-Cleaning**: Removes completed timers automatically

### Auto-Scheduler Features

- **Exact Timing**: Posts publish at precisely their scheduled time
- **Status Display**: 
  ```
  📅 Auto-scheduler: ACTIVE (next run: 11:35:00 WIB)
  ```
- **Detailed Status**: View active timers, pending posts, and next execution times
- **Background Operation**: Runs silently in the background

## Multiple Post Deletion

Delete single or multiple posts efficiently:

```
Delete Posts
============
Enter one or more post IDs to delete:
- Single post: 5
- Multiple posts: 1,3,5 or 1 3 5

Enter post ID(s): 1,3,5
You are about to delete 3 posts with IDs: [1 3 5]
Are you sure? (y/N): y
✅ Successfully deleted 3 post(s).
```

## Timezone Configuration

Configure your local timezone for accurate scheduling:

- **Auto-Detection**: Detects your system timezone automatically
- **Common Timezones**: Choose from predefined options
- **Custom Timezones**: Enter any IANA timezone identifier
- **Dynamic Updates**: Changes take effect immediately

## Architecture

The application follows Go best practices with clear separation of concerns and modular design:

- **Models** (`internal/models/`) - Data structures and business entities
- **Scheduler** (`internal/scheduler/`) - Core business logic for post management
- **CLI** (`internal/cli/`) - User interface and interaction handling
- **Cron** (`internal/cron/`) - Automatic scheduling and timer management
- **Config** (`internal/config/`) - Configuration and timezone management
- **Storage** (`pkg/storage/`) - Data persistence layer (JSON file storage)
- **LinkedIn** (`pkg/linkedin/`) - LinkedIn API client and authentication
- **CMD** (`cmd/scheduler/`) - Application entry point and dependency wiring

### Core Features

1. **Schedule Posts** - Add new posts with future dates/times in your timezone
2. **List Posts** - View all scheduled posts with status indicators and countdown timers
3. **Check Due Posts** - Review posts ready for publishing
4. **Delete Posts** - Remove single or multiple scheduled posts
5. **Automatic Publishing** - Timer-based automatic posting at exact scheduled times
6. **Timezone Management** - Configure and manage timezone settings
7. **LinkedIn Integration** - OAuth authentication and direct posting
8. **Persistent Storage** - JSON-based data persistence

The modular architecture makes the codebase maintainable and allows for easy extension (e.g., adding different storage backends, notification systems, or social media platforms).

## Development

### Code Quality

- **Linting**: Uses golangci-lint for code quality checks
- **Pre-commit Hooks**: Automatic linting on git commits
- **Error Handling**: Comprehensive error handling throughout
- **Logging**: Detailed logging for debugging and monitoring
- **Testing**: Structured for easy unit testing

### Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `make dev` to ensure code quality
5. Submit a pull request

## Troubleshooting

### Common Issues

1. **Timezone Issues**: Use option 9 to configure your local timezone
2. **LinkedIn Authentication**: Use option 8 to debug authentication issues
3. **Posts Not Publishing**: Check option 10 for auto-scheduler status
4. **Build Issues**: Run `make clean && make build`

### Debug Mode

Enable verbose logging by checking the auto-scheduler status (option 10) which shows:
- Current timezone configuration
- Active timers and their next execution times
- Pending posts with countdown timers
- System status and health checks