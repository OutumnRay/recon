import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import './Dashboard.css';

interface DashboardStats {
  users: {
    total: number;
    active: number;
  };
  groups: {
    total: number;
  };
  workers: {
    transcription: {
      total: number;
      active: number;
    };
    summarization: {
      total: number;
      active: number;
    };
  };
  storage: {
    used: number;
    total: number;
  };
  recordings: {
    total: number;
    processing: number;
  };
}

interface StatCardProps {
  title: string;
  value: number | string;
  subtitle?: string;
  icon: string;
  trend?: {
    value: number;
    positive: boolean;
  };
}

const StatCard: React.FC<StatCardProps> = ({ title, value, subtitle, icon, trend }): ReactElement => {
  return (
    <div className="stat-card">
      <div className="stat-card-header">
        <div className="stat-icon">{icon}</div>
        <h3 className="stat-title">{title}</h3>
      </div>
      <div className="stat-value">{value}</div>
      {subtitle && <div className="stat-subtitle">{subtitle}</div>}
      {trend && (
        <div className={`stat-trend ${trend.positive ? 'positive' : 'negative'}`}>
          {trend.positive ? '↑' : '↓'} {Math.abs(trend.value)}%
        </div>
      )}
    </div>
  );
};

interface WorkerStatusProps {
  type: string;
  total: number;
  active: number;
}

const WorkerStatus: React.FC<WorkerStatusProps> = ({ type, total, active }): ReactElement => {
  const percentage = total > 0 ? Math.round((active / total) * 100) : 0;
  return (
    <div className="worker-status">
      <div className="worker-header">
        <span className="worker-type">{type}</span>
        <span className="worker-count">{active}/{total}</span>
      </div>
      <div className="worker-progress-bar">
        <div
          className="worker-progress-fill"
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  );
};

export const Dashboard: React.FC = (): ReactElement | null => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  useEffect(() => {
    fetchDashboardStats();
    // Refresh every 30 seconds
    const interval = setInterval(() => fetchDashboardStats(true), 30000);
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardStats = async (isBackgroundRefresh = false) => {
    try {
      if (isBackgroundRefresh) {
        setRefreshing(true);
      }

      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/dashboard/stats', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch dashboard stats');
      }

      const data = await response.json();
      setStats(data);
      setError(null);
      setLastUpdated(new Date());
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard');
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  const handleManualRefresh = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    fetchDashboardStats(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    window.location.href = '/';
  };

  if (loading) {
    return (
      <div className="dashboard-container">
        <div className="dashboard-loading">
          <div className="loading-spinner"></div>
          <p>Loading dashboard...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="dashboard-container">
        <div className="dashboard-error">
          <p>{error}</p>
          <button onClick={() => fetchDashboardStats()} className="btn btn-primary">
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  const storagePercentage = Math.round((stats.storage.used / stats.storage.total) * 100);

  return (
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="header-left">
          <img src="/logo.png" alt="Recontext Logo" className="header-logo" />
          <h1 className="header-title">Managing Portal</h1>
          {lastUpdated && (
            <div className="last-updated">
              Last updated: {lastUpdated.toLocaleTimeString()}
              {refreshing && <span className="refresh-indicator"> Refreshing...</span>}
            </div>
          )}
        </div>
        <div className="header-right">
          <button
            onClick={handleManualRefresh}
            className="btn btn-secondary"
            disabled={refreshing}
          >
            {refreshing ? 'Refreshing...' : 'Refresh'}
          </button>
          <button onClick={handleLogout} className="btn btn-secondary">
            Logout
          </button>
        </div>
      </header>

      <main className="dashboard-main">
        <section className="stats-section">
          <h2 className="section-title">Overview</h2>
          <div className="stats-grid">
            <StatCard
              title="Total Users"
              value={stats.users.total}
              subtitle={`${stats.users.active} active`}
              icon=""
            />
            <StatCard
              title="Groups"
              value={stats.groups.total}
              icon=""
            />
            <StatCard
              title="Recordings"
              value={stats.recordings.total}
              subtitle={`${stats.recordings.processing} processing`}
              icon=""
            />
            <StatCard
              title="Storage"
              value={`${storagePercentage}%`}
              subtitle={`${stats.storage.used} GB / ${stats.storage.total} GB`}
              icon=""
            />
          </div>
        </section>

        <section className="workers-section">
          <h2 className="section-title">Worker Status</h2>
          <div className="workers-container">
            <div className="worker-card">
              <h3 className="worker-card-title">Transcription Workers</h3>
              <WorkerStatus
                type="Active Workers"
                total={stats.workers.transcription.total}
                active={stats.workers.transcription.active}
              />
              <div className="worker-details">
                <div className="worker-detail-item">
                  <span className="detail-label">Total:</span>
                  <span className="detail-value">{stats.workers.transcription.total}</span>
                </div>
                <div className="worker-detail-item">
                  <span className="detail-label">Active:</span>
                  <span className="detail-value success">{stats.workers.transcription.active}</span>
                </div>
                <div className="worker-detail-item">
                  <span className="detail-label">Inactive:</span>
                  <span className="detail-value error">
                    {stats.workers.transcription.total - stats.workers.transcription.active}
                  </span>
                </div>
              </div>
            </div>

            <div className="worker-card">
              <h3 className="worker-card-title">Summarization Workers</h3>
              <WorkerStatus
                type="Active Workers"
                total={stats.workers.summarization.total}
                active={stats.workers.summarization.active}
              />
              <div className="worker-details">
                <div className="worker-detail-item">
                  <span className="detail-label">Total:</span>
                  <span className="detail-value">{stats.workers.summarization.total}</span>
                </div>
                <div className="worker-detail-item">
                  <span className="detail-label">Active:</span>
                  <span className="detail-value success">{stats.workers.summarization.active}</span>
                </div>
                <div className="worker-detail-item">
                  <span className="detail-label">Inactive:</span>
                  <span className="detail-value error">
                    {stats.workers.summarization.total - stats.workers.summarization.active}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="metrics-section">
          <h2 className="section-title">System Metrics</h2>
          <div className="metrics-container">
            <div className="metric-panel">
              <h3 className="metric-panel-title">Processing Queue</h3>
              <div className="metric-panel-content">
                <div className="metric-value-large">{stats.recordings.processing}</div>
                <div className="metric-label">Items in Queue</div>
                <div className="metric-details">
                  <div className="metric-detail-item">
                    <span className="detail-indicator processing"></span>
                    <span>Currently processing</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="metric-panel">
              <h3 className="metric-panel-title">Worker Utilization</h3>
              <div className="metric-panel-content">
                <div className="metric-value-large">
                  {Math.round(((stats.workers.transcription.active + stats.workers.summarization.active) /
                    (stats.workers.transcription.total + stats.workers.summarization.total || 1)) * 100)}%
                </div>
                <div className="metric-label">Active Workers</div>
                <div className="metric-details">
                  <div className="metric-detail-item">
                    <span className="detail-indicator success"></span>
                    <span>{stats.workers.transcription.active + stats.workers.summarization.active} active</span>
                  </div>
                  <div className="metric-detail-item">
                    <span className="detail-indicator error"></span>
                    <span>{(stats.workers.transcription.total - stats.workers.transcription.active) +
                           (stats.workers.summarization.total - stats.workers.summarization.active)} inactive</span>
                  </div>
                </div>
              </div>
            </div>

            <div className="metric-panel">
              <h3 className="metric-panel-title">Storage Health</h3>
              <div className="metric-panel-content">
                <div className="metric-value-large">{storagePercentage}%</div>
                <div className="metric-label">Used Capacity</div>
                <div className="metric-progress">
                  <div className="metric-progress-bar">
                    <div
                      className={`metric-progress-fill ${storagePercentage > 80 ? 'warning' : ''}`}
                      style={{ width: `${storagePercentage}%` }}
                    />
                  </div>
                  <div className="metric-progress-label">{stats.storage.used} GB / {stats.storage.total} GB</div>
                </div>
              </div>
            </div>

            <div className="metric-panel">
              <h3 className="metric-panel-title">Total Recordings</h3>
              <div className="metric-panel-content">
                <div className="metric-value-large">{stats.recordings.total}</div>
                <div className="metric-label">Processed</div>
                <div className="metric-details">
                  <div className="metric-detail-item">
                    <span className="detail-indicator info"></span>
                    <span>All-time total</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default Dashboard;
