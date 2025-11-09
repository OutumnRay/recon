import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import './Groups.css';

interface Group {
  id: string;
  name: string;
  description?: string;
  permissions?: Record<string, any>;
  created_at?: string;
  updated_at?: string;
}

export const Groups: React.FC = (): ReactElement => {
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchGroups();
  }, []);

  const fetchGroups = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/groups', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch groups');
      }

      const data = await response.json();
      setGroups(data.items || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load groups');
    } finally {
      setLoading(false);
    }
  };


  const handleDeleteGroup = async (groupId: string) => {
    if (!confirm('Are you sure you want to delete this group?')) {
      return;
    }

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/groups/${groupId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to delete group');
      }

      fetchGroups();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete group');
    }
  };


  if (loading) {
    return (
      <div className="groups-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading groups...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="groups-container">
      <header className="page-header">
        <h1 className="page-title">Groups Management</h1>
        <div className="header-right">
          <button onClick={() => window.location.href = '/groups/new'} className="btn btn-primary">
            + Add Group
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <main className="groups-main">
        <div className="groups-table-container">
          <table className="groups-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {groups.map((group) => (
                <tr key={group.id}>
                  <td>
                    <span className="group-name">{group.name}</span>
                  </td>
                  <td>
                    {group.description ? (
                      <span>{group.description}</span>
                    ) : (
                      <span className="text-muted">No description</span>
                    )}
                  </td>
                  <td>
                    {group.created_at ? new Date(group.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' }) : '-'}
                  </td>
                  <td>
                    <div className="action-buttons">
                      <button
                        onClick={() => window.location.href = `/groups/${group.id}/edit`}
                        className="btn btn-small btn-secondary"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteGroup(group.id)}
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

          {groups.length === 0 && (
            <div className="empty-state">
              <p>No groups found</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default Groups;
