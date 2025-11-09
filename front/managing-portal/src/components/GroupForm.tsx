import React, { useState, useEffect } from 'react';
import './UserForm.css';

interface Group {
  id: string;
  name: string;
  description?: string;
  permissions?: Record<string, any>;
  created_at?: string;
  updated_at?: string;
}

export const GroupForm: React.FC = () => {
  // Extract group ID from URL path
  const pathParts = window.location.pathname.split('/');
  const id = pathParts.includes('edit') ? pathParts[pathParts.length - 2] : null;
  const isEditMode = !!id;

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  useEffect(() => {
    if (isEditMode && id) {
      fetchGroup(id);
    }
  }, [id]);

  const fetchGroup = async (groupId: string) => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/groups/${groupId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch group');
      }

      const group: Group = await response.json();
      setName(group.name);
      setDescription(group.description || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load group');
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');

      if (isEditMode) {
        // Update group
        const response = await fetch(`/api/v1/groups/${id}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            name,
            description,
          }),
        });

        if (!response.ok) {
          throw new Error('Failed to update group');
        }
      } else {
        // Create group
        const response = await fetch('/api/v1/groups', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            name,
            description,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to create group');
        }
      }

      window.location.href = '/groups';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  if (loading && isEditMode) {
    return (
      <div className="user-form-page">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading group...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-form-page">
      <div className="form-header">
        <h1>{isEditMode ? 'Edit Group' : 'Create New Group'}</h1>
        <button onClick={() => window.location.href = '/groups'} className="btn btn-secondary">
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
          <h2>Group Information</h2>

          <div className="form-group">
            <label htmlFor="name">Group Name *</label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder="Enter group name"
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">Description</label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              placeholder="Enter group description (optional)"
            />
          </div>
        </div>

        <div className="form-actions">
          <button type="button" onClick={() => window.location.href = '/groups'} className="btn btn-secondary">
            Cancel
          </button>
          <button type="submit" disabled={loading} className="btn btn-primary">
            {loading ? 'Saving...' : (isEditMode ? 'Update Group' : 'Create Group')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default GroupForm;
