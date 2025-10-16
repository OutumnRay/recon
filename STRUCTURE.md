# Project Structure - Recontext.online

## Current Directory Structure

```
Recontext.online/
├── .git/                          # Git version control
├── .github/                       # GitHub configuration
│   └── workflows/                 # GitHub Actions workflows
│       └── docker-build-push.yml  # Docker image build and push workflow
├── .idea/                         # JetBrains IDE configuration
├── cmd/                           # Command-line applications (service entry points)
│   ├── managing-portal/           # Managing Portal service
│   │   ├── main.go                # Entry point with auth and Swagger
│   │   ├── handlers_users.go     # User CRUD endpoints
│   │   ├── handlers_groups.go    # Group management endpoints
│   │   └── handlers_metrics.go   # Metrics collection endpoints
│   ├── user-portal/               # User Portal service
│   │   └── main.go                # Entry point with file upload and search
│   ├── jitsi-agent/               # Custom Jitsi recording agent
│   │   └── main.go                # WebRTC recording service
│   ├── transcription-worker/      # Transcription worker service
│   │   └── main.go                # Whisper-based transcription with RabbitMQ
│   └── summarization-worker/      # Summarization worker service
│       └── main.go                # LLM-based summarization with RabbitMQ
├── internal/                      # Private application code
│   ├── config/                    # Configuration management
│   │   └── config.go              # Config loading from environment variables
│   ├── models/                    # Data models
│   │   ├── service.go             # Service models and types
│   │   ├── auth.go                # Authentication models (User, Token, Login)
│   │   ├── user_crud.go           # User CRUD request/response models
│   │   ├── groups.go              # User groups and dynamic permissions
│   │   ├── metrics.go             # Metrics and telemetry models
│   │   ├── media.go               # Media models (Recording, Transcript, Upload)
│   │   └── jitsi.go               # Jitsi models (Session, WebRTC connection)
│   ├── storage/                   # Storage layer (placeholder)
│   └── queue/                     # Message queue integration (placeholder)
├── pkg/                           # Public libraries
│   ├── auth/                      # Authentication utilities
│   │   ├── jwt.go                 # JWT token generation and verification
│   │   └── middleware.go          # Auth middleware for HTTP handlers
│   └── logger/                    # Logging utility
│       └── logger.go              # Simple logger implementation
├── api/                           # API definitions (placeholder)
│   └── proto/                     # gRPC protocol definitions (placeholder)
├── deployments/                   # Deployment configurations
│   └── docker/                    # Dockerfiles
│       ├── Dockerfile.managing-portal
│       ├── Dockerfile.user-portal
│       ├── Dockerfile.jitsi-agent
│       ├── Dockerfile.transcription-worker
│       └── Dockerfile.summarization-worker
├── .dockerignore                  # Docker build ignore file
├── CLAUDE.md                      # Guidelines for Claude Code
├── README.md                      # Project vision and architecture (Russian)
├── READY.md                       # Project progress tracking
├── STRUCTURE.md                   # This file - project structure documentation
├── docker-compose.yml             # Docker Compose configuration for all services
├── go.mod                         # Go module definition
├── go.sum                         # Go dependencies checksum
└── main.go                        # Legacy main entry point (to be removed)
```

## Key Files

### Root Level
- **go.mod**: Go module file defining the module name (`Recontext.online`) and Go version (1.24)
- **go.sum**: Go dependencies checksums for reproducible builds
- **docker-compose.yml**: Orchestration file for all 11 services (managing portal, user portal, RabbitMQ, workers, Jitsi, MinIO, PostgreSQL, Qdrant)
- **.dockerignore**: Files to exclude from Docker builds
- **main.go**: Legacy main entry point (placeholder, to be removed)
- **README.md**: Comprehensive Russian-language documentation of the platform architecture and vision
- **CLAUDE.md**: Development guidelines and instructions for Claude Code
- **READY.md**: Progress tracking - what's done and what's next
- **STRUCTURE.md**: This file - documents the codebase structure

### Services (cmd/)
- **cmd/managing-portal/main.go**: Managing Portal API server (Port 8080)
  - **Authentication**: `/api/v1/auth/login`, `/api/v1/auth/register` (admin only)
  - **Monitoring**: `/health`, `/api/v1/status`
  - **Service Management**: `/api/v1/services`, `/api/v1/services/register`, `/api/v1/services/heartbeat`
  - **User Management** (Admin only):
    - `GET /api/v1/users` - List all users
    - `GET /api/v1/users/{id}` - Get user by ID
    - `PUT /api/v1/users/{id}` - Update user
    - `DELETE /api/v1/users/{id}` - Delete user
    - `PUT /api/v1/users/password` - Change password
  - **Group Management** (Admin only):
    - `GET /api/v1/groups` - List all groups
    - `GET /api/v1/groups/{id}` - Get group by ID
    - `POST /api/v1/groups` - Create new group
    - `PUT /api/v1/groups/{id}` - Update group
    - `DELETE /api/v1/groups/{id}` - Delete group
    - `POST /api/v1/groups/add-user` - Add user to group
    - `POST /api/v1/groups/check-permission` - Check user permissions
  - **Metrics & Telemetry**:
    - `POST /api/v1/metrics` - Receive metrics from services
    - `GET /api/v1/metrics` - Query metrics
    - `GET /api/v1/metrics/system` - Get system-wide metrics
    - `POST /api/v1/logs` - Receive logs from services
  - **Security**: JWT authentication, role-based authorization + dynamic JSON permissions
  - **Swagger**: Full OpenAPI/Swagger documentation
  - Default credentials: admin/admin123
  - Default groups: Editors (read/write), Viewers (read-only)

- **cmd/user-portal/main.go**: User Portal API server (Port 8081)
  - **Authentication**: `/api/v1/auth/login`
  - **Recordings**: `/api/v1/recordings/upload`, `/api/v1/recordings`, `/api/v1/recordings/{id}`
  - **Search**: `/api/v1/search` - semantic and keyword search
  - **Security**: JWT authentication required for all endpoints
  - **Swagger**: Full OpenAPI/Swagger documentation
  - Default credentials: user/user123

- **cmd/jitsi-agent/main.go**: Custom Jitsi Recording Agent (Port 8084)
  - **Recording Control**: `/api/v1/recording/start`, `/api/v1/recording/stop`
  - **Session Management**: `/api/v1/sessions`, `/api/v1/status`
  - **WebRTC**: Connects to Jitsi Meet conferences via WebRTC
  - **Concurrent Recording**: Supports up to 10 simultaneous sessions with goroutines
  - **Swagger**: Full OpenAPI/Swagger documentation

- **cmd/transcription-worker/main.go**: Transcription Worker Service (Port 8082)
  - **Health Check**: `/health` - Service health status
  - **Status**: `/status` - Worker status and statistics
  - **RabbitMQ Integration**: Consumes transcription tasks from queue (pending)
  - **Whisper Processing**: Placeholder for audio/video transcription
  - **Speaker Diarization**: Placeholder for speaker identification
  - **Concurrent Processing**: Task management with goroutines
  - **Swagger**: Full OpenAPI/Swagger documentation

- **cmd/summarization-worker/main.go**: Summarization Worker Service (Port 8083)
  - **Health Check**: `/health` - Service health status
  - **Status**: `/status` - Worker status and statistics
  - **RabbitMQ Integration**: Consumes summarization tasks from queue (pending)
  - **LLM Processing**: Placeholder for transcript summarization
  - **Concurrent Processing**: Task management with goroutines
  - **Swagger**: Full OpenAPI/Swagger documentation

### Internal Packages
- **internal/config/config.go**: Configuration loading from environment variables
  - Database, RabbitMQ, MinIO, Jitsi configurations
  - Helper functions for env parsing
- **internal/models/service.go**: Service-related data models
  - ServiceInfo, ServiceStatus, ServiceType, SystemStatus
- **internal/models/auth.go**: Authentication and authorization models
  - User (with groups and active status), UserRole, LoginRequest/Response, TokenClaims
- **internal/models/user_crud.go**: User management models
  - UpdateUserRequest, ChangePasswordRequest, ListUsersRequest/Response
- **internal/models/groups.go**: User groups and permissions models
  - UserGroup, Permission, PermissionSet, GroupMembership
  - Dynamic JSON-based permissions system
- **internal/models/metrics.go**: Telemetry and metrics models
  - Metric, MetricsBatch, ServiceMetricsSummary, SystemMetrics
  - LogEntry for centralized logging
- **internal/models/media.go**: Media and recording models
  - Recording, Transcript, TranscriptSegment, UploadRequest/Response
  - TranscriptionTask, SummarizationTask - Worker task models
- **internal/models/jitsi.go**: Jitsi-specific models
  - JitsiSession, WebRTCConnection, Recording control requests/responses

### Public Packages
- **pkg/auth/jwt.go**: JWT token management
  - Token generation and verification
  - HMAC SHA256 signing
  - Password hashing (upgrade to bcrypt in production)
- **pkg/auth/middleware.go**: HTTP authentication middleware
  - JWT extraction and validation
  - Role-based access control
  - Context injection for user claims
- **pkg/logger/logger.go**: Simple logging utility with Info, Error, Debug, Fatal levels

### Deployment
- **deployments/docker/**: Dockerfiles for all Go services
  - Multi-stage builds using Go 1.24 and Alpine Linux
  - managing-portal, user-portal, jitsi-agent: Alpine + Go binary
  - Transcription worker: Python + Whisper + Go
  - Summarization worker: Python + transformers + Go
  - Jitsi agent: Alpine + FFmpeg + GStreamer + Go
- **.github/workflows/docker-build-push.yml**: CI/CD pipeline
  - Builds 5 Docker images: managing-portal, user-portal, jitsi-agent, transcription-worker, summarization-worker
  - Tags images as :latest and :<commit-sha>
  - Pushes to Docker Hub (sivanov2018/recontext-*)
  - Uses DOCKER_USERNAME and DOCKER_PASSWORD secrets

## Docker Compose Services

The `docker-compose.yml` defines 11 services in the Recontext.online platform:

1. **managing-portal** (Port 8080): Monitors and manages all services
2. **user-portal** (Port 8081): User-facing API for system interaction
3. **rabbitmq** (Ports 5672, 15672): Message broker with management UI
4. **transcription-worker**: Processes audio/video with Whisper (GPU support)
5. **summarization-worker**: Summarizes transcripts with transformers
6. **rag-service (Qdrant)** (Ports 6333, 6334): Vector database for semantic search
7. **minio** (Ports 9000, 9001): S3-compatible object storage
8. **postgres** (Port 5432): PostgreSQL database for metadata
9. **jitsi-web** (Ports 8443, 8000): Jitsi Meet web interface
10. **jitsi-agent (Custom)** (Port 8084): Custom Go recording agent with WebRTC support
11. **jitsi-prosody, jitsi-jicofo, jitsi-jvb**: Jitsi Meet supporting services

All services are connected via the `recontext-network` bridge network.

## Planned Additions

The following components are planned but not yet implemented:

### Storage Integration
- **internal/storage/**: Clients for PostgreSQL, MinIO, Qdrant
- **migrations/**: Database migration files

### Queue Integration
- **internal/queue/**: RabbitMQ client and message handlers

### Additional Utilities
- **pkg/transcription/**: Whisper integration utilities
- **pkg/diarization/**: Speaker diarization utilities
- **pkg/vectorization/**: Text embedding utilities

### API Definitions
- **api/proto/**: gRPC protocol definitions
- **api/openapi/**: OpenAPI/Swagger specifications

## Navigation Guide

### Finding Components
- **Service entry points**: Look in `cmd/` subdirectories for `main.go` files
- **Configuration**: `internal/config/config.go` - all environment variable configuration
- **Data models**: `internal/models/` - shared types and structures
- **Utilities**: `pkg/` packages - logging and other reusable code
- **Docker builds**: `deployments/docker/` - Dockerfiles for each service
- **CI/CD**: `.github/workflows/` - GitHub Actions workflows
- **Local development**: `docker-compose.yml` - start all services locally

### Running the Project

#### Local Development with Docker Compose
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f managing-portal

# Stop all services
docker-compose down
```

#### Building Individual Services
```bash
# Build managing portal
go build -o bin/managing-portal ./cmd/managing-portal

# Run locally
./bin/managing-portal
```

#### Building Docker Images
```bash
# Build specific service
docker build -f deployments/docker/Dockerfile.managing-portal -t recontext-managing-portal .

# Build all services via docker-compose
docker-compose build
```

### Current Implementation Status
✅ **Fully Implemented**:
- **Managing Portal API**: Monitoring, service registration, heartbeat tracking, user CRUD, groups, metrics
- **User Portal API**: File upload, recording management, search endpoints
- **Jitsi Agent**: Custom WebRTC recording agent with session management
- **Transcription Worker**: Basic structure with health checks and task management
- **Summarization Worker**: Basic structure with health checks and task management
- **Authentication System**: JWT tokens, role-based authorization, password hashing
- **Authorization Middleware**: Protected endpoints, role checking
- **User Management**: Full CRUD operations for users (admin only)
- **Group Management**: Dynamic JSON-based permissions system
- **Metrics & Telemetry**: Centralized metrics collection and system-wide aggregation
- **Data Models**: Service, Auth, Media, Jitsi, Groups, Metrics, User CRUD models
- **Swagger Documentation**: All APIs documented with OpenAPI annotations
- **Configuration System**: Environment-based config management
- **Logging Utility**: Structured logging across all services
- **Dockerfiles**: All 5 services containerized with multi-stage builds
- **Docker Compose**: Full 11-service orchestration
- **CI/CD Pipeline**: GitHub Actions with multi-service builds (main branch only)

⚠️ **Partially Implemented**:
- **Worker Services**: Basic Go structure implemented, pending:
  - RabbitMQ message consumption
  - Actual Whisper transcription processing
  - Actual LLM summarization processing
  - MinIO integration for file storage
  - PostgreSQL integration for metadata

⚠️ **Not Yet Implemented**:
- Storage layer integrations (PostgreSQL, MinIO, Qdrant clients)
- RabbitMQ queue client implementation
- Actual Whisper/LLM processing logic
- Database persistence (currently in-memory for managing portal)
- End-to-end processing pipeline
