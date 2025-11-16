import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import './Organizations.css';

interface Organization {
  id: string;
  name: string;
  description: string;
  domain?: string;
  logo_url?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface OrganizationStats {
  users: number;
  departments: number;
  meetings: number;
}

export const Organizations: React.FC = (): ReactElement => {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [stats, setStats] = useState<Record<string, OrganizationStats>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchOrganizations();
  }, []);

  const fetchOrganizations = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/organizations', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch organizations');
      }

      const data = await response.json();
      setOrganizations(data || []);

      // Fetch stats for each organization
      if (data && data.length > 0) {
        const statsPromises = data.map((org: Organization) => fetchOrganizationStats(org.id));
        const statsResults = await Promise.all(statsPromises);
        const statsMap: Record<string, OrganizationStats> = {};
        data.forEach((org: Organization, index: number) => {
          if (statsResults[index]) {
            statsMap[org.id] = statsResults[index];
          }
        });
        setStats(statsMap);
      }

      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load organizations');
    } finally {
      setLoading(false);
    }
  };

  const fetchOrganizationStats = async (orgId: string): Promise<OrganizationStats | null> => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/organizations/${orgId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        return null;
      }

      return await response.json();
    } catch (err) {
      console.error(`Error fetching stats for organization ${orgId}:`, err);
      return null;
    }
  };

  const handleDeleteOrganization = async (organizationId: string) => {
    if (!confirm('Are you sure you want to delete this organization? This cannot be undone.')) {
      return;
    }

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/organizations/${organizationId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const error = await response.text();
        throw new Error(error || 'Failed to delete organization');
      }

      fetchOrganizations();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete organization');
    }
  };

  if (loading) {
    return (
      <div className="organizations-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading organizations...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="organizations-container">
      <header className="page-header">
        <h1 className="page-title">Organizations Management</h1>
        <div className="header-right">
          <button onClick={() => window.location.href = '/organizations/new'} className="btn btn-primary">
            + Add Organization
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <main className="organizations-main">
        <div className="organizations-table-container">
          <table className="organizations-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Domain</th>
                <th>Users</th>
                <th>Departments</th>
                <th>Meetings</th>
                <th>Status</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {organizations.map((org) => (
                <tr key={org.id}>
                  <td>
                    <span className="organization-name">{org.name}</span>
                  </td>
                  <td>
                    {org.description ? (
                      <span>{org.description}</span>
                    ) : (
                      <span className="text-muted">No description</span>
                    )}
                  </td>
                  <td>
                    {org.domain ? (
                      <span>{org.domain}</span>
                    ) : (
                      <span className="text-muted">-</span>
                    )}
                  </td>
                  <td>
                    <span className="stat-value">{stats[org.id]?.users ?? 0}</span>
                  </td>
                  <td>
                    <span className="stat-value">{stats[org.id]?.departments ?? 0}</span>
                  </td>
                  <td>
                    <span className="stat-value">{stats[org.id]?.meetings ?? 0}</span>
                  </td>
                  <td>
                    <span className={`status-badge ${org.is_active ? 'status-active' : 'status-inactive'}`}>
                      {org.is_active ? 'Active' : 'Inactive'}
                    </span>
                  </td>
                  <td>
                    {new Date(org.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })}
                  </td>
                  <td>
                    <div className="action-buttons">
                      <button
                        onClick={() => window.location.href = `/organizations/${org.id}/edit`}
                        className="btn btn-small btn-secondary"
                      >
                        Edit
                      </button>
                      <button
                        onClick={() => handleDeleteOrganization(org.id)}
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

          {organizations.length === 0 && (
            <div className="empty-state">
              <p>No organizations found</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default Organizations;
