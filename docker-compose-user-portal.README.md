# User Portal Development Environment

This minimal docker-compose setup includes only the essential services needed to run and test the user portal:

- **User Portal** - The main user-facing web application
- **PostgreSQL** - Database for user authentication and data
- **MinIO** - S3-compatible object storage for file uploads

## Quick Start

### For Linux:

```bash
docker-compose -f docker-compose-user-portal.yml up -d
```

### For macOS:

```bash
docker-compose -f docker-compose-user-portal-mac.yml up -d
```

## Access Points

- **User Portal**: http://localhost:10081
- **MinIO Console**: http://localhost:9001
  - Username: `minioadmin`
  - Password: `minioadmin`
- **PostgreSQL**: localhost:5432
  - Database: `recontext`
  - Username: `recontext`
  - Password: `recontext`

## Default Login Credentials

- Username: `user`
- Password: `user123`

## Build and Start

```bash
# Build and start all services
docker-compose -f docker-compose-user-portal.yml up --build -d

# View logs
docker-compose -f docker-compose-user-portal.yml logs -f user-portal

# Stop all services
docker-compose -f docker-compose-user-portal.yml down

# Stop and remove volumes (clean start)
docker-compose -f docker-compose-user-portal.yml down -v
```

## Services Included

### 1. PostgreSQL (port 5432)
- Stores user accounts and session data
- Health check enabled
- Data persists in `postgres-data-dev` volume

### 2. MinIO (ports 9000, 9001)
- S3-compatible object storage
- Console UI on port 9001
- API on port 9000
- Data persists in `minio-data-dev` volume

### 3. User Portal (port 10081)
- Main web application
- Connects to PostgreSQL and MinIO
- Hot reload enabled in development mode

## Environment Variables

The user portal uses these environment variables:

```env
PORT=8081
DB_HOST=postgres
DB_PORT=5432
DB_USER=recontext
DB_PASSWORD=recontext
DB_NAME=recontext
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_USE_SSL=false
JWT_SECRET=your-secret-key-change-in-production
LOG_LEVEL=debug
```

## Development Workflow

1. Start the services:
   ```bash
   docker-compose -f docker-compose-user-portal-mac.yml up -d
   ```

2. Wait for health checks to pass (about 30 seconds)

3. Access the portal at http://localhost:10081

4. Make changes to the code and rebuild:
   ```bash
   docker-compose -f docker-compose-user-portal-mac.yml up --build -d user-portal
   ```

5. View logs:
   ```bash
   docker-compose -f docker-compose-user-portal-mac.yml logs -f user-portal
   ```

## Troubleshooting

### Port already in use
If you get a "port already allocated" error, another service might be using the ports. You can either:
- Stop the conflicting service
- Change the ports in the docker-compose file

### MinIO not accessible
Make sure the MinIO service is healthy:
```bash
docker-compose -f docker-compose-user-portal-mac.yml ps
```

### Database connection errors
Check PostgreSQL health:
```bash
docker-compose -f docker-compose-user-portal-mac.yml logs postgres
```

### Clean restart
To start fresh with no data:
```bash
docker-compose -f docker-compose-user-portal-mac.yml down -v
docker-compose -f docker-compose-user-portal-mac.yml up -d
```

## Network

All services run on the `recontext-dev` bridge network, allowing them to communicate using service names as hostnames.

## Volumes

- `postgres-data-dev` - PostgreSQL data directory
- `minio-data-dev` - MinIO storage directory

Data persists across container restarts but can be removed with `docker-compose down -v`.
