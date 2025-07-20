# PostedIn - LinkedIn Post Scheduler

A full-stack application for scheduling LinkedIn posts with Go backend and React web client, featuring automatic publishing and timezone-aware scheduling.

## Features

### Core Features
- **Smart Scheduling** - Schedule posts with specific dates and times in your timezone
- **Automatic Publishing** - Timer-based automatic posting at exact scheduled times
- **Multiple Post Management** - Delete single or multiple posts at once
- **Timezone Support** - Configure your local timezone for accurate scheduling
- **LinkedIn API Integration** - Automatically publish to LinkedIn
- **OAuth2 Authentication** - Secure LinkedIn login
- **Auto-publish** - Bulk publish all due posts
- **Real-time Status** - Live status display with countdown timers
- **Persistent JSON storage** - Reliable data storage

### Web Interface
- **React Web Client** - Modern web interface built with React and TypeScript
- **RESTful API** - Comprehensive REST API with OpenAPI/Swagger documentation
- **Real-time Dashboard** - Web-based dashboard for managing posts
- **Responsive Design** - Works on desktop and mobile devices

### Architecture
- **Clean modular architecture** - Well-organized codebase
- **API Versioning** - Versioned REST API (`/api/v1/`) for backward compatibility
- **Static File Serving** - Integrated web client serving

## Project Structure

```
PostedIn/
├── bin/                    # Binary files and web client (gitignored)
│   └── web/
│       └── dist/          # Built web client files
├── cmd/
│   ├── scheduler/          # CLI application entry point
│   │   └── main.go
│   └── web-api/           # Web API server entry point
│       └── main.go
├── internal/
│   ├── models/            # Data models (Post with CronEntryID)
│   │   └── post.go
│   ├── scheduler/         # Core scheduling logic
│   │   └── scheduler.go
│   ├── cli/              # Command-line interface (11 menu options)
│   │   └── cli.go
│   ├── cron/             # Automatic scheduling system (timer-based)
│   │   └── cron.go
│   ├── config/           # Configuration and timezone management
│   │   └── config.go
│   ├── timezone/         # Timezone handling utilities
│   │   └── timezone.go
│   ├── auth/             # LinkedIn OAuth authentication
│   │   └── auth.go
│   ├── debug/            # Authentication debugging utilities
│   │   └── auth.go
│   └── api/              # RESTful API server (versioned)
│       ├── router.go     # API routing and middleware
│       ├── posts.go      # Post management endpoints
│       ├── auth.go       # Authentication endpoints
│       ├── scheduler.go  # Scheduler control endpoints
│       └── timezone.go   # Timezone configuration endpoints
├── web-client/            # React/TypeScript web client
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── services/     # API client
│   │   ├── store/        # State management (Zustand)
│   │   └── models/       # TypeScript types
│   ├── dist/            # Built web client (copied to bin/web/dist)
│   ├── package.json
│   └── vite.config.ts
├── pkg/
│   ├── storage/          # JSON storage implementation
│   │   └── json.go
│   └── linkedin/         # LinkedIn API client
│       └── client.go
├── docs/                 # Swagger/OpenAPI documentation
├── go.mod
├── config.json          # Configuration file (auto-created)
├── posts.json           # Data storage (auto-created)
└── linkedin_token.json  # OAuth token storage (auto-created)
```

## Building and Running

### Using Makefile (Recommended):
```bash
make help              # Show all available targets
make build             # Build the CLI application
make build-web-api     # Build the web API server
make build-web-client  # Build the React web client
make build-all         # Build all applications and web client
make run               # Run the CLI application directly
make run-web-api       # Run the web API server directly
make start-daemon      # Start scheduler daemon with auto-publishing
make clean             # Clean all build artifacts
```

### Manual Commands:
```bash
# Run CLI application directly
go run cmd/scheduler/main.go

# Run web API server directly
go run cmd/web-api/main.go

# Build CLI binary
go build -o bin/linkedin-scheduler cmd/scheduler/main.go
./bin/linkedin-scheduler

# Build web API server binary
go build -o bin/web-api-server cmd/web-api/main.go
./bin/web-api-server

# Build web client
cd web-client && npm ci && npm run build
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

### CLI Application
1. **Build and run**: `make run` or `go run cmd/scheduler/main.go`
2. **Configure timezone**: Choose option 9 to set your local timezone
3. **Setup LinkedIn**: See [LINKEDIN_SETUP.md](LINKEDIN_SETUP.md) and use option 5 to authenticate
4. **Schedule posts**: Use option 1 to schedule posts
5. **Watch them publish automatically**: The app will publish at exact scheduled times

### Web Application
1. **Build everything**: `make build-all`
2. **Start web server**: `make run-web-api` or `go run cmd/web-api/main.go`
3. **Open in browser**: Navigate to [http://localhost:8080](http://localhost:8080)
4. **Configure and authenticate**: Use the web interface to set up LinkedIn integration
5. **Manage posts**: Schedule, edit, and manage posts through the web dashboard

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

## Web Interface

The application includes a modern React-based web interface that provides all the functionality of the CLI in a user-friendly web dashboard.

### Features
- **Dashboard**: Overview of all scheduled posts with real-time status
- **Post Management**: Create, edit, delete, and publish posts
- **Scheduler Control**: Start/stop automatic scheduling
- **Authentication**: LinkedIn OAuth integration
- **Timezone Configuration**: Set and manage your timezone
- **Real-time Updates**: Live countdown timers and status updates

### Accessing the Web Interface

1. **Start the web API server**:
   ```bash
   make run-web-api
   # or
   go run cmd/web-api/main.go
   ```

2. **Open in browser**: [http://localhost:8080](http://localhost:8080)

3. **Alternative: Build and serve separately**:
   ```bash
   make build-all                    # Build everything
   ./bin/web-api-server             # Start server
   ```

## API Documentation (Swagger/OpenAPI)

The web API is fully documented using Swagger (OpenAPI 3.0). All endpoints are versioned under `/api/v1/` for backward compatibility.

### API Endpoints
- **Base URL**: `http://localhost:8080/api/v1/`
- **Posts**: `/api/v1/posts` - Manage scheduled posts
- **Authentication**: `/api/v1/auth` - LinkedIn OAuth
- **Scheduler**: `/api/v1/scheduler` - Control automatic publishing
- **Timezone**: `/api/v1/timezone` - Configure timezone settings
- **Callback**: `/api/v1/callback` - LinkedIn OAuth callback

### How to Generate Docs

- To generate or update the Swagger docs after editing API code or comments:
  ```bash
  make swagger
  ```
- To remove generated docs:
  ```bash
  make docs-clean
  ```

### How to View the API Docs

1. Start the web API server:
   ```bash
   make run-web-api
   ```
2. Open your browser and go to:
   [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

This interactive UI allows you to try out all endpoints, see request/response schemas, and explore the API.

### API Versioning

The API uses semantic versioning with URL-based versioning:
- **Current version**: `v1`
- **Full endpoint format**: `http://localhost:8080/api/v1/{endpoint}`
- **Backward compatibility**: Maintained across minor version updates