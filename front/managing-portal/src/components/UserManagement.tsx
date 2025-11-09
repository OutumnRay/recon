import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import './UserManagement.css';

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

export const UserManagement: React.FC = (): ReactElement => {
  const [users, setUsers] = useState<User[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchUsers();
    fetchGroups();
  }, []);

  const fetchUsers = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/users', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch users');
      }

      const data = await response.json();
      setUsers(data.items || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load users');
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


  const handleDeleteUser = async (userId: string) => {
    if (!confirm('Are you sure you want to delete this user?')) {
      return;
    }

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to delete user');
      }

      fetchUsers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete user');
    }
  };

  const handleToggleActive = async (user: User) => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${user.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email: user.email,
          role: user.role,
          groups: user.groups,
          is_active: !user.is_active
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to update user status');
      }

      fetchUsers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update user status');
    }
  };

  const handleToggleFileUploader = async (user: User) => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const isFileUploader = user.groups?.includes('group-file-uploaders') || false;
      const endpoint = isFileUploader ? '/api/v1/groups/remove-user' : '/api/v1/groups/add-user';

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: user.id,
          group_id: 'group-file-uploaders'
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to ${isFileUploader ? 'remove user from' : 'add user to'} File Uploaders group`);
      }

      fetchUsers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update file uploader permission');
    }
  };

  const handleToggleRAGUser = async (user: User) => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const isRAGUser = user.groups?.includes('group-rag-users') || false;
      const endpoint = isRAGUser ? '/api/v1/groups/remove-user' : '/api/v1/groups/add-user';

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          user_id: user.id,
          group_id: 'group-rag-users'
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to ${isRAGUser ? 'remove user from' : 'add user to'} RAG Users group`);
      }

      fetchUsers();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to update RAG user permission');
    }
  };


  if (loading) {
    return (
      <div className="user-management-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading users...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-management-container">
      <header className="page-header">
        <h1 className="page-title">User Management</h1>
        <div className="header-right">
          <button onClick={() => window.location.href = '/users/new'} className="btn btn-primary">
            + Add User
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <main className="users-main">
        <div className="users-table-container">
          <table className="users-table">
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>File Uploader</th>
                <th>RAG User</th>
                <th>Groups</th>
                <th>Status</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td>{user.username}</td>
                  <td>{user.email}</td>
                  <td>
                    <span className={`role-badge role-${user.role}`}>
                      {user.role}
                    </span>
                  </td>
                  <td>
                    <input
                      type="checkbox"
                      checked={user.groups?.includes('group-file-uploaders') || false}
                      onChange={() => handleToggleFileUploader(user)}
                      className="file-uploader-checkbox"
                    />
                  </td>
                  <td>
                    <input
                      type="checkbox"
                      checked={user.groups?.includes('group-rag-users') || false}
                      onChange={() => handleToggleRAGUser(user)}
                      className="rag-user-checkbox"
                    />
                  </td>
                  <td>
                    {user.groups && user.groups.length > 0 ? (
                      <div className="group-tags">
                        {user.groups.map((groupId) => {
                          const group = groups.find(g => g.id === groupId);
                          return (
                            <span key={groupId} className="group-tag">
                              {group?.name || groupId}
                            </span>
                          );
                        })}
                      </div>
                    ) : (
                      <span className="text-muted">No groups</span>
                    )}
                  </td>
                  <td>
                    <button
                      onClick={() => handleToggleActive(user)}
                      className={`status-badge ${user.is_active ? 'active' : 'inactive'}`}
                    >
                      {user.is_active ? 'Active' : 'Inactive'}
                    </button>
                  </td>
                  <td>
                    <div className="action-buttons">
                      <button
                        onClick={() => window.location.href = `/users/${user.id}/edit`}
                        className="btn btn-small btn-secondary"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteUser(user.id)}
                        className="btn btn-small btn-danger"
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {users.length === 0 && (
            <div className="empty-state">
              <p>No users found</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default UserManagement;
