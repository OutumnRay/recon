# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Recontext.online is an on-premise/edge-cloud platform for audio/video processing, speech recognition, and semantic search of conversations. The platform aims to provide:

- Automated audio/video processing (including streaming)
- Speech recognition with diarization (speaker identification) and timestamps
- Contextual search across transcripts and semantic queries
- Automatic summarization and tagging of conversations
- Video conference recording/analysis with Jitsi integration
- Integration with external communication systems

**Language**: Go (version 1.24)

## Architecture Overview

The platform is designed with the following modular architecture:

### 1. Media Ingestion Layer
- Accepts audio/video files or streams (MP4, WAV, WebRTC, RTSP)
- Message queue for scalable processing (Kafka/NATS/RabbitMQ)
- Integration with telephony systems (Asterisk, 3CX, Twilio, SIP)

### 2. Speech Processing
- Whisper (GPU) for multi-language speech recognition
- PyAnnote/NeMo/Silero for speaker diarization
- Generates transcripts with timestamps in JSON+SRT format

### 3. Semantic Layer
- Vectorization using BERT/InstructorXL/OpenAI Embeddings/local LLM
- Storage in Qdrant or Milvus for semantic search
- Contextual search capabilities ("who talked about budget?", "what was decided on the project?")

### 4. Summarization & Insight Extraction
- Automatic summarization by topics/speakers
- Keywords, tags, sentiment analysis
- Structured meeting minutes generation

### 5. Jitsi Integration
- Conference creation/joining from Recontext
- Recording and storage of sessions (audio/video + transcripts)
- Optional bot-recorder for existing calls

### 6. Storage Layer
- MinIO or S3-compatible storage for media
- PostgreSQL/ClickHouse for metadata and analytics
- Containerization via Docker Compose or Kubernetes

### 7. Integration Layer
- REST/gRPC API for accessing records, transcripts, search
- WebSocket API for streaming processing
- External integrations: Asterisk/3CX/Twilio, CRMs (Bitrix24, amoCRM, Salesforce), MS Teams/Zoom/Telegram/WhatsApp

## Development Commands

### Build and Run
```bash
# Build the project
go build -o recontext main.go

# Run the project
go run main.go

# Run with specific Go version
go1.24 run main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestName ./path/to/package
```

### Code Quality
```bash
# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Run linter (if golangci-lint is installed)
golangci-lint run
```

### Dependencies
```bash
# Download dependencies
go mod download

# Tidy dependencies
go mod tidy

# Vendor dependencies
go mod vendor
```

## Documentation Maintenance

**IMPORTANT**: When working on this project, you must maintain two critical documentation files:

### READY.md
Always read and update this file to track project progress:
- **What has been completed**: List all implemented features, modules, and tasks
- **What needs to be done next**: Describe the next steps, pending tasks, and upcoming work
- Update this file after completing any significant work
- Read this file at the start of each session to understand current progress

### STRUCTURE.md
Maintain this file to document the project structure:
- **Directory structure**: Document the organization of the codebase
- **Module locations**: Brief descriptions of where to find specific functionality
- **Key files**: Note important configuration files, entry points, and core modules
- **Navigation guide**: Help future developers quickly locate components
- **CRITICAL**: After ANY operation that changes the project structure (creating directories, adding new files, moving files, creating new packages/modules, reorganizing code), you MUST immediately update STRUCTURE.md to reflect these changes

Always read both files before starting work and update them as the project evolves.

DO NOT CREATE ANOTHER MD FIELS!
AFTER MAKE CHANGES FIX IT ON GIT COMMIT AND DESCRIBE RUSSIAN MESSAGE!

## Database Access and ORM

**CRITICAL**: This project uses GORM ORM for all database operations. You MUST follow these rules:

### GORM Usage Rules

1. **NEVER use raw SQL queries** (`db.DB.Raw()`, `db.DB.Exec()`, or any SQL string interpolation)
2. **ALWAYS use GORM methods** for all database operations:
   - `db.DB.Table()` - for table selection
   - `db.DB.Select()` - for column selection
   - `db.DB.Where()` - for filtering
   - `db.DB.Joins()` - for table joins
   - `db.DB.Order()` - for sorting
   - `db.DB.Create()` - for inserts
   - `db.DB.Updates()` - for updates
   - `db.DB.Delete()` - for deletes
   - `db.DB.First()`, `db.DB.Find()`, `db.DB.Scan()` - for queries

3. **Models location**: All database models are defined in `internal/models/` and `pkg/database/models.go`

4. **Example of CORRECT GORM usage**:
```go
// Query with joins
var results []struct {
    TrackID         string
    ParticipantName string
    PlaylistURL     string
}
err := db.DB.
    Table("livekit_tracks").
    Select("livekit_tracks.id as track_id, COALESCE(p.name, 'Unknown') as participant_name, er.playlist_url").
    Joins("LEFT JOIN livekit_participants p ON livekit_tracks.participant_sid = p.sid").
    Joins("LEFT JOIN livekit_egress_recordings er ON er.track_sid = livekit_tracks.sid").
    Where("livekit_tracks.room_sid = ?", roomSID).
    Where("livekit_tracks.type IN ?", []string{"audio", "video"}).
    Scan(&results).Error

// Update
db.DB.Table("meetings").
    Where("id = ?", meetingID).
    Updates(map[string]interface{}{
        "video_playlist_url": playlistURL,
        "updated_at":         time.Now(),
    })
```

5. **Example of INCORRECT usage** (DO NOT DO THIS):
```go
// ❌ WRONG - Never use Raw or Exec
db.DB.Raw("SELECT * FROM meetings WHERE id = ?", id).Scan(&result)
db.DB.Exec("UPDATE meetings SET title = ? WHERE id = ?", title, id)
```

### Why GORM Only?

- **Security**: Automatic SQL injection prevention through parameter binding
- **Maintainability**: Consistent API across the codebase
- **Type safety**: Compile-time checking with Go models
- **Database portability**: Easy to switch database backends
- **Testing**: Easier to mock and test

If you need to write a database query, ALWAYS check existing GORM code in the project for patterns and examples.

## Development Notes

### Current State
The project is in its initial stage with only a basic Go module setup. The main.go file contains a placeholder "Hello World" program. Implementation of the architecture described above is pending.

### Future Implementation Considerations
- Multi-service architecture likely required (separate services for ingestion, processing, search, API)
- Consider microservices pattern with shared message queue
- GPU support needed for Whisper model inference
- Vector database integration for semantic search
- Object storage integration for media files
- Database design for metadata, users, and analytics
