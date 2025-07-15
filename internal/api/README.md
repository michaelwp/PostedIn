# Internal API Package

This package contains the unified Web API implementation for the LinkedIn Post Scheduler using the Fiber framework. It includes both REST API endpoints and OAuth callback handling in a single server.

## Structure

```
internal/api/
├── README.md          # This file
├── router.go          # Main router setup and middleware
├── posts.go           # Posts management endpoints
├── auth.go            # Authentication endpoints
├── timezone.go        # Timezone configuration endpoints
└── scheduler.go       # Scheduler status endpoints
```

## Architecture

### Router (`router.go`)
- **Purpose**: Central router configuration and middleware setup
- **Key Components**:
  - `Router` struct holds all dependencies (config, scheduler, cron)
  - `NewRouter()` creates a new router instance
  - `SetupRoutes()` configures all API routes with middleware
  - Includes CORS and logging middleware

### Posts (`posts.go`)
- **Purpose**: Handle all post-related operations
- **Endpoints**:
  - `GET /api/posts` - List all posts (sorted by scheduled time)
  - `POST /api/posts` - Create new post
  - `GET /api/posts/:id` - Get specific post
  - `PUT /api/posts/:id` - Update post
  - `DELETE /api/posts/:id` - Delete specific post
  - `DELETE /api/posts` - Delete multiple posts
  - `GET /api/posts/due` - Get posts ready for publishing
  - `POST /api/posts/:id/publish` - Publish specific post
  - `POST /api/posts/publish-due` - Publish all due posts

### Authentication (`auth.go`)
- **Purpose**: Handle LinkedIn authentication and OAuth callbacks
- **Endpoints**:
  - `GET /api/auth/linkedin` - Get LinkedIn OAuth URL
  - `GET /api/auth/status` - Check authentication status
  - `GET /api/auth/debug` - Debug authentication issues
- **OAuth Callback Routes**:
  - `GET /` - Authentication home page with LinkedIn auth button
  - `GET /callback` - OAuth callback handler for LinkedIn authorization

### Timezone (`timezone.go`)
- **Purpose**: Manage timezone configuration
- **Endpoints**:
  - `GET /api/timezone` - Get current timezone
  - `POST /api/timezone` - Update timezone

### Scheduler (`scheduler.go`)
- **Purpose**: Monitor auto-scheduler status
- **Endpoints**:
  - `GET /api/scheduler/status` - Get scheduler status and next run time

## Features

### Unified Server
- **Single Server**: Combines REST API and OAuth callback handling
- **No Separate Callback Server**: All functionality in one unified server
- **Shared Configuration**: Uses the same config and dependencies throughout

### Middleware
- **CORS**: Enables cross-origin requests for web clients
- **Logging**: Structured request logging with timing
- **Error Handling**: Consistent error response format

### Response Format
All endpoints return JSON responses with consistent structure:
```json
{
  "success": true,
  "data": {...},
  "message": "Optional message"
}
```

Error responses:
```json
{
  "success": false,
  "error": "Error description"
}
```

### Validation
- Input validation for all POST/PUT requests
- Date/time format validation
- Business logic validation (e.g., no past scheduling)

### OAuth Integration
- **Complete OAuth Flow**: Full LinkedIn OAuth 2.0 implementation
- **Beautiful UI**: Styled authentication pages with error handling
- **Security**: Proper state validation and error handling
- **Token Management**: Automatic token saving and profile retrieval

### Integration
- Seamless integration with existing scheduler and cron components
- Automatic cron scheduler updates when posts are created/updated
- Timezone-aware operations

## Usage

### Starting the Server
```bash
# Using Makefile
make build-web-api
make run-web-api-bin

# Or directly
go run cmd/web-api/main.go
```

### Configuration
The API uses the same configuration as the CLI application:
- `config.json` for LinkedIn credentials and settings
- `posts.json` for post storage
- `linkedin_token.json` for OAuth tokens

### Development
The organized structure makes it easy to:
- Add new endpoints in the appropriate handler file
- Modify existing endpoints
- Add new middleware
- Extend functionality

## Dependencies

- **Fiber v2**: Fast HTTP framework
- **CORS Middleware**: Cross-origin request handling
- **Logger Middleware**: Request logging
- **Internal Packages**: config, scheduler, cron, models