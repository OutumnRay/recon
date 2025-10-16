# Recontext Managing Portal - Standalone Setup

This setup runs only the Managing Portal with its minimal dependencies (PostgreSQL and MinIO).

## Services Included

- **Managing Portal** - Administrative interface (port 10080)
- **PostgreSQL** - Database (port 5432)
- **MinIO** - Object storage (ports 9000, 9001)

## Quick Start

### Using the convenience script (recommended)

```bash
# Start all services
./start-managing-portal.sh up

# View logs
./start-managing-portal.sh logs

# Stop all services
./start-managing-portal.sh down
```

### Using docker-compose directly

**For macOS (Apple Silicon):**
```bash
docker-compose -f docker-compose-managing-portal-mac.yml --env-file .env.managing-portal up -d
```

**For Linux:**
```bash
docker-compose -f docker-compose-managing-portal.yml --env-file .env.managing-portal up -d
```

## Configuration

1. Copy the example environment file:
   ```bash
   cp .env.managing-portal.example .env.managing-portal
   ```

2. Edit `.env.managing-portal` to customize:
   - Database credentials
   - MinIO credentials
   - Port mappings
   - JWT secret

## Access URLs

After starting the services:

- **Managing Portal**: http://localhost:10080
- **MinIO Console**: http://localhost:9001
- **PostgreSQL**: localhost:5432

## Default Credentials

**Managing Portal:**
- Username: `admin`
- Password: `admin123`

**MinIO:**
- Access Key: `minioadmin`
- Secret Key: `minioadmin`

**PostgreSQL:**
- User: `recontext`
- Password: `recontext`
- Database: `recontext`

## Available Commands

The `start-managing-portal.sh` script supports the following commands:

- `up` - Start all services (default)
- `down` - Stop all services
- `restart` - Restart all services
- `logs` - View logs (add service name for specific service, e.g., `logs managing-portal`)
- `build` - Rebuild Docker images
- `clean` - Stop and remove all data (including volumes)
- `ps` - Show running services

## Troubleshooting

### View logs for a specific service
```bash
./start-managing-portal.sh logs managing-portal
./start-managing-portal.sh logs postgres
./start-managing-portal.sh logs minio
```

### Rebuild images
If you've made changes to the code:
```bash
./start-managing-portal.sh build
./start-managing-portal.sh up
```

### Clean start (removes all data)
```bash
./start-managing-portal.sh clean
./start-managing-portal.sh up
```

### Port conflicts
If ports are already in use, edit `.env.managing-portal`:
```
MANAGING_PORTAL_PORT=10080  # Change to available port
POSTGRES_PORT=5432          # Change to available port
MINIO_API_PORT=9000         # Change to available port
MINIO_CONSOLE_PORT=9001     # Change to available port
```

## Data Persistence

Data is persisted in Docker volumes:
- `postgres_data` - Database data
- `minio_data` - Object storage data

To remove all data:
```bash
./start-managing-portal.sh clean
```

## Next Steps

- Access the managing portal at http://localhost:10080
- Log in with default credentials
- Create users and groups
- Configure system settings
