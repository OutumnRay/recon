# Recontext Both Portals - Dual Setup

This setup runs both the Managing Portal and User Portal with their shared dependencies (PostgreSQL and MinIO).

## Services Included

- **Managing Portal** - Administrative interface (port 10080)
- **User Portal** - User interface (port 10081)
- **PostgreSQL** - Shared database (port 5432)
- **MinIO** - Shared object storage (ports 9000, 9001)

## Quick Start

### Using the convenience script (recommended)

```bash
# Start all services
./start-both-portals.sh up

# View logs
./start-both-portals.sh logs

# Stop all services
./start-both-portals.sh down
```

### Using docker-compose directly

**For macOS (Apple Silicon):**
```bash
docker-compose -f docker-compose-both-portals-mac.yml --env-file .env.both-portals up -d
```

**For Linux:**
```bash
docker-compose -f docker-compose-both-portals.yml --env-file .env.both-portals up -d
```

## Configuration

1. Copy the example environment file:
   ```bash
   cp .env.both-portals.example .env.both-portals
   ```

2. Edit `.env.both-portals` to customize:
   - Database credentials
   - MinIO credentials
   - Port mappings for both portals
   - JWT secret

## Access URLs

After starting the services:

- **Managing Portal**: http://localhost:10080
- **User Portal**: http://localhost:10081
- **MinIO Console**: http://localhost:9001
- **PostgreSQL**: localhost:5432

## Default Credentials

**Managing Portal:**
- Username: `admin`
- Password: `admin123`

**User Portal:**
- Username: `user`
- Password: `user123`

**MinIO:**
- Access Key: `minioadmin`
- Secret Key: `minioadmin`

**PostgreSQL:**
- User: `recontext`
- Password: `recontext`
- Database: `recontext`

## Available Commands

The `start-both-portals.sh` script supports the following commands:

- `up` - Start all services (default)
- `down` - Stop all services
- `restart` - Restart all services
- `logs` - View logs (add service name for specific service, e.g., `logs user-portal`)
- `build` - Rebuild Docker images
- `clean` - Stop and remove all data (including volumes)
- `ps` - Show running services

## Troubleshooting

### View logs for a specific service
```bash
./start-both-portals.sh logs managing-portal
./start-both-portals.sh logs user-portal
./start-both-portals.sh logs postgres
./start-both-portals.sh logs minio
```

### Rebuild images
If you've made changes to the code:
```bash
./start-both-portals.sh build
./start-both-portals.sh up
```

### Clean start (removes all data)
```bash
./start-both-portals.sh clean
./start-both-portals.sh up
```

### Port conflicts
If ports are already in use, edit `.env.both-portals`:
```
MANAGING_PORTAL_PORT=10080  # Change to available port
USER_PORTAL_PORT=10081      # Change to available port
POSTGRES_PORT=5432          # Change to available port
MINIO_API_PORT=9000         # Change to available port
MINIO_CONSOLE_PORT=9001     # Change to available port
```

## Data Persistence

Data is persisted in Docker volumes:
- `postgres_data` - Shared database data
- `minio_data` - Shared object storage data

Both portals share the same database and storage, so data is consistent across both interfaces.

To remove all data:
```bash
./start-both-portals.sh clean
```

## Portal Workflow

1. **Managing Portal** (http://localhost:10080)
   - Create and manage user accounts
   - Create and manage groups
   - View system metrics and worker status
   - Monitor recordings and processing

2. **User Portal** (http://localhost:10081)
   - View personal meetings
   - Search through transcripts
   - Access documents and summaries
   - Manage personal settings

## Next Steps

- Access the managing portal at http://localhost:10080 to set up users and groups
- Access the user portal at http://localhost:10081 to test the user experience
- Both portals share the same authentication system
- Data created in one portal is accessible in the other
