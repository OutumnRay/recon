# Department Management User Guide

## Overview

The Department Management feature allows administrators to organize users into a hierarchical structure of departments and sub-departments. This enables better organization, permission management, and resource allocation within the Recontext platform.

## Key Features

- **Hierarchical Structure**: Create departments with unlimited levels of nesting
- **User Assignment**: Assign users to specific departments
- **Permission Management**: Control user permissions including meeting scheduling rights
- **Statistics Tracking**: View user counts and sub-department counts for each department
- **Tree and List Views**: View departments in a hierarchical tree or flat list format

## Accessing Department Management

1. Log in to the Recontext Management Portal
2. Navigate to **Departments** from the sidebar menu (building icon)

## Managing Departments

### Creating a Department

1. Click the **Create Department** button in the top-right corner
2. Fill in the department details:
   - **Name** (required): The department name
   - **Description**: Optional description of the department
   - **Parent Department**: Select a parent department or leave empty for root-level department
3. Click **Create** to save the department

### Viewing Departments

#### Tree View (Default)
- Displays departments in a hierarchical tree structure
- Click the arrow icons to expand/collapse sub-departments
- Click on a department name to view its details in the right panel

#### List View
- Displays all departments in a grid of cards
- Shows department name, level, path, and description
- Click on any card to view detailed statistics

### Editing a Department

1. Locate the department you want to edit (in either tree or list view)
2. Click the **Edit** (pencil icon) button
3. Update the department information
4. Click **Update** to save changes

**Note**: Changing a department's parent will automatically recalculate paths for all child departments.

### Deleting a Department

1. Locate the department you want to delete
2. Click the **Delete** (trash icon) button
3. Confirm the deletion in the popup dialog

**Important**: You cannot delete a department that has active users assigned to it. Reassign users to other departments first.

### Department Details Panel

When you select a department, the details panel on the right shows:

- **Description**: Full department description
- **Hierarchy Information**:
  - Level: Depth in the organization hierarchy (0 for root)
  - Path: Full hierarchical path (e.g., "Organization/IT/Development")
- **Statistics**:
  - Direct Users: Number of users directly assigned to this department
  - Child Departments: Number of immediate sub-departments
  - Total Users: Total users including all sub-departments
- **Status**: Active or Inactive

## Managing User Department Assignments

### Assigning Users to Departments

1. Navigate to **Users** from the sidebar menu
2. Click on a user to edit their profile
3. Select a department from the **Department** dropdown
4. Click **Save** to update the user's assignment

### User Permissions

When assigning or editing users, you can grant the following permissions:

- **Can Schedule Meetings**: Allows the user to create and schedule video meetings
- **Can Manage Department**: Allows the user to manage their department's settings
- **Can Approve Recordings**: Allows the user to approve meeting recordings

## Best Practices

### Organizational Structure

1. **Start with a Root Department**: Create a root-level department representing your organization
2. **Logical Grouping**: Organize departments by function, location, or team structure
3. **Limit Depth**: While unlimited nesting is supported, try to keep hierarchy depth reasonable (3-5 levels max)
4. **Consistent Naming**: Use clear, consistent naming conventions across all departments

### User Management

1. **Assign Users Promptly**: Assign new users to departments as soon as they're created
2. **Review Regularly**: Periodically review user assignments and update as needed
3. **Permission Control**: Grant meeting scheduling permissions based on user roles and needs
4. **Department Managers**: Designate users with "Can Manage Department" permission for each department

### Maintenance

1. **Avoid Empty Departments**: Remove or merge departments with no users
2. **Update Descriptions**: Keep department descriptions current and meaningful
3. **Monitor Statistics**: Use the statistics panel to understand department size and growth
4. **Plan Changes**: Before restructuring, plan moves carefully to avoid disruption

## Common Use Cases

### Scenario 1: Creating a New Team

1. Navigate to Departments
2. Click **Create Department**
3. Set the parent department (e.g., "Engineering")
4. Enter team name (e.g., "Frontend Team")
5. Add description
6. Assign team members via the Users page

### Scenario 2: Reorganizing Departments

1. Identify departments to be moved
2. Edit each department
3. Change the parent department
4. System automatically updates all paths and child relationships

### Scenario 3: Granting Meeting Permissions

1. Navigate to Users page
2. Select user to edit
3. Check "Can Schedule Meetings" permission
4. Save changes
5. User can now create video meetings from their account

## Troubleshooting

### Cannot Delete Department

**Problem**: Error message when trying to delete a department

**Solution**:
- Check if the department has active users assigned
- Reassign users to other departments first
- Try deleting again

### Circular Reference Error

**Problem**: Error when changing parent department

**Solution**:
- You cannot make a department its own ancestor
- Choose a different parent department
- Ensure the parent is not a child of the current department

### User Not Seeing Meeting Option

**Problem**: User cannot schedule meetings

**Solution**:
- Verify user has "Can Schedule Meetings" permission enabled
- Check if user is assigned to a department
- Ensure user role permits meeting scheduling

## API Integration

For developers integrating with the Department API, refer to the [Department API Documentation](./DEPARTMENTS_API_GUIDE.md) for detailed endpoint information.

## Security Considerations

- Only administrators can create, edit, or delete departments
- User department assignments are logged for audit purposes
- Permission changes are tracked in the system logs
- Department hierarchy prevents circular references automatically

## Support

For additional help or to report issues:
- Submit issues to the project repository
- Contact your system administrator
- Refer to the [Technical Documentation](./DEPARTMENTS_DEV_GUIDE.md) for implementation details
