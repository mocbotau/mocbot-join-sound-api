# MOCBOT Join Sound API

A custom API for allowing users to upload and manage custom join sounds for Discord servers.

## Project Structure

```
.
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── database/         # Database layer
│   ├── handlers/         # HTTP handlers
│   ├── middleware/       # HTTP middlewares (including Auth0)
│   ├── models/           # Data models
│   └── utils/            # Utility functions
├── data/                 # Database and file storage
```

## Authentication

This API uses **Auth0** for authentication and authorization:

- JWT tokens are required for protected endpoints
- The JWT subject field contains Discord user ID for authorization
- Public endpoints (like file retrieval) don't require authentication
- Protected endpoints validate user ownership of resources

## API Endpoints

### Public Endpoints

#### Health Check

```bash
GET /api/v1/ping
```

#### Get Sound File

```bash
GET /api/v1/sound/:soundId
```

#### Get User Sounds

```bash
GET /api/v1/sounds/:guildId/:userId
```

#### Get User Settings

```bash
GET /api/v1/settings/:guildId/:userId
```

### Protected Endpoints (Require Auth0 JWT)

#### Upload User Sounds

```bash
POST /api/v1/sounds/:guildId/:userId
Content-Type: multipart/form-data
Authorization: Bearer <jwt-token>

Parameters:
- files: Audio files to upload (max 5 files, max 10MB each)
```

#### Delete Sound

```bash
DELETE /api/v1/sound/:soundId
Authorization: Bearer <jwt-token>
```

#### Update User Settings

```bash
PATCH /api/v1/settings/:guildId/:userId
Content-Type: application/json
Authorization: Bearer <jwt-token>

Body:
{
  "active_sound_id": "sound-id-here",
  "mode": "enabled/disabled"
}
```

## File Constraints

- **Maximum file size**: 10MB per file
- **Maximum files per user**: 5 files
- **Maximum audio duration**: 5 seconds
- **Supported formats**: MP3 (.mp3), WAV (.wav)
- **Maximum payload size**: 50MB

## Development

### Prerequisites

- Go 1.25+
- SQLite3
- Auth0 account and application setup

### Environment Variables

The following environment variables are required:

#### Required Variables

- `AUTH0_DOMAIN`: Your Auth0 domain (e.g., `your-domain.auth0.com`)
- `AUTH0_AUDIENCE`: Your Auth0 API audience identifier

#### Optional Variables

- `PORT`: Server port (default: `8081`)
- `DB_PATH`: SQLite database path (default: `./data/main.db`)
- `SOUNDS_PATH`: Directory for storing sound files (default: `./data/sounds`)
- `GIN_MODE`: Gin framework mode (`debug`, `release`, default: `debug`)

### Local Development

1. Set up your Auth0 application and get your domain and audience
2. Create a `.env` file with your Auth0 credentials:

```bash
AUTH0_DOMAIN=your-domain.auth0.com
AUTH0_AUDIENCE=your-api-audience
```

3. Install dependencies and run:

```bash
# Install dependencies
go mod tidy

# Run the application
go run cmd/api/main.go
```

### Linting and testing

To ensure code quality, you can run linting and testing commands:

```bash
# Run linting
make lint

# Run format
make fmt

# Run tests
go test ./...
```

### Docker Development

```bash
# Make sure your .env file is in the project root
# Build and run with docker-compose
docker-compose up --build -d
```

The API will be available at `http://localhost:8081`
