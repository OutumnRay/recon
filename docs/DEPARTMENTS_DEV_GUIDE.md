# Department Management Developer Guide

## Overview

This guide provides technical documentation for the Department Management feature in Recontext.online. It covers the backend implementation, database schema, API endpoints, and frontend components.

## Architecture

### Backend Components

```
pkg/database/
├── department_repository.go  # CRUD operations for departments
└── database.go               # Migration definitions

internal/models/
├── department.go             # Department data models
└── auth.go                   # User model with department support

cmd/managing-portal/
├── handlers_department.go    # Department API handlers
└── main.go                   # Route registration
```

### Frontend Components

```
front/managing-portal/src/
├── services/
│   └── departments.ts        # API client for department operations
├── components/
│   ├── Departments.tsx       # Main department management UI
│   └── Departments.css       # Styling
└── App.tsx                   # Route configuration
```

## Database Schema

### departments Table

```sql
CREATE TABLE IF NOT EXISTS departments (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id VARCHAR(255) REFERENCES departments(id) ON DELETE SET NULL,
    level INTEGER NOT NULL DEFAULT 0,
    path TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_departments_parent_id ON departments(parent_id);
CREATE INDEX idx_departments_path ON departments(path);
CREATE INDEX idx_departments_is_active ON departments(is_active);
CREATE INDEX idx_departments_name ON departments(name);
```

### users Table (Department Integration)

```sql
-- Added columns
ALTER TABLE users ADD COLUMN department_id VARCHAR(255)
    REFERENCES departments(id) ON DELETE SET NULL;
ALTER TABLE users ADD COLUMN permissions JSONB NOT NULL
    DEFAULT '{"can_schedule_meetings": false, "can_manage_department": false, "can_approve_recordings": false}';

-- Index
CREATE INDEX idx_users_department_id ON users(department_id);
```

## Data Models

### Department

```go
type Department struct {
    ID          string     `json:"id" db:"id"`
    Name        string     `json:"name" db:"name"`
    Description string     `json:"description" db:"description"`
    ParentID    *string    `json:"parent_id,omitempty" db:"parent_id"`
    Level       int        `json:"level" db:"level"`
    Path        string     `json:"path" db:"path"`
    IsActive    bool       `json:"is_active" db:"is_active"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
```

**Field Descriptions**:
- `ID`: Unique identifier (format: `dept-{uuid}`)
- `Name`: Department name
- `Description`: Optional description
- `ParentID`: Reference to parent department (NULL for root)
- `Level`: Depth in hierarchy (0 for root)
- `Path`: Full hierarchical path (e.g., "Organization/IT/Development")
- `IsActive`: Soft delete flag
- `CreatedAt`, `UpdatedAt`: Timestamps

### DepartmentTreeNode

```go
type DepartmentTreeNode struct {
    Department
    Children []*DepartmentTreeNode `json:"children,omitempty"`
}
```

Used for hierarchical tree representation.

### DepartmentWithStats

```go
type DepartmentWithStats struct {
    Department
    UserCount       int `json:"user_count"`
    ChildCount      int `json:"child_count"`
    TotalUsersCount int `json:"total_users_count"`
}
```

Extended model with statistical information.

### UserPermissions

```go
type UserPermissions struct {
    CanScheduleMeetings bool `json:"can_schedule_meetings"`
    CanManageDepartment bool `json:"can_manage_department"`
    CanApproveRecordings bool `json:"can_approve_recordings"`
}
```

## Repository Methods

### DepartmentRepository

Location: `pkg/database/department_repository.go`

#### Create(dept *models.Department) error

Creates a new department with automatic level and path calculation.

```go
// Automatically calculates:
// - Level: parent.Level + 1 (or 0 for root)
// - Path: parent.Path + "/" + dept.Name
```

**Error Cases**:
- Parent department not found
- Database insert failure

#### GetByID(id string) (*models.Department, error)

Retrieves a department by its ID.

**Error Cases**:
- Department not found (sql.ErrNoRows)
- Database query failure

#### List(parentID *string, includeAll bool) ([]*models.Department, error)

Lists departments with optional filters.

**Parameters**:
- `parentID`: Filter by parent (nil for all)
- `includeAll`: Include inactive departments

**Returns**: Departments ordered by path (ASC)

#### GetTree(rootID *string) (*models.DepartmentTreeNode, error)

Builds hierarchical tree structure.

**Parameters**:
- `rootID`: Root department ID (nil for organization root)

**Algorithm**:
1. Fetch all active departments
2. Create map of department nodes
3. Link children to parents
4. Return root node

#### Update(dept *models.Department) error

Updates department with automatic path recalculation for children.

**Features**:
- Circular reference prevention
- Automatic level and path recalculation
- Recursive child path updates

**Error Cases**:
- Parent not found
- Circular reference detected
- Department not found
- Database update failure

#### updateChildrenPaths(parentID, parentPath string) error

Recursively updates paths for all child departments.

**Called automatically by Update()**

#### Delete(id string) error

Soft deletes a department (sets is_active = false).

**Important**: Cannot delete if users are assigned. Check with `GetWithStats` first.

#### GetWithStats(id string) (*models.DepartmentWithStats, error)

Retrieves department with statistics.

**Statistics Calculated**:
- `UserCount`: Direct users in this department
- `ChildCount`: Immediate child departments
- `TotalUsersCount`: Users in this department and all sub-departments

#### GetChildren(parentID string) ([]*models.Department, error)

Retrieves all direct child departments.

**Returns**: Active children ordered by name (ASC)

#### NameExists(name string, parentID *string, excludeID string) (bool, error)

Checks if department name exists at the same level.

**Parameters**:
- `name`: Department name to check
- `parentID`: Parent department (ensures uniqueness within level)
- `excludeID`: Exclude this ID from check (for updates)

## API Endpoints

### Authentication

All endpoints require admin authentication (`Bearer <token>`)

### POST /api/v1/departments

Create a new department.

**Request Body**:
```json
{
  "name": "Development Team",
  "description": "Software development department",
  "parent_id": "dept-root"
}
```

**Response** (201 Created):
```json
{
  "id": "dept-abc123",
  "name": "Development Team",
  "description": "Software development department",
  "parent_id": "dept-root",
  "level": 1,
  "path": "Organization/Development Team",
  "is_active": true,
  "created_at": "2025-01-09T10:00:00Z",
  "updated_at": "2025-01-09T10:00:00Z"
}
```

**Error Responses**:
- 400: Invalid request body or validation failed
- 409: Department name already exists at this level

### GET /api/v1/departments

List departments with optional filters.

**Query Parameters**:
- `parent_id` (optional): Filter by parent department
- `include_all` (optional): Include inactive departments (default: false)
- `tree` (optional): Return hierarchical tree structure (default: false)

**Response** (200 OK) - Flat List:
```json
[
  {
    "id": "dept-root",
    "name": "Organization",
    "level": 0,
    "path": "Organization",
    ...
  },
  {
    "id": "dept-abc123",
    "name": "Development Team",
    "level": 1,
    "path": "Organization/Development Team",
    ...
  }
]
```

**Response** (200 OK) - Tree (when `tree=true`):
```json
{
  "id": "dept-root",
  "name": "Organization",
  "level": 0,
  "children": [
    {
      "id": "dept-abc123",
      "name": "Development Team",
      "level": 1,
      "children": []
    }
  ]
}
```

### GET /api/v1/departments/{id}

Get department by ID.

**Query Parameters**:
- `stats` (optional): Include statistics (default: false)

**Response** (200 OK) - Without Stats:
```json
{
  "id": "dept-abc123",
  "name": "Development Team",
  ...
}
```

**Response** (200 OK) - With Stats (`stats=true`):
```json
{
  "id": "dept-abc123",
  "name": "Development Team",
  "user_count": 15,
  "child_count": 3,
  "total_users_count": 45,
  ...
}
```

### PUT /api/v1/departments/{id}

Update department.

**Request Body** (all fields optional):
```json
{
  "name": "Updated Name",
  "description": "Updated description",
  "parent_id": "new-parent-id",
  "is_active": true
}
```

**Response** (200 OK):
```json
{
  "id": "dept-abc123",
  "name": "Updated Name",
  ...
}
```

**Error Responses**:
- 400: Circular reference detected
- 404: Department not found
- 409: Name already exists at this level

### DELETE /api/v1/departments/{id}

Soft delete department.

**Response** (200 OK):
```json
{
  "message": "Department deleted successfully",
  "department_id": "dept-abc123"
}
```

**Error Responses**:
- 400: Department has active users (cannot delete)
- 404: Department not found

### GET /api/v1/departments/{id}/children

Get child departments.

**Response** (200 OK):
```json
[
  {
    "id": "dept-child1",
    "name": "Frontend Team",
    "parent_id": "dept-abc123",
    ...
  },
  {
    "id": "dept-child2",
    "name": "Backend Team",
    "parent_id": "dept-abc123",
    ...
  }
]
```

## Frontend Implementation

### Department API Service

Location: `front/managing-portal/src/services/departments.ts`

```typescript
export const departmentsApi = {
  getDepartments(parentId?: string, includeAll?: boolean): Promise<Department[]>
  getDepartmentTree(rootId?: string): Promise<DepartmentTreeNode>
  getDepartment(id: string, includeStats?: boolean): Promise<Department | DepartmentWithStats>
  createDepartment(data: CreateDepartmentRequest): Promise<Department>
  updateDepartment(id: string, data: UpdateDepartmentRequest): Promise<Department>
  deleteDepartment(id: string): Promise<void>
  getChildren(id: string): Promise<Department[]>
}
```

### Departments Component

Location: `front/managing-portal/src/components/Departments.tsx`

**Features**:
- Tree view with expand/collapse functionality
- List view with card layout
- Create/Edit modal form
- Department details panel with statistics
- Real-time statistics display

**State Management**:
```typescript
const [departments, setDepartments] = useState<Department[]>([])
const [departmentTree, setDepartmentTree] = useState<DepartmentTreeNode | null>(null)
const [viewMode, setViewMode] = useState<'list' | 'tree'>('tree')
const [selectedDepartment, setSelectedDepartment] = useState<DepartmentWithStats | null>(null)
```

## Key Implementation Details

### Materialized Path Pattern

The department hierarchy uses the **materialized path** pattern:

- **Path Storage**: Full hierarchical path stored as string
- **Example**: `"Organization/IT/Development"`
- **Benefits**:
  - Fast ancestor/descendant queries
  - Efficient tree traversal
  - Simple hierarchy visualization

**Path Maintenance**:
- Calculated on creation based on parent's path
- Recalculated recursively when parent changes
- Format: `parent.Path + "/" + dept.Name`

### Circular Reference Prevention

Prevents creating circular hierarchies:

```go
// In Update method
if strings.HasPrefix(parent.Path, dept.Path+"/") {
    return fmt.Errorf("cannot set child department as parent (circular reference)")
}
```

**Example**:
- Cannot make "IT" a child of "Development" if "Development" is already under "IT"

### Soft Delete Pattern

Departments use soft deletion:

- Sets `is_active = false` instead of removing rows
- Preserves historical data and relationships
- Can be filtered out in queries
- Can be reactivated if needed

## Testing

### Backend Tests

```bash
# Run all tests
go test ./pkg/database/...

# Test specific repository
go test -v ./pkg/database -run TestDepartmentRepository
```

### Frontend Tests

```bash
# Build frontend
cd front/managing-portal
npm run build

# Run tests (if configured)
npm test
```

## Common Development Tasks

### Adding New Department Fields

1. Update model in `internal/models/department.go`
2. Create migration in `pkg/database/database.go`
3. Update repository methods in `pkg/database/department_repository.go`
4. Update API handlers in `cmd/managing-portal/handlers_department.go`
5. Update TypeScript interfaces in `front/managing-portal/src/services/departments.ts`
6. Update frontend components

### Adding New Permissions

1. Update `UserPermissions` struct in `internal/models/auth.go`
2. Update migration default JSON in `pkg/database/database.go`
3. Update frontend user edit forms

### Custom Department Queries

Example: Get all departments under "IT":

```go
query := `
    SELECT * FROM departments
    WHERE path LIKE $1 AND is_active = true
    ORDER BY level, name
`
pathPattern := "IT%"
rows, err := db.Query(query, pathPattern)
```

## Performance Considerations

### Database Indexes

Critical indexes for performance:
- `idx_departments_parent_id`: Fast child lookups
- `idx_departments_path`: Efficient hierarchy queries
- `idx_departments_is_active`: Quick active department filtering

### Query Optimization

- Tree queries fetch all departments once, build tree in memory
- Statistics use JOIN and GROUP BY for efficient counting
- Path-based queries use LIKE with indexes

### Frontend Optimization

- Tree view: Expand nodes on demand to reduce initial render
- List view: Virtual scrolling for large department lists (future enhancement)
- Debounce search/filter inputs
- Cache department tree structure

## Security Considerations

### Authorization

- All department endpoints require admin role
- JWT token validation on every request
- User permissions stored in encrypted database

### Data Validation

- Name uniqueness validated at database level
- Circular reference prevention in application logic
- Parent ID validation before operations
- SQL injection prevention via parameterized queries

### Audit Trail

- All department changes logged with timestamps
- User assignment changes tracked
- Permission changes recorded

## Troubleshooting

### Migration Issues

```bash
# Check current migration state
psql -d recontext -c "SELECT * FROM schema_migrations"

# Reset and rerun migrations (DEV ONLY)
psql -d recontext -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public"
# Restart application to run migrations
```

### Debug Department Paths

```sql
-- Check department hierarchy
SELECT id, name, level, path, parent_id
FROM departments
ORDER BY path;

-- Find broken hierarchies
SELECT d1.id, d1.name, d1.parent_id
FROM departments d1
LEFT JOIN departments d2 ON d1.parent_id = d2.id
WHERE d1.parent_id IS NOT NULL AND d2.id IS NULL;
```

### Frontend Debugging

```javascript
// Enable API logging
localStorage.setItem('DEBUG_API', 'true')

// Check department state in console
console.log(departmentsApi.getDepartmentTree())
```

## Future Enhancements

- [ ] Department-level permission inheritance
- [ ] Bulk user assignment to departments
- [ ] Department templates for quick setup
- [ ] Export department structure to CSV/JSON
- [ ] Department activity dashboard
- [ ] Role-based department management (non-admin managers)
- [ ] Department-specific settings and configurations

## References

- [User Guide](./DEPARTMENTS_USER_GUIDE.md)
- [API Documentation](./DEPARTMENTS_API_GUIDE.md)
- [Database Schema](../pkg/database/database.go)
- [Repository Implementation](../pkg/database/department_repository.go)
