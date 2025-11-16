import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { LuBuilding2, LuUsers, LuFolderTree, LuCalendar } from 'react-icons/lu';
import './Organizations.css';

interface Organization {
  id: string;
  name: string;
  description: string;
  domain?: string;
  logo_url?: string;
  is_active: boolean;
  settings: string;
  created_at: string;
  updated_at: string;
}

interface OrganizationStats {
  users: number;
  departments: number;
  meetings: number;
}

const Organizations: React.FC = () => {
  const { t } = useTranslation();
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<Record<string, OrganizationStats>>({});

  useEffect(() => {
    fetchOrganizations();
  }, []);

  const fetchOrganizations = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/organizations', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
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
    } catch (err) {
      console.error('Error fetching organizations:', err);
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const fetchOrganizationStats = async (orgId: string): Promise<OrganizationStats | null> => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/organizations/${orgId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
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

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="organizations-container">
        <div className="loading-state">
          <div className="spinner"></div>
          <p>{t('common.loading', 'Loading...')}</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="organizations-container">
        <div className="error-state">
          <p className="error-message">{error}</p>
          <button onClick={fetchOrganizations} className="retry-btn">
            {t('common.retry', 'Retry')}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="organizations-container">
      <div className="organizations-header">
        <div className="header-info">
          <h1>{t('organizations.title', 'Organizations')}</h1>
          <p className="subtitle">{t('organizations.subtitle', 'Manage organizations')}</p>
        </div>
      </div>

      {organizations.length === 0 ? (
        <div className="empty-state">
          <LuBuilding2 className="empty-icon" />
          <p>{t('organizations.empty', 'No organizations found')}</p>
        </div>
      ) : (
        <div className="organizations-grid">
          {organizations.map((org) => (
            <div key={org.id} className={`organization-card ${!org.is_active ? 'inactive' : ''}`}>
              <div className="card-header">
                <div className="org-title-section">
                  <LuBuilding2 className="org-icon" />
                  <h3>{org.name}</h3>
                  {!org.is_active && (
                    <span className="inactive-badge">{t('common.inactive', 'Inactive')}</span>
                  )}
                </div>
              </div>

              <div className="card-body">
                {org.description && (
                  <p className="org-description">{org.description}</p>
                )}

                {org.domain && (
                  <p className="org-domain">
                    <strong>{t('organizations.domain', 'Domain')}:</strong> {org.domain}
                  </p>
                )}

                <div className="org-stats">
                  {stats[org.id] && (
                    <>
                      <div className="stat-item">
                        <LuUsers className="stat-icon" />
                        <span>{stats[org.id].users} {t('organizations.users', 'Users')}</span>
                      </div>
                      <div className="stat-item">
                        <LuFolderTree className="stat-icon" />
                        <span>{stats[org.id].departments} {t('organizations.departments', 'Departments')}</span>
                      </div>
                      <div className="stat-item">
                        <LuCalendar className="stat-icon" />
                        <span>{stats[org.id].meetings} {t('organizations.meetings', 'Meetings')}</span>
                      </div>
                    </>
                  )}
                </div>

                <p className="org-created">
                  <strong>{t('organizations.created', 'Created')}:</strong> {formatDate(org.created_at)}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Organizations;
