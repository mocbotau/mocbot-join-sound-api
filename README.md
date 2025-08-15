# MOCBOT Join Sound API

A custom API for allowing users to upload and manage custom join sounds.

## Project Structure

```
.
├── cmd/api/               # Application entrypoint
├── internal/
│   ├── database/          # Database layer
│   ├── handlers/          # HTTP handlers
│   ├── models/           # Data models
│   └── utils/            # Utility functions
├── files/                # Uploaded files storage
├── data/                 # Database files
├── docker-compose.yml    # Docker composition
└── Dockerfile           # Container definition
```

## API Endpoints

### Upload Files

```bash
POST /api/v1/upload
Content-Type: multipart/form-data

Parameters:
- guild_id: Discord guild ID
- user_id: Discord user ID
- files: Audio files to upload
```

### Get File

```bash
GET /api/v1/file/:fileId
```

### Get User Files

```bash
GET /api/v1/files/:guildId/:userId
```

### Health Check

```bash
GET /api/v1/ping
```

## Development

### Prerequisites

- Go 1.25+
- SQLite3

### Local Development

```bash
# Install dependencies
go mod tidy

# Run the application
go run cmd/api/main.go
```

### Docker Development

```bash
# Build and run with docker-compose
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop services
docker-compose down
```

## Environment Variables

- `PORT`: Server port (default: 8081)
- `DB_PATH`: SQLite database path (default: ./data/sounds.db)
- `GIN_MODE`: Gin mode (development/release)
