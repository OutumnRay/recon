# Shared Database Setup for Both Portals

This document explains how users and groups created in the managing portal appear in the user portal through a shared PostgreSQL database.

## Architecture

Both portals (Managing Portal and User Portal) now share the same PostgreSQL database, ensuring that:
- Users created in the managing portal can log into the user portal
- Groups and permissions are synchronized across both portals
- Authentication tokens work on both portals

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    groups TEXT[] DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Groups Table
```sql
CREATE TABLE groups (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

### Group Memberships Table
```sql
CREATE TABLE group_memberships (
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    group_id VARCHAR(255) NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT NOW(),
    added_by VARCHAR(255),
    PRIMARY KEY (user_id, group_id)
);
```

## Default Users

The database is automatically seeded with:

**Admin User** (for Managing Portal):
- Username: `admin`
- Password: `admin123`
- Role: `admin`

**Regular User** (for User Portal):
- Username: `user`
- Password: `user123`
- Role: `user`

## Default Groups

**Editors Group**:
- Permissions: read/write access to recordings, read access to transcripts
- Scope: all

**Viewers Group**:
- Permissions: read-only access to recordings and transcripts
- Scope: all

## How It Works

### 1. Creating Users in Managing Portal

When you create a user via the Managing Portal:

```
POST /api/v1/auth/register
{
  "username": "newuser",
  "email": "newuser@example.com",
  "password": "password123"
}
```

The user is stored in the shared PostgreSQL `users` table.

### 2. User Logs Into User Portal

The newly created user can immediately log into the User Portal:

```
POST /api/v1/auth/login
{
  "username": "newuser",
  "password": "password123"
}
```

Both portals query the same `users` table for authentication.

### 3. Creating Groups in Managing Portal

When you create a group via the Managing Portal:

```
POST /api/v1/groups
{
  "name": "Transcribers",
  "description": "Users who can transcribe recordings",
  "permissions": {
    "recordings": {
      "actions": ["read", "write"],
      "scope": "own"
    }
  }
}
```

The group is stored in the shared `groups` table.

### 4. Adding Users to Groups

When you add a user to a group:

```
POST /api/v1/groups/add-user
{
  "user_id": "user-123",
  "group_id": "group-transcribers"
}
```

This:
1. Adds an entry to `group_memberships` table
2. Updates the user's `groups` array in the `users` table

### 5. Permission Checking

Both portals can check permissions:

```
POST /api/v1/groups/check-permission
{
  "user_id": "user-123",
  "resource": "recordings",
  "action": "write"
}
```

This queries the user's groups and their associated permissions.

## Environment Variables

Both portals need the same database configuration:

```bash
DATABASE_URL=postgresql://recontext:recontext@postgres:5432/recontext?sslmode=disable
```

Or individually:
```bash
DB_HOST=postgres
DB_PORT=5432
DB_USER=recontext
DB_PASSWORD=recontext
DB_NAME=recontext
DB_SSL_MODE=disable
```

## Database Migrations

Migrations are automatically run on portal startup. The database package includes:

1. `createUsersTable` - Creates users table with indexes
2. `createGroupsTable` - Creates groups table with indexes
3. `createGroupMembershipsTable` - Creates group memberships table
4. `insertDefaultData` - Inserts default admin/user and default groups

## API Endpoints

### Managing Portal (Admin Only)

**User Management:**
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/{id}` - Get user by ID
- `POST /api/v1/auth/register` - Create new user
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

**Group Management:**
- `GET /api/v1/groups` - List all groups
- `GET /api/v1/groups/{id}` - Get group by ID
- `POST /api/v1/groups` - Create new group
- `PUT /api/v1/groups/{id}` - Update group
- `DELETE /api/v1/groups/{id}` - Delete group
- `POST /api/v1/groups/add-user` - Add user to group
- `POST /api/v1/groups/check-permission` - Check user permission

### User Portal (Authenticated Users)

**Authentication:**
- `POST /api/v1/auth/login` - Login (uses shared users table)

**User Actions:**
- All authenticated users can access their assigned resources based on group permissions

## Testing the Flow

### Step 1: Start Both Portals
```bash
./start-both-portals.sh up
```

### Step 2: Create a User in Managing Portal
```bash
# Login as admin
curl -X POST http://localhost:10080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# Save the token from response
TOKEN="<token-from-response>"

# Create a new user
curl -X POST http://localhost:10080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "username":"testuser",
    "email":"test@example.com",
    "password":"test123"
  }'
```

### Step 3: Login to User Portal with New User
```bash
curl -X POST http://localhost:10081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"test123"}'
```

You should receive a valid JWT token, proving the user exists in both portals.

### Step 4: Create a Group and Add User
```bash
# Create a group
curl -X POST http://localhost:10080/api/v1/groups \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name":"Contributors",
    "description":"Users who can contribute",
    "permissions":{
      "recordings":{"actions":["read","write"],"scope":"own"}
    }
  }'

# Add user to group
curl -X POST http://localhost:10080/api/v1/groups/add-user \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "user_id":"<user-id-from-step-2>",
    "group_id":"<group-id-from-create>"
  }'
```

## Connection Pooling

The database package uses connection pooling:
- Max Open Connections: 25
- Max Idle Connections: 5
- Connection Max Lifetime: 5 minutes

## Security Considerations

1. **Password Hashing**: All passwords are hashed using bcrypt before storage
2. **JWT Tokens**: Both portals use the same JWT secret for token validation
3. **SQL Injection**: All queries use parameterized statements
4. **Role-Based Access**: Admin endpoints are protected by role middleware

## Troubleshooting

### Users created in managing portal don't appear in user portal

**Check database connection:**
```bash
docker-compose -f docker-compose-both-portals.yml logs postgres
```

**Verify user exists in database:**
```bash
docker exec -it recontext-postgres psql -U recontext -d recontext \
  -c "SELECT id, username, email, role FROM users;"
```

### Permission denied errors

**Check user's groups:**
```bash
docker exec -it recontext-postgres psql -U recontext -d recontext \
  -c "SELECT id, username, groups FROM users WHERE username='testuser';"
```

**Check group permissions:**
```bash
docker exec -it recontext-postgres psql -U recontext -d recontext \
  -c "SELECT id, name, permissions FROM groups;"
```

### Database migration errors

**Reset database:**
```bash
./start-both-portals.sh clean
./start-both-portals.sh up
```

## Implementation Status

✅ Database package created (`pkg/database/`)
✅ User repository with CRUD operations
✅ Group repository with CRUD operations
✅ SQL migrations for tables
✅ Default data seeding

⏳ Integration with managing portal (in progress)
⏳ Integration with user portal (in progress)

## Next Steps

To fully implement this:

1. Update `cmd/managing-portal/main.go` to initialize database connection
2. Update `cmd/user-portal/main.go` to initialize database connection
3. Replace in-memory stores with database repositories
4. Add database URL to environment configurations
5. Test end-to-end user creation flow
6. Add database health checks to both portals

