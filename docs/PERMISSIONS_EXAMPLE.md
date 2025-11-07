# Dynamic JSON Permissions - Usage Examples

## Overview

The Recontext.online platform uses a flexible JSON-based permission system that allows you to define fine-grained access control for user groups.

## Permission Structure

Permissions are defined as JSON objects where:
- **Resource**: The resource type (recordings, transcripts, users, etc.)
- **Actions**: Array of allowed actions (read, write, delete, etc.)
- **Scope**: Access scope (all, own, group)
- **Filters**: Optional additional filters

## Example 1: Editors Group

Users who can read and write recordings, and read transcripts:

```json
{
  "recordings": {
    "actions": ["read", "write"],
    "scope": "all"
  },
  "transcripts": {
    "actions": ["read"],
    "scope": "all"
  }
}
```

## Example 2: Viewers Group

Read-only access to recordings and transcripts:

```json
{
  "recordings": {
    "actions": ["read"],
    "scope": "all"
  },
  "transcripts": {
    "actions": ["read"],
    "scope": "all"
  }
}
```

## Example 3: Team Managers

Can manage their own team's recordings and users:

```json
{
  "recordings": {
    "actions": ["read", "write", "delete"],
    "scope": "group",
    "filters": {
      "team_id": "$user.team_id"
    }
  },
  "users": {
    "actions": ["read", "write"],
    "scope": "group",
    "filters": {
      "team_id": "$user.team_id"
    }
  },
  "transcripts": {
    "actions": ["read", "write"],
    "scope": "group"
  }
}
```

## Example 4: Transcription Operators

Can only work with their own uploads:

```json
{
  "recordings": {
    "actions": ["read", "write"],
    "scope": "own",
    "filters": {
      "user_id": "$user.id"
    }
  },
  "transcripts": {
    "actions": ["read"],
    "scope": "own"
  }
}
```

## Example 5: Advanced - Custom Rules

Complex permission with custom rules:

```json
{
  "recordings": {
    "actions": ["read", "write", "delete"],
    "scope": "all",
    "filters": {
      "status": ["completed", "processing"],
      "created_after": "2024-01-01"
    }
  },
  "search": {
    "actions": ["read"],
    "scope": "all",
    "max_results": 100
  },
  "custom_rules": {
    "can_export": true,
    "can_share": false,
    "max_file_size_mb": 500
  }
}
```

## API Usage

### Create a Group

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Authorization: Bearer $TOKEN" \
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
```

### Add User to Group

```bash
curl -X POST http://localhost:8080/api/v1/groups/add-user \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "group_id": "group-001"
  }'
```

### Check Permission

```bash
curl -X POST http://localhost:8080/api/v1/groups/check-permission \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-123",
    "resource": "recordings",
    "action": "write"
  }'
```

Response:
```json
{
  "allowed": true,
  "reason": "Permission granted via group Content Editors"
}
```

## Permission Scopes

- **all**: Access to all resources
- **own**: Access only to resources owned by the user
- **group**: Access to resources within the user's group/team

## Available Actions

- **read**: View resources
- **write**: Create and update resources
- **delete**: Delete resources
- **export**: Export data
- **share**: Share with others
- **admin**: Administrative actions

## Best Practices

1. **Principle of Least Privilege**: Grant only necessary permissions
2. **Use Groups**: Assign permissions to groups, not individual users
3. **Test Permissions**: Always test new permission sets before deploying
4. **Document Custom Rules**: Clearly document any custom permission rules
5. **Regular Audits**: Periodically review and update permissions
