import React from 'react';
import type { ReactElement } from 'react';
import './Layout.css';

interface LayoutProps {
  children: React.ReactNode;
  currentPath: string;
}

export const Layout: React.FC<LayoutProps> = ({ children, currentPath }): ReactElement => {
  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    window.location.href = '/';
  };

  const isActive = (path: string) => {
    return currentPath === path ? 'active' : '';
  };

  return (
    <div className="layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <img src="/logo.png" alt="Recontext Logo" className="sidebar-logo" />
          <h1 className="sidebar-title">Recontext</h1>
        </div>

        <nav className="sidebar-nav">
          <a
            href="/dashboard"
            className={`nav-item ${isActive('/dashboard')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/dashboard';
            }}
          >
            <span className="nav-icon">📊</span>
            <span className="nav-label">Dashboard</span>
          </a>

          <a
            href="/users"
            className={`nav-item ${isActive('/users')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/users';
            }}
          >
            <span className="nav-icon">👥</span>
            <span className="nav-label">Users</span>
          </a>

          <a
            href="/groups"
            className={`nav-item ${isActive('/groups')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/groups';
            }}
          >
            <span className="nav-icon">🏢</span>
            <span className="nav-label">Groups</span>
          </a>
        </nav>

        <div className="sidebar-footer">
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon">🚪</span>
            <span className="nav-label">Logout</span>
          </button>
        </div>
      </aside>

      <div className="main-content">
        <header className="top-header">
          <div className="header-left">
            <h2 className="page-current-title">
              {currentPath === '/dashboard' && 'Dashboard'}
              {currentPath === '/users' && 'User Management'}
              {currentPath === '/groups' && 'Groups Management'}
            </h2>
          </div>
          <div className="header-right">
            <span className="user-info">Admin</span>
          </div>
        </header>
        <div className="content-area">
          {children}
        </div>
      </div>
    </div>
  );
};

export default Layout;
