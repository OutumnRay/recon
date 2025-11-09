import React, { useState, useEffect } from 'react';
import { SearchableSelect } from './SearchableSelect';
import './UserForm.css';

interface UserPermissions {
  can_schedule_meetings: boolean;
  can_manage_department: boolean;
  can_approve_recordings: boolean;
}

interface User {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'user' | 'operator';
  department_id?: string | null;
  groups?: string[];
  permissions: UserPermissions;
  is_active?: boolean;
}

interface Group {
  id: string;
  name: string;
  description?: string;
}

interface Department {
  id: string;
  name: string;
  path?: string;
}

export const UserForm: React.FC = () => {
  // Extract user ID from URL path
  const pathParts = window.location.pathname.split('/');
  const id = pathParts.includes('edit') ? pathParts[pathParts.length - 2] : null;
  const isEditMode = !!id;

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'admin' | 'user' | 'operator'>('user');
  const [departmentId, setDepartmentId] = useState('');
  const [selectedGroups, setSelectedGroups] = useState<string[]>([]);
  const [permissions, setPermissions] = useState<UserPermissions>({
    can_schedule_meetings: false,
    can_manage_department: false,
    can_approve_recordings: false,
  });
  const [isActive, setIsActive] = useState(true);

  const [groups, setGroups] = useState<Group[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);

  useEffect(() => {
    fetchGroups();
    fetchDepartments();

    if (isEditMode && id) {
      fetchUser(id);
    }
  }, [id]);

  const fetchUser = async (userId: string) => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user');
      }

      const user: User = await response.json();
      setUsername(user.username);
      setEmail(user.email);
      setRole(user.role);
      setDepartmentId(user.department_id || '');
      setSelectedGroups(user.groups || []);
      setPermissions(user.permissions);
      setIsActive(user.is_active ?? true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load user');
    } finally {
      setLoading(false);
    }
  };

  const fetchGroups = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/groups', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setGroups(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch groups:', err);
    }
  };

  const fetchDepartments = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/departments', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setDepartments(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch departments:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!isEditMode && !password) {
      setError('Password is required for new users');
      return;
    }

    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');

      if (isEditMode) {
        // Update user
        const response = await fetch(`/api/v1/users/${id}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            email,
            role,
            department_id: departmentId || null,
            groups: selectedGroups,
            permissions,
            is_active: isActive,
          }),
        });

        if (!response.ok) {
          throw new Error('Failed to update user');
        }
      } else {
        // Create user
        const registerResponse = await fetch('/api/v1/auth/register', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username,
            email,
            password,
          }),
        });

        if (!registerResponse.ok) {
          const errorData = await registerResponse.json();
          throw new Error(errorData.message || 'Failed to create user');
        }

        // Get the created user to get their ID
        const createdUser = await registerResponse.json();

        // Update user with additional fields (role, department, groups, permissions)
        const updateResponse = await fetch(`/api/v1/users/${createdUser.id}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            role,
            department_id: departmentId || null,
            groups: selectedGroups,
            permissions,
            language: 'en',
            is_active: isActive,
          }),
        });

        if (!updateResponse.ok) {
          throw new Error('User created but failed to update additional fields');
        }
      }

      window.location.href = '/users';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const toggleGroup = (groupId: string) => {
    setSelectedGroups(prev =>
      prev.includes(groupId)
        ? prev.filter(id => id !== groupId)
        : [...prev, groupId]
    );
  };

  if (loading && isEditMode) {
    return (
      <div className="user-form-page">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading user...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-form-page">
      <div className="form-header">
        <h1>{isEditMode ? 'Edit User' : 'Create New User'}</h1>
        <button onClick={() => window.location.href = '/users'} className="btn btn-secondary">
          Cancel
        </button>
      </div>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="user-form">
        <div className="form-section">
          <h2>Account Information</h2>

          <div className="form-group">
            <label htmlFor="username">Username {!isEditMode && '*'}</label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={isEditMode}
              required={!isEditMode}
              placeholder="Enter username"
            />
            {isEditMode && <p className="form-hint">Username cannot be changed</p>}
          </div>

          <div className="form-group">
            <label htmlFor="email">Email *</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="user@example.com"
            />
          </div>

          {!isEditMode && (
            <div className="form-group">
              <label htmlFor="password">Password *</label>
              <input
                type="password"
                id="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                minLength={8}
                placeholder="Minimum 8 characters"
              />
            </div>
          )}

          <div className="form-group">
            <label htmlFor="role">Role *</label>
            <select
              id="role"
              value={role}
              onChange={(e) => setRole(e.target.value as any)}
            >
              <option value="user">User</option>
              <option value="operator">Operator</option>
              <option value="admin">Admin</option>
            </select>
          </div>

          {isEditMode && (
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={isActive}
                  onChange={(e) => setIsActive(e.target.checked)}
                />
                <span>Account is active</span>
              </label>
            </div>
          )}
        </div>

        <div className="form-section">
          <h2>Organization</h2>

          <div className="form-group">
            <SearchableSelect
              id="department"
              label="Department"
              value={departmentId}
              onChange={setDepartmentId}
              options={departments.map(dept => ({
                value: dept.id,
                label: dept.path || dept.name,
              }))}
              placeholder="Select department..."
              emptyPlaceholder="No department"
            />
          </div>

          <div className="form-group">
            <label>Groups</label>
            {groups.length === 0 ? (
              <p className="form-hint">No groups available</p>
            ) : (
              <div className="checkbox-list">
                {groups.map((group) => (
                  <label key={group.id} className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={selectedGroups.includes(group.id)}
                      onChange={() => toggleGroup(group.id)}
                    />
                    <span>
                      {group.name}
                      {group.description && (
                        <small> - {group.description}</small>
                      )}
                    </span>
                  </label>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="form-section">
          <h2>Permissions</h2>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_schedule_meetings}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_schedule_meetings: e.target.checked
                })}
              />
              <span>Can schedule meetings</span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_manage_department}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_manage_department: e.target.checked
                })}
              />
              <span>Can manage department</span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_approve_recordings}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_approve_recordings: e.target.checked
                })}
              />
              <span>Can approve recordings</span>
            </label>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" onClick={() => window.location.href = '/users'} className="btn btn-secondary">
            Cancel
          </button>
          <button type="submit" disabled={loading} className="btn btn-primary">
            {loading ? 'Saving...' : (isEditMode ? 'Update User' : 'Create User')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default UserForm;
