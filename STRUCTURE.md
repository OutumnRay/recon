# Project Structure - Recontext.online

## Current Directory Structure

```
Recontext.online/
├── .git/                 # Git version control
├── .idea/                # JetBrains IDE configuration
├── CLAUDE.md             # Guidelines for Claude Code
├── README.md             # Project vision and architecture (Russian)
├── READY.md              # Project progress tracking
├── STRUCTURE.md          # This file - project structure documentation
├── go.mod                # Go module definition
└── main.go               # Main entry point (placeholder)
```

## Key Files

### Root Level
- **go.mod**: Go module file defining the module name (`Recontext.online`) and Go version (1.24)
- **main.go**: Main entry point - currently contains placeholder "Hello World" code
- **README.md**: Comprehensive Russian-language documentation of the platform architecture and vision
- **CLAUDE.md**: Development guidelines and instructions for Claude Code
- **READY.md**: Progress tracking - what's done and what's next
- **STRUCTURE.md**: This file - documents the codebase structure

## Planned Directory Structure

The following structure is planned but not yet implemented:

```
Recontext.online/
├── cmd/                          # Command-line applications
│   ├── api-gateway/             # API Gateway service
│   ├── ingestion-service/       # Media ingestion service
│   ├── speech-service/          # Speech processing service
│   └── semantic-service/        # Semantic search service
├── internal/                     # Private application code
│   ├── config/                  # Configuration management
│   ├── storage/                 # Storage layer (DB, S3, Vector DB)
│   ├── queue/                   # Message queue integration
│   └── models/                  # Data models
├── pkg/                         # Public libraries
│   ├── transcription/           # Transcription utilities
│   ├── diarization/             # Speaker diarization
│   └── vectorization/           # Text vectorization
├── api/                         # API definitions
│   ├── rest/                    # REST API specs (OpenAPI/Swagger)
│   ├── grpc/                    # gRPC protocol definitions
│   └── websocket/               # WebSocket protocol definitions
├── scripts/                     # Build and deployment scripts
├── deployments/                 # Deployment configurations
│   ├── docker/                  # Dockerfiles
│   └── kubernetes/              # K8s manifests
├── migrations/                  # Database migrations
├── configs/                     # Configuration files
└── docs/                        # Additional documentation
```

## Module Locations (Planned)

### Core Services
- **API Gateway**: `cmd/api-gateway/` - REST/gRPC/WebSocket endpoints
- **Media Ingestion**: `cmd/ingestion-service/` - Accepts and queues media files
- **Speech Processing**: `cmd/speech-service/` - Whisper + diarization
- **Semantic Search**: `cmd/semantic-service/` - Vector search and embeddings

### Shared Libraries
- **Configuration**: `internal/config/` - Application configuration
- **Storage Abstractions**: `internal/storage/` - DB, S3, Vector DB clients
- **Queue Integration**: `internal/queue/` - Message queue clients
- **Data Models**: `internal/models/` - Shared data structures

### Public Packages
- **Transcription**: `pkg/transcription/` - Reusable transcription code
- **Diarization**: `pkg/diarization/` - Speaker identification
- **Vectorization**: `pkg/vectorization/` - Text embedding utilities

## Navigation Guide

### Finding Components (When Implemented)
- **Entry points**: Look in `cmd/` subdirectories for `main.go` files
- **Business logic**: Check `internal/` packages (not importable from outside)
- **Reusable utilities**: Find in `pkg/` (can be imported by external projects)
- **API contracts**: Review `api/` for REST/gRPC/WebSocket definitions
- **Database schema**: Check `migrations/` directory
- **Deployment configs**: Look in `deployments/docker/` or `deployments/kubernetes/`

### Current State
⚠️ **Note**: Most of the structure above is planned but not yet created. Currently, the project only contains:
- Root level configuration files (go.mod)
- Single main.go with placeholder code
- Documentation files (README.md, CLAUDE.md, READY.md, STRUCTURE.md)

The actual implementation of the directory structure will happen in the next development phase.
