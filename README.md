# PostedIn - LinkedIn Post Scheduler

A command-line application for scheduling LinkedIn posts with Go.

## Features

- Schedule posts with specific dates and times
- List all scheduled posts with status indicators
- Check for posts ready to publish
- Delete scheduled posts
- **LinkedIn API Integration** - Automatically publish to LinkedIn
- **OAuth2 Authentication** - Secure LinkedIn login
- **Auto-publish** - Bulk publish all due posts
- Persistent JSON storage
- Clean modular architecture

## Project Structure

```
PostedIn/
├── bin/                  # Binary files (gitignored)
├── cmd/
│   └── scheduler/          # Application entry point
│       └── main.go
├── internal/
│   ├── models/            # Data models
│   │   └── post.go
│   ├── scheduler/         # Core scheduling logic
│   │   └── scheduler.go
│   └── cli/              # Command-line interface
│       └── cli.go
├── pkg/
│   └── storage/          # Storage implementations
│       └── json.go
├── go.mod
└── posts.json           # Data storage (created automatically)
```

## Building and Running

### Using Makefile (Recommended):
```bash
make help      # Show all available targets
make build     # Build the binary
make run       # Run the application directly
make run-bin   # Build and run the binary
make clean     # Clean build artifacts
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
make dev       # Format, vet, and build
make fmt       # Format code
make vet       # Run go vet
make test      # Run tests
make tidy      # Tidy modules
```

## LinkedIn API Setup

To use LinkedIn posting features, you need to set up a LinkedIn app:

1. See [LINKEDIN_SETUP.md](LINKEDIN_SETUP.md) for detailed setup instructions
2. Run the app and it will create a `config.json` template
3. Fill in your LinkedIn app credentials
4. Use option 5 to authenticate with LinkedIn
5. Start posting!

## New Menu Options

- **Option 5**: Authenticate with LinkedIn (one-time setup)
- **Option 6**: Publish specific post to LinkedIn
- **Option 7**: Auto-publish all due posts to LinkedIn

## Usage

The application provides an interactive menu with the following options:

1. **Schedule a new post** - Enter content and target date/time
2. **List scheduled posts** - View all posts with their status
3. **Check due posts** - Review posts ready for publishing
4. **Delete a post** - Remove scheduled posts
5. **Authenticate with LinkedIn** - Set up LinkedIn API access
6. **Publish specific post to LinkedIn** - Manually publish a post
7. **Auto-publish all due posts** - Bulk publish ready posts
8. **Exit** - Close the application

## Architecture

- **Models**: Define data structures (Post)
- **Storage**: Handle data persistence (JSON file storage)
- **Scheduler**: Core business logic for managing posts
- **CLI**: User interface and interaction handling
- **CMD**: Application entry point

The application follows Go best practices with clear separation of concerns and modular design.