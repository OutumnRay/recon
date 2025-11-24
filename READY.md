# Project Progress - Recontext.online

## What Has Been Completed

### Initial Setup
- ✅ Created Go module (`go.mod`) with Go 1.24
- ✅ Created basic `main.go` with placeholder "Hello World" program
- ✅ Initialized Git repository
- ✅ Created comprehensive `README.md` with full architecture vision (in Russian)
- ✅ Created `CLAUDE.md` with development guidelines and architecture overview
- ✅ Created `STRUCTURE.md` to document project structure
- ✅ Created this `READY.md` file for progress tracking

### Documentation
- ✅ Documented 7-layer architecture in README.md:
  - Media Ingestion Layer
  - Speech Processing
  - Semantic Layer
  - Summarization & Insight Extraction
  - Jitsi Integration
  - Storage Layer
  - Integration Layer

### Project Structure (Phase 1 - Completed)
- ✅ Created complete Go project directory structure:
  - `cmd/` for service entry points (managing-portal, user-portal, workers)
  - `internal/` for private application code (config, models, storage, queue)
  - `pkg/` for public libraries (logger)
  - `api/` for API definitions (proto)
  - `deployments/docker/` for Dockerfiles
  - `.github/workflows/` for CI/CD

### Configuration and Infrastructure
- ✅ Implemented configuration management (`internal/config/config.go`)
  - Environment variable loading
  - Database, RabbitMQ, MinIO, Jitsi configurations
- ✅ Created data models (`internal/models/service.go`)
  - ServiceInfo, ServiceStatus, ServiceType
  - SystemStatus, HealthResponse, ErrorResponse
- ✅ Implemented simple logger utility (`pkg/logger/logger.go`)

### Managing Portal Service (Completed)
- ✅ Implemented Managing Portal API (`cmd/managing-portal/main.go`)
  - Health check endpoint: `GET /health`
  - System status endpoint: `GET /api/v1/status`
  - Service management endpoints: `GET /api/v1/services`
  - Service registration: `POST /api/v1/services/register`
  - Service heartbeat: `POST /api/v1/services/heartbeat`
  - Automatic heartbeat monitoring (2-minute timeout)
  - Runs on port 8080

### Docker Infrastructure (Completed)
- ✅ Created Dockerfiles for all Go services:
  - `Dockerfile.managing-portal` - Multi-stage build with Go 1.24
  - `Dockerfile.user-portal` - Alpine-based runtime
  - `Dockerfile.transcription-worker` - Python + Whisper + Go
  - `Dockerfile.summarization-worker` - Python + transformers + Go
- ✅ Created comprehensive `docker-compose.yml` with 16 services:
  1. managing-portal (Port 8080)
  2. user-portal (Port 8081)
  3. rabbitmq (Ports 5672, 15672, 15692)
  4. transcription-worker (GPU support)
  5. summarization-worker
  6. rag-service/Qdrant (Ports 6333, 6334)
  7. minio (Ports 9000, 9001)
  8. postgres (Port 5432) with pgvector extension
  9. ollama (Port 11434) - Self-hosted LLM and embeddings
  10. jitsi-web (Ports 8443, 8000)
  11. jitsi-agent (Port 8084) - Custom WebRTC recording agent
  12. jitsi-prosody, jitsi-jicofo, jitsi-jvb
  13. prometheus (Port 9090) - Metrics collection
  14. grafana (Port 3000) - Metrics visualization
  15. cadvisor (Port 8089) - Container metrics
  16. postgres-exporter (Port 9187) - PostgreSQL metrics
  17. jitsi-exporter (Port 9888) - Jitsi metrics
  18. watchtower - Automatic container updates from Docker Hub
- ✅ Created `.dockerignore` for optimized builds

### CI/CD Pipeline (Completed)
- ✅ Created GitHub Actions workflow (`.github/workflows/docker-build-push.yml`)
  - Builds Docker images on push to main
  - Multi-service build matrix (5 services)
  - Tags images as `:latest` and `:<commit-sha>`
  - Pushes to Docker Hub: `sivanov2018/recontext-<service-name>`
  - Uses `DOCKER_USERNAME` and `DOCKER_PASSWORD` secrets
  - Build caching for faster builds
  - `fail-fast: false` - continues building other services on failure
  - Per-service build summaries with status indicators
  - Detailed error reporting and troubleshooting tips
  - Fixed: Removed unused imports from `pkg/database/rag_repository.go`

### Phase 2: Authentication & Authorization (Completed)
- ✅ Implemented JWT-based authentication system (`pkg/auth/jwt.go`)
  - Token generation with HMAC SHA256 signing
  - Token verification and expiration checking
  - Password hashing (SHA256, upgrade to bcrypt recommended)
- ✅ Created authentication middleware (`pkg/auth/middleware.go`)
  - Bearer token extraction and validation
  - Role-based access control (RBAC)
  - Context injection for user claims
- ✅ Created comprehensive auth models (`internal/models/auth.go`)
  - User model with roles (Admin, User, Operator, Service)
  - LoginRequest/Response, RegisterRequest
  - TokenClaims, UserInfo structures
- ✅ Added Swagger/OpenAPI documentation to all APIs
  - Full API documentation with request/response examples
  - BearerAuth security scheme definition
  - Tags and descriptions for all endpoints

### Phase 2: User Portal Service (Completed)
- ✅ Implemented User Portal API (`cmd/user-portal/main.go`)
  - Authentication endpoint: `POST /api/v1/auth/login`
  - File upload: `POST /api/v1/recordings/upload` (multipart/form-data)
  - List recordings: `GET /api/v1/recordings`
  - Get recording details: `GET /api/v1/recordings/{id}`
  - Semantic search: `GET /api/v1/search`
  - Full Swagger documentation
  - Runs on port 8081
  - Default credentials: user/user123
- ✅ Created media models (`internal/models/media.go`)
  - Recording model with status tracking
  - Transcript and TranscriptSegment models
  - Upload/List/Search request/response structures

### Phase 2: Jitsi Recorder Service (Completed)
- ✅ Implemented Jitsi Recorder Lite (`jitsy/jitsi-recorder-lite/recorder.py`):
  - **Stream-based recording**: each participant connection = separate file (one stream = one file)
  - **Reconnection handling**: new file created on each reconnect with `isReconnection` flag
  - **Immediate S3 upload**: files uploaded right after participant leaves (not waiting for conference end)
  - **Smart reconnect detection**: automatically detects reconnections within 30 seconds
  - Filename format: `{room}_{participant_id}_{timestamp}.opus`
  - Folder structure: `recordings/{room}/{conference_id}/`
  - S3 metadata: participant ID, duration, join offset, reconnection status
  - Webhook notifications on conference end with all recording details
  - Redis coordination for multi-instance deployment
  - Health check endpoint: `GET /health`
- ✅ Fixed race conditions and error handling:
  - Fixed `ProcessLookupError` when stopping already-terminated FFmpeg processes
  - Fixed `KeyError: 'focus'` when system components leave conference
  - Added system component filtering (skip 'focus' endpoint in joins)
  - Changed missing file errors to warnings (expected when FFmpeg fails)
  - Added safe dictionary deletion with try/except
  - Process state checking before termination
- ✅ Enhanced S3 integration:
  - Detailed S3 initialization logging with endpoint and bucket info
  - S3 connection test on startup with automatic bucket creation
  - Enhanced upload logging with file size and path details
  - Comprehensive error reporting with S3 endpoint, bucket, and key
  - Conference end summary with S3 path and uploaded recordings list
- ✅ Created comprehensive documentation:
  - `jitsy/jitsi-recorder-lite/README.md` with full documentation
  - Stream-based recording explanation with reconnection examples
  - Configuration guide with MinIO port clarification (9000 vs 9001)
  - File structure, metadata format, webhook payload examples
  - Troubleshooting guide for common issues (S3 port, MinIO access)
  - Health check and monitoring documentation
- ✅ Added MinIO to docker-compose.yml:
  - MinIO service with ports 9000 (API) and 9001 (Console)
  - Volume for persistent storage
  - Health check configuration
  - Recorder depends on MinIO

### Phase 2: Jitsi Agent Service (Completed)
- ✅ Implemented custom Jitsi Agent (`cmd/jitsi-agent/main.go`)
  - Start recording: `POST /api/v1/recording/start`
  - Stop recording: `POST /api/v1/recording/stop`
  - List sessions: `GET /api/v1/sessions`
  - Agent status: `GET /api/v1/status`
  - Concurrent recording support (up to 10 sessions)
  - WebRTC connection management with goroutines
  - Full Swagger documentation
  - Runs on port 8084
- ✅ Created Jitsi models (`internal/models/jitsi.go`)
  - JitsiSession, JitsiSessionStatus
  - WebRTCConnection state tracking
  - Recording control request/response structures
- ✅ Created Dockerfile for Jitsi Agent
  - Alpine + FFmpeg + GStreamer for media processing
  - Multi-stage build with Go 1.24
- ✅ Updated docker-compose.yml to use custom Jitsi Agent
  - Replaced standard Jibri with custom WebRTC agent
  - Integrated with managing portal and storage

### Updated Infrastructure
- ✅ Updated Managing Portal with authentication
  - Login endpoint: `POST /api/v1/auth/login`
  - Register endpoint: `POST /api/v1/auth/register` (admin only)
  - Protected endpoints with JWT validation
  - Role-based authorization for admin operations
  - Default admin credentials: admin/admin123
- ✅ Updated CI/CD pipeline
  - Added jitsi-agent to build matrix
  - Now builds 5 services total
- ✅ Updated Docker infrastructure
  - Added Dockerfile.jitsi-agent
  - Updated docker-compose with custom agent
  - All 11 services properly orchestrated

### Phase 3: User Management & Metrics (Completed)
- ✅ Implemented User CRUD endpoints (`cmd/managing-portal/handlers_users.go`)
  - List users: `GET /api/v1/users` - Paginated user listing with filters
  - Get user by ID: `GET /api/v1/users/{id}` - Detailed user information
  - Update user: `PUT /api/v1/users/{id}` - Modify email, role, groups, active status
  - Delete user: `DELETE /api/v1/users/{id}` - Soft delete user accounts
  - Change password: `PUT /api/v1/users/password` - Secure password updates
  - All endpoints admin-only with full Swagger documentation
- ✅ Implemented User Groups system (`cmd/managing-portal/handlers_groups.go`)
  - List groups: `GET /api/v1/groups` - View all user groups
  - Get group: `GET /api/v1/groups/{id}` - Detailed group information
  - Create group: `POST /api/v1/groups` - Define new groups with permissions
  - Update group: `PUT /api/v1/groups/{id}` - Modify group settings and permissions
  - Delete group: `DELETE /api/v1/groups/{id}` - Remove groups
  - Add user to group: `POST /api/v1/groups/add-user` - Group membership management
  - Check permission: `POST /api/v1/groups/check-permission` - Permission validation
  - Dynamic JSON-based permissions system
  - Resource-action-scope permission model (all, own, group)
  - Default groups created: Editors (read/write), Viewers (read-only)
- ✅ Implemented Metrics & Telemetry Collection (`cmd/managing-portal/handlers_metrics.go`)
  - Send metrics: `POST /api/v1/metrics` - Services send telemetry data
  - Query metrics: `GET /api/v1/metrics` - Filter by service_id and metric name
  - System metrics: `GET /api/v1/metrics/system` - Aggregated system-wide statistics
  - Send logs: `POST /api/v1/logs` - Centralized logging from all services
  - In-memory storage with aggregation and analysis
  - Tracks requests, errors, latency, custom metrics per service
- ✅ Created comprehensive data models
  - `internal/models/user_crud.go` - User management request/response models
  - `internal/models/groups.go` - UserGroup, Permission, PermissionSet structures
  - `internal/models/metrics.go` - Metric, SystemMetrics, ServiceMetricsSummary, LogEntry
  - Consistent structure across all new endpoints
- ✅ Created permission documentation (`docs/PERMISSIONS_EXAMPLE.md`)
  - JSON permission structure examples for different roles
  - Editors, Viewers, Team Managers, Transcription Operators examples
  - Advanced custom rules with filters
  - API usage examples with curl commands
  - Permission scopes and available actions documentation
  - Best practices for permission management
- ✅ Updated CI/CD pipeline for main-only deployments
  - Modified `.github/workflows/docker-build-push.yml`
  - Added conditional checks: `github.ref == 'refs/heads/main'`
  - Docker Hub pushes only from main branch
  - Feature branches build but don't push images
  - Prevents accidental deployments from development branches

### Phase 4: Worker Services Structure (Completed)
- ✅ Implemented Transcription Worker structure (`cmd/transcription-worker/main.go`)
  - Health check endpoint: `GET /health` - Service health status
  - Status endpoint: `GET /status` - Worker statistics and task count
  - Worker loop with goroutine-based task management
  - Placeholder for RabbitMQ message consumption
  - Placeholder for Whisper transcription processing
  - Placeholder for speaker diarization
  - Runs on port 8082
  - Full Swagger documentation
  - Successfully compiles and builds
- ✅ Implemented Summarization Worker structure (`cmd/summarization-worker/main.go`)
  - Health check endpoint: `GET /health` - Service health status
  - Status endpoint: `GET /status` - Worker statistics and task count
  - Worker loop with goroutine-based task management
  - Placeholder for RabbitMQ message consumption
  - Placeholder for LLM summarization processing
  - Runs on port 8083
  - Full Swagger documentation
  - Successfully compiles and builds
- ✅ Created worker task models (`internal/models/media.go`)
  - TranscriptionTask - Task structure for transcription queue
  - SummarizationTask - Task structure for summarization queue
  - Status tracking, timestamps, error handling
- ✅ Updated Transcript model
  - Added Status, ProcessedAt, DurationSecs fields
  - Enhanced for worker integration

### Phase 6: Department Management System (Completed)

#### Backend
- ✅ Created Department data models (`internal/models/department.go`)
  - Department model with hierarchical structure (parent_id, level, path)
  - DepartmentTreeNode for tree visualization
  - DepartmentWithStats for statistics display
  - CreateDepartmentRequest/UpdateDepartmentRequest models
  - ListDepartmentsRequest/Response for paginated results
- ✅ Updated User model with department support (`internal/models/auth.go`)
  - Added DepartmentID field to link users to departments
  - Created UserPermissions struct with meeting permissions:
    - can_schedule_meetings: Permission to schedule video meetings
    - can_manage_department: Permission to manage department
    - can_approve_recordings: Permission to approve recordings
  - Updated UserInfo to include department and permissions
- ✅ Implemented database migrations (`pkg/database/database.go`)
  - `departments` table with hierarchical structure support
  - Materialized path pattern for efficient tree queries
  - Level tracking for depth-based operations
  - Soft delete support (is_active flag)
  - Root department created by default ("Organization")
  - Added department_id column to users table
  - Added permissions JSONB column to users table
  - Default permissions for new users (all false)
  - Admin users get all permissions by default
- ✅ Created DepartmentRepository (`pkg/database/department_repository.go`)
  - Create: Automatic level and path calculation
  - GetByID: Retrieve department by ID
  - List: List departments with optional filters (parent, active status)
  - GetTree: Build hierarchical tree structure
  - Update: Update with automatic path recalculation for children
  - updateChildrenPaths: Recursive path updates for descendants
  - Delete: Soft delete with user validation
  - GetWithStats: Department with user count and child count statistics
  - GetChildren: Get direct child departments
  - NameExists: Check name uniqueness within level
  - Circular reference prevention in hierarchy
- ✅ Implemented Department API handlers (`cmd/managing-portal/handlers_department.go`)
  - `POST /api/v1/departments` - Create new department
  - `GET /api/v1/departments` - List departments (flat or tree)
  - `GET /api/v1/departments/{id}` - Get department (with optional stats)
  - `PUT /api/v1/departments/{id}` - Update department
  - `DELETE /api/v1/departments/{id}` - Soft delete department
  - `GET /api/v1/departments/{id}/children` - Get child departments
  - All endpoints admin-only with full Swagger documentation
  - Name uniqueness validation within same level
  - User count validation before deletion
- ✅ Updated user handlers (`cmd/managing-portal/handlers_users.go`)
  - Added department_id field to user responses
  - Added permissions field to user responses
  - Support for updating user department assignment
  - Support for updating user permissions
  - Default permissions set on user creation
- ✅ Integrated with Managing Portal (`cmd/managing-portal/main.go`)
  - Added DepartmentRepository to ManagingPortal struct
  - Registered all department endpoints
  - All routes protected with JWT authentication and admin role
  - Backend successfully compiles and builds

#### Frontend (Managing Portal UI)
- ✅ Created Department API service (`front/managing-portal/src/services/departments.ts`)
  - TypeScript interfaces for Department, DepartmentTreeNode, DepartmentWithStats
  - API methods for all department operations
  - Type-safe request/response models
  - Automatic token injection for authenticated requests
- ✅ Implemented Departments component (`front/managing-portal/src/components/Departments.tsx`)
  - Dual view modes: Tree view and List view
  - Tree view:
    - Hierarchical tree display with expand/collapse
    - Visual depth indication with indentation
    - Level badges showing hierarchy depth
    - Click to select and view details
  - List view:
    - Grid of department cards
    - Shows name, level, path, description
    - Responsive card layout
  - Department details panel:
    - Description and hierarchy information
    - Statistics: direct users, child departments, total users (including sub-departments)
    - Active/inactive status badge
  - Create/Edit modal:
    - Form with name, description, parent department selection
    - Parent department dropdown with full paths
    - Validation and error handling
  - Inline edit/delete actions:
    - Edit button opens pre-filled modal
    - Delete with confirmation
    - Cannot delete departments with active users
  - Real-time updates after create/edit/delete operations
- ✅ Integrated Departments into navigation
  - Added "Departments" menu item with building icon in sidebar
  - Added route in App.tsx: `/departments`
  - Active state highlighting for departments section
  - Updated page title in Layout component
- ✅ Fixed TypeScript compilation issues
  - Corrected icon imports (LuPencil, LuCircleCheck, LuCircleX, LuActivity)
  - Used `import type` syntax for type-only imports
  - Removed unused functions
  - Frontend successfully compiles and builds

#### Documentation
- ✅ Created comprehensive user guide (`docs/DEPARTMENTS_USER_GUIDE.md`)
  - Feature overview and key capabilities
  - Step-by-step instructions for all operations
  - Department hierarchy and user assignment
  - User permissions management
  - Best practices and organizational structure guidelines
  - Common use cases and scenarios
  - Troubleshooting guide
  - Security considerations
- ✅ Created detailed developer guide (`docs/DEPARTMENTS_DEV_GUIDE.md`)
  - Technical architecture overview
  - Database schema documentation
  - Data model specifications
  - Repository method documentation
  - Complete API endpoint reference with examples
  - Frontend implementation details
  - Materialized path pattern explanation
  - Circular reference prevention algorithm
  - Soft delete pattern implementation
  - Performance considerations and optimization
  - Testing guidelines
  - Development tasks and troubleshooting

### Phase 5: LiveKit Webhook Integration (Completed)

#### Backend
- ✅ Created LiveKit webhook data models (`internal/models/livekit_webhook.go`)
  - Room model with codec support and lifecycle tracking
  - Participant model with state management and permissions
  - Track model supporting both audio and video with detailed metadata
  - WebhookEventLog for complete event audit trail
  - WebhookRequest/Response models for API handling
- ✅ Implemented database migrations (`pkg/database/database.go`)
  - `livekit_rooms` table with status tracking and timestamps
  - `livekit_participants` table with join/leave tracking
  - `livekit_tracks` table with publish/unpublish tracking
  - `livekit_webhook_events` table for complete event logging
  - Foreign key relationships ensuring data integrity
  - Comprehensive indexes for fast queries
- ✅ Created LiveKit repository (`pkg/database/livekit_repository.go`)
  - Room operations: Create, GetBySID, FinishRoom, ListRooms
  - Participant operations: Create, UpdateParticipantLeft, GetBySID, ListByRoom
  - Track operations: Create, UnpublishTrack, GetBySID, ListByParticipant, ListByRoom
  - Webhook event logging: LogWebhookEvent, GetWebhookEvents with filters
  - Full CRUD support with proper error handling
- ✅ Implemented webhook handler (`cmd/managing-portal/handlers_livekit.go`)
  - Public webhook endpoint: `POST /webhook/meet` (no auth required)
  - Processes 6 event types:
    - `room_started` - Creates room record
    - `participant_joined` - Tracks participant joining
    - `track_published` - Records audio/video track publication
    - `track_unpublished` - Updates track status
    - `participant_left` - Marks participant as disconnected
    - `room_finished` - Finalizes room session
  - Management API endpoints (authenticated):
    - `GET /api/v1/livekit/rooms` - List all rooms with filters
    - `GET /api/v1/livekit/rooms/{sid}` - Get room details
    - `GET /api/v1/livekit/participants?room_sid=X` - List participants
    - `GET /api/v1/livekit/tracks?room_sid=X` - List tracks
    - `GET /api/v1/livekit/webhook-events` - Query event logs
  - Full Swagger/OpenAPI documentation
  - Comprehensive logging for debugging
- ✅ Integrated with Managing Portal (`cmd/managing-portal/main.go`)
  - Added LiveKitRepository to ManagingPortal struct
  - Registered all webhook and management endpoints
  - Webhook endpoint publicly accessible for LiveKit server
  - Management endpoints protected with JWT authentication

#### Frontend (Managing Portal UI)
- ✅ Created LiveKit API service (`front/managing-portal/src/services/livekit.ts`)
  - TypeScript interfaces for Room, Participant, Track
  - API methods for fetching rooms, participants, tracks, events
  - Automatic token injection for authenticated requests
- ✅ Implemented Rooms list component (`front/managing-portal/src/components/Rooms.tsx`)
  - Grid view of all meeting rooms
  - Status filtering (All, Active, Finished)
  - Auto-refresh every 5 seconds for active rooms
  - Real-time status indicators with pulse animation
  - Click to view room details
  - Responsive card layout
- ✅ Implemented Room Details component (`front/managing-portal/src/components/RoomDetails.tsx`)
  - Complete room information dashboard
  - Statistics cards: participants count, tracks count, duration, timestamps
  - Tabbed interface: Participants tab and Tracks tab
  - Participants view:
    - Avatar with user icon
    - Join/leave timestamps
    - Connection state (Active/Disconnected)
    - Disconnect reason when available
  - Tracks view:
    - Audio/video track cards
    - Track type icons (microphone, camera, screen share)
    - Codec information and resolution for video
    - Simulcast status
    - Publish/unpublish timestamps
  - Back navigation to rooms list
- ✅ Integrated Rooms into navigation
  - Added "Rooms" menu item with video icon in sidebar
  - Added routes in App.tsx: `/rooms` and `/rooms/:sid`
  - Active state highlighting for rooms section
  - Updated page titles in Layout component
- ✅ Implemented Transcription Worker structure (`cmd/transcription-worker/main.go`)
  - Health check endpoint: `GET /health` - Service health status
  - Status endpoint: `GET /status` - Worker statistics and task count
  - Worker loop with goroutine-based task management
  - Placeholder for RabbitMQ message consumption
  - Placeholder for Whisper transcription processing
  - Placeholder for speaker diarization
  - Runs on port 8082
  - Full Swagger documentation
  - Successfully compiles and builds
- ✅ Implemented Summarization Worker structure (`cmd/summarization-worker/main.go`)
  - Health check endpoint: `GET /health` - Service health status
  - Status endpoint: `GET /status` - Worker statistics and task count
  - Worker loop with goroutine-based task management
  - Placeholder for RabbitMQ message consumption
  - Placeholder for LLM summarization processing
  - Runs on port 8083
  - Full Swagger documentation
  - Successfully compiles and builds
- ✅ Created worker task models (`internal/models/media.go`)
  - TranscriptionTask - Task structure for transcription queue
  - SummarizationTask - Task structure for summarization queue
  - Status tracking, timestamps, error handling
- ✅ Updated Transcript model
  - Added Status, ProcessedAt, DurationSecs fields
  - Enhanced for worker integration

### Phase 8: API Documentation & Quality Assurance (Completed)

#### Swagger/OpenAPI Documentation
- ✅ Added Swagger documentation to both portals
  - User Portal: http://localhost:20081/swagger/index.html
  - Managing Portal: http://localhost:20080/swagger/index.html
  - Complete API specification with all 47 endpoints
  - Request/response schemas, authentication details
  - BearerAuth security scheme
- ✅ Installed swaggo dependencies:
  - github.com/swaggo/swag for documentation generation
  - github.com/swaggo/http-swagger for UI serving
  - Automatic docs generation from code annotations
- ✅ Generated comprehensive API specifications:
  - `/cmd/user-portal/docs/` - User Portal OpenAPI spec
  - `/cmd/managing-portal/docs/` - Managing Portal OpenAPI spec
  - JSON and YAML formats available

#### Password Reset Feature (Completed)
- ✅ Implemented complete password reset flow
  - `POST /api/v1/auth/password-reset/request` - Request reset code
  - `POST /api/v1/auth/password-reset/verify` - Verify 6-digit code
  - `POST /api/v1/auth/password-reset/reset` - Set new password
- ✅ Email integration with existing SMTP service
  - HTML email templates (English/Russian)
  - 6-digit verification codes
  - 15-minute code expiration
  - Security warnings in emails
- ✅ Frontend pages (User Portal):
  - ForgotPassword.tsx - Email input and code request
  - ResetPassword.tsx - Code verification and password reset
  - Full i18n support (en/ru translations)
  - Responsive design with validation

#### GORM Migration System (Completed)
- ✅ Replaced SQL migrations with GORM AutoMigrate
  - All 16 tables automatically created/updated
  - Explicit column name mapping for consistency
  - pgvector extension enabled
  - Default data insertion (departments, groups)
- ✅ Created comprehensive GORM models (`pkg/database/models.go`):
  - User, Department, Group, Meeting, MeetingSubject
  - MeetingParticipant, MeetingDepartment
  - LiveKitRoom, LiveKitParticipant, LiveKitTrack
  - LiveKitWebhookEvent, PasswordResetToken
  - Recording, Transcript, RAGDocument, RAGChunk
  - Total: 16 models with proper relationships
- ✅ Converted database operations to GORM:
  - UserRepository.GetByEmail uses GORM queries
  - All new repositories use GORM by default
  - Compatibility wrappers for legacy SQL code
- ✅ Applied to both portals:
  - User Portal uses GORM migrations
  - Managing Portal uses GORM migrations
  - Single source of truth for schema

#### Frontend-Backend Correspondence Audit (Completed)
- ✅ Comprehensive API audit (`API_FRONTEND_BACKEND_AUDIT.md`)
  - Verified all 47 backend endpoints
  - Checked all frontend API calls (19 files total)
  - Confirmed 100% correspondence
  - Documented all API usage patterns
- ✅ Fixed frontend issues:
  - Removed TODO comments from MeetingForm.tsx
  - Confirmed `/api/v1/users` and `/api/v1/departments` work correctly
  - Verified authentication consistency across both portals
- ✅ Statistics:
  - Backend endpoints: 47
  - Frontend usage: ~35 endpoints (74%)
  - Unused endpoints: system/monitoring APIs (expected)
  - All critical user-facing endpoints implemented

### Phase 7: Video Meetings System (Completed)

#### Backend
- ✅ Created Meeting data models (`internal/models/meeting.go`)
  - MeetingType: presentation (with speaker) or conference (all participants equal)
  - MeetingStatus: scheduled, in_progress, completed, cancelled
  - MeetingRecurrence: none, daily, weekly, monthly
  - MeetingSubject: topic categories linked to departments for organization
  - Meeting: core entity with all required fields (title, date/time, duration, recording flags)
  - MeetingParticipant: participants with roles (speaker/participant) and status tracking
  - MeetingDepartment: junction table for department-wide invitations
  - MeetingWithDetails: complete meeting info with all related data loaded
  - Paginated response format: `{items, offset, page_size, total}`
- ✅ Implemented database migrations (`pkg/database/database.go`)
  - `meeting_subjects` table with department_ids array for multi-department linking
  - `meetings` table with all fields including LiveKit room integration
  - `meeting_participants` table with roles and status (invited, accepted, declined, tentative)
  - `meeting_departments` table for bulk department invitations
  - Foreign keys ensuring data integrity with proper cascade/restrict rules
  - Comprehensive indexes for all search filters
- ✅ Created MeetingRepository (`pkg/database/meeting_repository.go`, ~900 lines)
  - Subject CRUD: Create, List (with filters), GetByID, Update, Delete (soft)
  - Meeting CRUD: Create, GetByID, GetWithDetails, Update, Delete
  - ListMeetings with comprehensive filtering:
    - Status, Type, SubjectID filters
    - Date range filtering (date_from, date_to)
    - Speaker filter for presentations
    - UserID filter for "My Meetings" view (participant or speaker)
    - Pagination support
  - Participant management: Add, Get, List by meeting
  - Department management: Add, Get, List by meeting
  - Efficient batch loading to avoid N+1 queries:
    - loadMeetingSubjects, loadMeetingParticipants
    - loadMeetingDepartments, loadMeetingCreators
- ✅ Implemented Meeting Subject API handlers (Managing Portal, `cmd/managing-portal/handlers_meeting_subjects.go`)
  - `POST /api/v1/meeting-subjects` - Create subject (admin only)
  - `GET /api/v1/meeting-subjects` - List subjects with pagination and filters
  - `GET /api/v1/meeting-subjects/{id}` - Get subject details
  - `PUT /api/v1/meeting-subjects/{id}` - Update subject (admin only)
  - `DELETE /api/v1/meeting-subjects/{id}` - Soft delete subject (admin only)
  - Full Swagger documentation
- ✅ Implemented Meeting API handlers (User Portal, `cmd/user-portal/handlers_meetings.go`, ~470 lines)
  - `POST /api/v1/meetings` - Create meeting with permission check
    - Validates can_schedule_meetings permission for non-admin users
    - Supports both individual participants and department-wide invitations
    - Automatically adds speaker for presentation type
    - All participants start with "invited" status
  - `GET /api/v1/meetings` - List "My Meetings" with filters
    - Shows only meetings where user is participant or speaker
    - Supports all search filters: status, type, subject, date range, speaker
    - Paginated response format as requested
  - `GET /api/v1/meetings/{id}` - Get meeting details with access control
    - Only accessible to participants, speakers, or admins
    - Returns full details with nested data
  - `PUT /api/v1/meetings/{id}` - Update meeting (creator or admin only)
    - Can update all meeting fields
    - Can change participants and departments
    - Recalculates participant lists
  - `DELETE /api/v1/meetings/{id}` - Delete meeting (creator or admin only)
    - Hard delete with cascade to participants and departments
  - Full Swagger documentation
  - All CRUD operations compile successfully
- ✅ Integrated with both portals
  - Managing Portal: Added MeetingRepository, registered subject management routes
  - User Portal: Added MeetingRepository, registered meeting CRUD routes
  - Both backends build successfully

#### Frontend (User Portal UI)
- ✅ Created comprehensive TypeScript types (`front/user-portal/src/types/meeting.ts`)
  - All enums: MeetingType, MeetingStatus, MeetingRecurrence, ParticipantRole, ParticipantStatus
  - All models: Meeting, MeetingSubject, MeetingParticipant, MeetingWithDetails
  - All request/response types matching backend exactly
  - Department and User interfaces for dropdowns
- ✅ Created Meeting API service (`front/user-portal/src/services/meetings.ts`)
  - Complete CRUD operations: create, update, delete, get, list
  - Subject operations: list, get
  - Automatic token injection from localStorage/sessionStorage
  - Comprehensive utility functions:
    - formatMeetingDate, formatDuration
    - getMeetingStatusInfo (with color classes)
    - getMeetingTypeInfo (with icons)
    - isMeetingUpcoming, isMeetingNow, isMeetingPast
    - Participant role and status helpers
- ✅ Implemented Meetings list page (`front/user-portal/src/pages/Meetings.tsx`, ~560 lines)
  - **Filter Panel** with all requested search capabilities:
    - Status dropdown (All, Scheduled, In Progress, Completed, Cancelled)
    - Type dropdown (All, Presentation, Conference)
    - Subject dropdown (populated from API)
    - Date range filters (From Date, To Date with datetime-local inputs)
    - Reset filters button
  - **Meetings List View**:
    - Card-based design with meeting cards
    - Each card shows: title, type icon, subject, status badge
    - Date/time, duration, participant count, recording icons
    - Participants preview with avatar circles (first 3 + count)
    - "View Details" button for each meeting
    - Visual indicators for meetings in progress (red pulse)
    - Grayed out past meetings
  - **Meeting Details View**:
    - Complete meeting information grid
    - Type, status, scheduled time, duration, subject, recurrence
    - Video/audio recording flags
    - Additional notes section
    - Participants list with roles (speaker/participant) and status badges
    - Departments list with tags
    - Meeting actions: Join (if in progress), Edit, Delete
  - **Create/Edit Integration**:
    - Floating action button (FAB) for creating meetings
    - Create button in empty state
    - Edit button in details view
    - Delete with confirmation dialog
  - **Pagination**: Previous/Next buttons with page info
  - **Empty states**: Different messages for filtered vs no meetings
  - **Loading states**: Spinner during data fetch
  - **Error handling**: Error message display
  - **Real-time updates**: Refetch after create/edit/delete
- ✅ Implemented Meeting Form component (`front/user-portal/src/components/MeetingForm.tsx`, ~400 lines)
  - **Dual mode**: Create new or Edit existing meeting
  - **Basic Information Section**:
    - Title input with validation
    - Type select (Conference/Presentation with icons)
    - Subject dropdown (loaded from API)
    - Scheduled date/time (datetime-local)
    - Duration in minutes (15-minute increments)
    - Recurrence select (None, Daily, Weekly, Monthly)
  - **Participants Section**:
    - Speaker dropdown (required for presentations, auto-populated in edit)
    - Individual participants checkboxes with full list
    - Departments checkboxes for bulk invitations
    - Participant count display
    - Speaker cannot be selected as regular participant
  - **Recording Options Section**:
    - Video recording checkbox
    - Audio recording checkbox
  - **Additional Notes Section**:
    - Multi-line textarea for agenda/notes
  - **Form Features**:
    - Comprehensive validation (required fields, duration > 0, etc.)
    - Error message display with shake animation
    - Loading states during save
    - Success/Cancel callbacks for integration
    - Pre-population for edit mode
    - API integration for users and departments (with graceful fallback)
  - **Responsive Design**: Mobile-friendly with column stacking
- ✅ Created comprehensive CSS styling
  - `Meetings.css`: Complete styling for list, details, filters, cards
  - `MeetingForm.css`: Form styling with validation states, focus effects
  - Status badge colors for all meeting states
  - Role and participant status badge styling
  - Responsive design with mobile breakpoints
  - Loading spinners and animations
  - Floating action button (FAB) styling
  - Empty state styling

#### Features Delivered
- ✅ **Complete "My Video Meetings" tab** as requested
  - Shows meetings where user is speaker or participant
  - List view with cards showing all key information
  - Details view with complete meeting data
- ✅ **All requested meeting fields**:
  - Title (название) ✓
  - Date and time (дата и время) ✓
  - Recurrence (периодичность) ✓
  - Type: presentation or conference (тип: доклад или совещание) ✓
  - Subject with department links (предмет встречи с привязкой к отделам) ✓
  - Duration (длительность) ✓
  - Participants handling ✓
    - For conferences: all participants ✓
    - For presentations: speaker + participants ✓
    - Add by departments or individually ✓
  - Status: scheduled, in progress, cancelled, etc. ✓
  - Video recording needed (нужна ли видеозапись) ✓
  - Audio recording needed (нужна ли аудиозапись) ✓
  - Additional comments (дополнительные комментарии) ✓
- ✅ **Exact pagination format requested**: `{items: array, offset, pageSize, total}` ✓
- ✅ **All requested search filters**:
  - By date and time (date range) ✓
  - By status ✓
  - By speaker ✓
  - By subject ✓
  - By theme (via subject) ✓

## What Needs To Be Done Next

### Phase 2: Storage Layer Implementation
- [ ] Implement PostgreSQL client (`internal/storage/postgres.go`)
  - Connection pooling
  - Basic CRUD operations
- [ ] Design and implement database schema
  - Users table
  - Recordings table
  - Transcripts table
  - Sessions table
- [ ] Create database migration system
  - Use golang-migrate or similar
  - Initial schema migrations
- [ ] Implement MinIO client (`internal/storage/minio.go`)
  - Upload/download operations
  - Bucket management
  - Presigned URLs
- [ ] Implement Qdrant client (`internal/storage/qdrant.go`)
  - Vector upload
  - Semantic search
  - Collection management

### Phase 3: Message Queue Integration
- [ ] Implement RabbitMQ client (`internal/queue/rabbitmq.go`)
  - Connection management
  - Queue declaration
  - Publisher/subscriber patterns
- [ ] Define queue structure and message formats
  - Transcription queue
  - Summarization queue
  - Processing results queue

### Phase 10: Mobile Application Improvements (Completed)

#### Timezone Support
- ✅ Created date/time utility (`mobile2/lib/utils/date_utils.dart`)
  - `parseToLocal()` - Converts server UTC dates to user's local timezone
  - `formatDateTime()` - Formats dates for display in user's locale
  - `formatDate()`, `formatTime()` - Specialized formatters
  - `formatRelativeDateTime()` - Smart formatting with localized labels (Сегодня/Today, Завтра/Tomorrow, weekday, etc.)
  - `toUtcString()` - Converts local time to UTC for server communication
  - Helper methods: `isToday()`, `isPast()`, `isFuture()`
- ✅ Updated all data models to use timezone-aware parsing
  - `models/meeting.dart` - Meeting, MeetingWithDetails, MeetingParticipant
  - `models/recording.dart` - RoomRecording, TrackRecording
  - All `DateTime.parse()` calls replaced with `AppDateUtils.parseToLocal()`
  - Ensures all dates are displayed in user's local timezone
- ✅ Updated UI screens to display localized times
  - `meetings_screen.dart` - Uses `formatRelativeDateTime()` with localized labels (l10n.today, l10n.tomorrow)
  - `meetings_screen.dart` - Duration displays use localized "min"/"мин" (l10n.minutes)
  - `meeting_detail_screen.dart` - Shows times in user's locale with localized duration units
  - All date/time displays now timezone-aware and fully localized

#### Meeting Creation Improvements
- ✅ Automatic creator participation (`create_meeting_screen.dart`)
  - Current user automatically added as participant on form load
  - Creator ID stored and passed to participant selection dialog
  - Creator cannot be removed from participant list (checkbox disabled)
  - Visual badge "Вы" (You) displayed next to creator's name
  - Ensures meeting creator is always a participant

#### UI/UX Enhancements
- ✅ Simplified join button text (`meetings_screen.dart`)
  - Changed from "Присоединиться к встрече" to "Присоединиться"
  - Shorter, cleaner button label across all meeting types
  - Permanent meetings always show join button
- ✅ Fixed filter pills shadow clipping (`meetings_screen.dart`)
  - Increased bottom padding from 16 to 24 pixels in filter container
  - Shadow from selected filter pills no longer gets cut off
  - Improved visual appearance of top filter slider
- ✅ Screen wake lock during video calls (`video_call_screen.dart`)
  - Added `wakelock_plus: ^1.2.8` dependency
  - Screen stays on during active video calls
  - Automatic wake lock enable on room connection
  - Automatic wake lock disable on disconnect/dispose
  - Prevents screen timeout during meetings
- ✅ Audio source selection (`video_call_screen.dart`)
  - Settings menu now shows available audio input devices
  - Users can switch between microphones during calls
  - Device selection with active device highlighting
  - Live audio source switching without reconnection
- ✅ Local video preview (`video_call_screen.dart`)
  - Shows user's own video immediately when joining alone
  - Replaces "waiting for participants" screen
  - Provides instant feedback that camera is working

#### Backend Fixes
- ✅ Fixed meeting_id population in LiveKit rooms (`handlers_livekit.go`)
  - Placeholder rooms (from participant_joined/track_published) now set meeting_id
  - UUID parsing from room name for meeting linkage
  - Ensures all rooms have meeting_id regardless of event order
- ✅ Fixed database column name (`livekit_repository.go`)
  - Changed UPSERT column from `enabled_codecs_json` to `enabled_codecs`
  - Matches actual database schema defined in GORM model
  - Prevents "column does not exist" errors
- ✅ Enhanced transcription service resilience (`transcription-service/worker.py`)
  - RabbitMQ connection retry logic (5 attempts with 5-second delays)
  - Standby mode when RabbitMQ unavailable
  - Automatic reconnection every 30 seconds
  - Prevents crash loops, provides clear status messages

### Phase 9: Track Recording Transcription System (Completed)

#### Backend (Go)
- ✅ Implemented RabbitMQ Publisher (`pkg/rabbitmq/publisher.go`)
  - Connection management with automatic reconnection
  - Queue declaration and message publishing
  - TranscriptionMessage struct with track_id, user_id, audio_url fields
  - Persistent message delivery mode
  - Added dependency: github.com/rabbitmq/amqp091-go v1.10.0
- ✅ Integrated RabbitMQ into Managing Portal (`cmd/managing-portal/main.go`)
  - Added RabbitMQ publisher to ManagingPortal struct
  - Non-blocking initialization (logs warning if RabbitMQ unavailable)
  - Environment variable configuration (RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASSWORD)
- ✅ Enhanced LiveKit webhook handler (`cmd/managing-portal/handlers_livekit.go`)
  - Modified `egress_ended` event handler to detect track completion
  - Constructs m3u8 playlist URLs from egress file paths
  - Queries database for track and room information
  - Sends transcription tasks to RabbitMQ queue when track recording completes
  - Comprehensive logging for debugging transcription workflow
  - Fixed model type references (models.Track, models.Room)

#### Python Transcription Service
- ✅ Created complete transcription service (`transcription-service/`)
  - `main.py` - Entry point that starts RabbitMQ consumer worker
  - `config.py` - Environment-based configuration for RabbitMQ, PostgreSQL, Whisper, MinIO
  - `worker.py` - RabbitMQ consumer that processes transcription tasks
  - `transcriber.py` - Core transcription logic with Faster Whisper (medium model)
  - `database.py` - PostgreSQL operations for storing transcription results
- ✅ Implemented M3U8 playlist handling (`transcriber.py`)
  - Downloads m3u8 playlists from MinIO storage
  - Parses playlist to extract segment URLs
  - Downloads all audio segments (.ts files)
  - Combines segments using FFmpeg into single audio file
  - Comprehensive logging of download progress
  - Temporary file cleanup after processing
- ✅ Implemented Faster Whisper integration (`transcriber.py`)
  - Loads Faster Whisper medium model on startup
  - Transcribes audio with phrase-level timestamps
  - Returns structured data: phrase_index, start_time, end_time, text
  - Configurable device (CPU/CUDA) and compute type (int8/float16/float32)
  - Model caching for faster subsequent loads
- ✅ Implemented database schema for transcriptions (`database.py`)
  - `transcription_phrases` table: stores individual phrases with timestamps
    - Columns: id, track_id, user_id, phrase_index, start_time, end_time, text, created_at
    - Indexed on track_id for fast retrieval
  - `transcription_status` table: tracks transcription job status
    - Columns: track_id (PK), status, phrase_count, error_message, started_at, completed_at
    - Status values: pending, processing, completed, failed
  - Automatic table creation on service startup
  - Methods: initialize_tables, save_transcription_phrases, update_transcription_status, mark_track_ready
- ✅ Created Dockerfile for transcription service
  - Base image: python:3.11-slim
  - System dependencies: FFmpeg, libpq-dev, gcc
  - Optimized layer caching for faster rebuilds
  - Separate RUN commands to reduce memory usage during build
  - Python dependencies: faster-whisper, pika, psycopg2-binary, python-dotenv, requests
  - Volume mount for Whisper model caching
  - Environment variables for configuration
- ✅ Added transcription service to docker-compose.yml
  - Container name: recontext-transcription-service
  - Image: sivanov2018/recontext-transcription-service:latest
  - Environment configuration:
    - RABBITMQ_HOST=192.168.5.153
    - DB_HOST=postgres
    - WHISPER_MODEL=medium
    - MINIO_SECRET_KEY=32a4953d5bff4a1c6aea4d4ccfb757e5
  - Volume: whisper-models for model caching
  - Network: recontext-network
  - Depends on: postgres, rabbitmq (external services)

#### Documentation
- ✅ Created comprehensive implementation guide (`TRANSCRIPTION_IMPLEMENTATION.md`)
  - Complete workflow from track completion to transcription
  - Backend changes documentation
  - Python service architecture
  - Database schema details
  - Configuration and deployment instructions
  - Testing procedures
- ✅ Created M3U8 handling documentation (`transcription-service/M3U8_HANDLING.md`)
  - Detailed explanation of playlist download process
  - Segment combining with FFmpeg
  - Code examples and logging output
  - Error handling and troubleshooting
- ✅ Created deployment guide (`BUILD_AND_DEPLOY.md`)
  - Quick start commands
  - Build and deployment steps
  - Service verification procedures
  - End-to-end testing workflow
  - Troubleshooting common issues
  - Performance tuning (model selection, GPU acceleration)
  - Maintenance and monitoring

#### Features Delivered
- ✅ **Automatic transcription workflow**:
  1. LiveKit track recording completes → egress_ended webhook
  2. Managing Portal detects track completion
  3. Constructs m3u8 playlist URL from egress file path
  4. Sends transcription message to RabbitMQ queue
  5. Python service consumes message
  6. Downloads and combines m3u8 segments
  7. Transcribes audio with Faster Whisper (medium model)
  8. Saves phrase-level transcriptions to database
  9. Marks track as ready when complete
- ✅ **M3U8 playlist support**: Full HLS stream handling
- ✅ **Phrase-level timestamps**: Each phrase has start and end time
- ✅ **Database persistence**: All transcriptions stored in PostgreSQL
- ✅ **Status tracking**: pending → processing → completed/failed
- ✅ **Error handling**: Comprehensive logging and error messages
- ✅ **Docker deployment**: Complete containerization with docker-compose
- ✅ **External service integration**: Uses external RabbitMQ, PostgreSQL, MinIO on 192.168.5.153

### Phase 11: Race Condition Fixes (Completed)

#### Critical Race Conditions Fixed
- ✅ **CRITICAL: WebSocket Hub Race Condition** (`cmd/user-portal/handlers_websocket.go`)
  - **Issue**: In the broadcast case (lines 97-110), the code held a read lock (`RLock`) while attempting to write to the clients map by calling `close(client.Send)` and `delete(clients, client)` in the default case
  - **Impact**: This violates Go's RWMutex contract - cannot write to a map while holding only a read lock
  - **Fix**: Modified to collect failed clients during read lock, release the lock, then acquire write lock to clean up failed clients
  - **Pattern**: Two-phase approach with Russian comments explaining the fix
  - **Lines changed**: 97-131 (added failedClients collection and separate cleanup phase)
- ✅ **Transcription Worker Map Safety** (`cmd/transcription-worker/main.go`)
  - **Issue**: `tasks` map accessed from multiple goroutines without clear documentation of protection
  - **Fix**: Added comprehensive Russian comments documenting the RWMutex protection for the tasks map
  - **Note**: The `isActive` bool is accessed from multiple goroutines but only written once at startup (documented as safe)
- ✅ **LiveKit Handler Safety Review** (`cmd/managing-portal/handlers_livekit.go`)
  - Reviewed all goroutine closures in webhook handlers
  - All closures correctly capture variables as parameters (lines 652, 844)
  - No shared state race conditions detected - proper closure parameter passing
- ✅ **Transcription Notifier Safety Review** (`cmd/user-portal/transcription_notifier.go`)
  - Reviewed RabbitMQ connection handling and message processing
  - All goroutines properly isolated with no shared map access
  - Channel usage is correct and race-free
- ✅ **Transcription Scheduler Safety Review** (`cmd/user-portal/transcription_scheduler.go`)
  - Ticker-based periodic processing with no shared state
  - All database operations are goroutine-safe (GORM handles concurrency)

#### Build Verification
- ✅ Built managing-portal with race detector (`go build -race`)
  - Command: `go build -race -o /tmp/managing-portal-race ./cmd/managing-portal`
  - Result: **SUCCESS** (minor linker warning about LC_DYSYMTAB is harmless)
- ✅ Built user-portal with race detector (`go build -race`)
  - Command: `go build -race -o /tmp/user-portal-race ./cmd/user-portal`
  - Result: **SUCCESS** (minor linker warning about LC_DYSYMTAB is harmless)

#### Documentation
- ✅ Added Russian comments throughout to explain concurrency safety
  - WebSocket Hub: Detailed explanation of two-phase cleanup
  - Transcription Worker: Documentation of mutex protection
  - Critical sections marked with "ВАЖНО" (IMPORTANT) and "CRITICAL" tags

#### Summary
- **Total race conditions found**: 1 critical, 0 potential
- **Total race conditions fixed**: 1
- **Services verified**: 2 (managing-portal, user-portal)
- **Build status**: All services compile successfully with race detector enabled
- **Next steps**: Run services with `-race` flag in production to catch any runtime race conditions

### Phase 5: Worker Services Full Implementation
- ✅ Complete Transcription Service integration
  - ✅ Implement RabbitMQ message consumption
  - ✅ Integrate Whisper for audio transcription (Faster Whisper medium model)
  - ✅ Add MinIO integration for m3u8 playlist download
  - ✅ Add PostgreSQL integration for transcription storage
  - ✅ Implement phrase-level transcription with timestamps
  - [ ] Implement speaker diarization pipeline (future enhancement)
  - [ ] Implement result publishing to RabbitMQ (future enhancement)
- [ ] Complete Summarization Worker integration
  - [ ] Implement RabbitMQ message consumption
  - [ ] Integrate LLM API (OpenAI/Anthropic/local model)
  - [ ] Implement text chunking for long transcripts
  - [ ] Add PostgreSQL integration for result storage
  - [ ] Implement result publishing to RabbitMQ

### Phase 6: Advanced Features
- [ ] RAG implementation
  - Text chunking and vectorization
  - Integrate with Qdrant for semantic search
  - Context retrieval from vector database
- [ ] WebRTC integration for Jitsi Agent
  - Implement actual WebRTC connection (pion/webrtc)
  - Audio/video stream capture
  - Media file generation (MP4/WebM)
- [ ] External system integrations
  - Asterisk/3CX/Twilio adapters
  - CRM integrations
  - Webhook system

### Immediate Next Steps
1. Implement PostgreSQL storage client and database schema
2. Implement RabbitMQ queue client
3. Implement Transcription Worker with Whisper integration
4. Implement Summarization Worker with LLM integration
5. Connect User Portal file upload to MinIO storage
6. Test end-to-end flow: upload → queue → transcribe → summarize → search

## Notes
- **Phase 1 (Infrastructure) is complete**: Project structure, Docker infrastructure, CI/CD pipeline
- **Phase 2 (Authentication & Core Services) is complete**: JWT auth, User Portal, Jitsi Agent, Swagger docs
- **Phase 3 (User Management & Metrics) is complete**: User CRUD, Groups with dynamic JSON permissions, Metrics collection
- **Phase 4 (Worker Services Structure) is complete**: Transcription and Summarization workers with health checks, basic structure
- **Phase 5 (LiveKit Webhook Integration) is complete**: Full webhook handling, room/participant/track tracking, event logging
- **Phase 6 (Department Management System) is complete**: Hierarchical departments, user assignment, meeting permissions, full UI with tree/list views
- All 5 services now build and run successfully:
  - Managing Portal (Port 8080): admin/admin123 - Admin dashboard with user/group management
  - User Portal (Port 8081): user/user123 - User-facing API for recordings
  - Transcription Worker (Port 8082): health checks, task management structure
  - Summarization Worker (Port 8083): health checks, task management structure
  - Jitsi Agent (Port 8084): service authentication - WebRTC recording agent
- Managing Portal now includes:
  - User CRUD operations (list, get, update, delete, change password)
  - Group management with dynamic JSON permissions
  - Metrics and telemetry collection from all services
  - Centralized logging system
  - Permission checking API for fine-grained access control
  - LiveKit webhook processing with database persistence
  - Room, participant, and track management APIs
  - Full UI for managing LiveKit rooms with real-time updates
  - **NEW**: Department management system with hierarchical organization structure
  - **NEW**: User permissions system (meeting scheduling, department management, recording approval)
  - **NEW**: Video Meetings system with full lifecycle management
- All APIs have complete Swagger/OpenAPI documentation
- Consistent data models across all services (auth, media, jitsi, livekit_webhook, service, groups, metrics, user_crud)
- Docker Compose ready with all 16 services (including observability stack + Watchtower)
- CI/CD builds and pushes 5 custom services to Docker Hub (main branch only)
- Next focus: storage integration (PostgreSQL, MinIO, Qdrant) and queue integration (RabbitMQ)

## How to Run

### Start all services locally:
```bash
docker-compose up -d
```

### Test the APIs:
```bash
# Check health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health  # Transcription Worker
curl http://localhost:8083/health  # Summarization Worker
curl http://localhost:8084/health

# Test LiveKit webhook (simulate event from LiveKit server)
curl -X POST http://localhost:8080/webhook/meet \
  -H "Content-Type: application/json" \
  -d '{
    "event": "room_started",
    "room": {
      "sid": "RM_TEST123",
      "name": "test-room",
      "emptyTimeout": 300,
      "departureTimeout": 20,
      "creationTime": "1762684951",
      "creationTimeMs": "1762684951036"
    },
    "id": "EV_TEST123",
    "createdAt": "1762684951"
  }'

# List LiveKit rooms (requires authentication)
curl http://localhost:8080/api/v1/livekit/rooms \
  -H "Authorization: Bearer $TOKEN"

# Get room participants
curl "http://localhost:8080/api/v1/livekit/participants?room_sid=RM_TEST123" \
  -H "Authorization: Bearer $TOKEN"

# Get webhook event logs
curl "http://localhost:8080/api/v1/livekit/webhook-events?room_sid=RM_TEST123" \
  -H "Authorization: Bearer $TOKEN"

# Login to Managing Portal
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Get system status (requires token)
TOKEN="your-jwt-token-here"
curl http://localhost:8080/api/v1/status \
  -H "Authorization: Bearer $TOKEN"

# Login to User Portal
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"user123"}'

# Upload file (requires token)
curl -X POST http://localhost:8081/api/v1/recordings/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "title=My Recording" \
  -F "file=@recording.mp4"

# User Management (admin only)
# List all users
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Get specific user
curl http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Update user
curl -X PUT http://localhost:8080/api/v1/users/user-123 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"email":"newemail@example.com","role":"user","is_active":true}'

# Group Management (admin only)
# Create a new group with permissions
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Content Editors",
    "description": "Users who can edit content",
    "permissions": {
      "recordings": {
        "actions": ["read", "write"],
        "scope": "all"
      },
      "transcripts": {
        "actions": ["read", "write"],
        "scope": "all"
      }
    }
  }'

# Add user to group
curl -X POST http://localhost:8080/api/v1/groups/add-user \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-123","group_id":"group-001"}'

# Check permission
curl -X POST http://localhost:8080/api/v1/groups/check-permission \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user-123","resource":"recordings","action":"write"}'

# Metrics and Telemetry
# Send metrics from a service
curl -X POST http://localhost:8080/api/v1/metrics \
  -H "Authorization: Bearer $SERVICE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service_id": "user-portal",
    "metrics": [
      {"name": "requests_total", "type": "counter", "value": 100},
      {"name": "request_latency_ms", "type": "gauge", "value": 45.2}
    ]
  }'

# Query metrics
curl "http://localhost:8080/api/v1/metrics?service_id=user-portal&name=requests_total" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

# Get system-wide metrics
curl http://localhost:8080/api/v1/metrics/system \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

### View logs:
```bash
docker-compose logs -f managing-portal
```

### Build and run managing portal standalone:
```bash
go build -o bin/managing-portal ./cmd/managing-portal
./bin/managing-portal
```
