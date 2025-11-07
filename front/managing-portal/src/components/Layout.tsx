import React from 'react';
import type { ReactElement } from 'react';
import { LuLayoutDashboard, LuUsers, LuBuilding2, LuLogOut } from 'react-icons/lu';
import { LanguageSwitcher } from './LanguageSwitcher';
import './Layout.css';

interface LayoutProps {
  children: React.ReactNode;
  currentPath: string;
}

export const Layout: React.FC<LayoutProps> = ({ children, currentPath }): ReactElement => {
  const getUsername = () => {
    const userStr = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        return user.username || user.email || 'Admin';
      } catch {
        return 'Admin';
      }
    }
    return 'Admin';
  };

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
            <span className="nav-icon"><LuLayoutDashboard /></span>
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
            <span className="nav-icon"><LuUsers /></span>
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
            <span className="nav-icon"><LuBuilding2 /></span>
            <span className="nav-label">Groups</span>
          </a>
        </nav>

        <div className="sidebar-footer">
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon"><LuLogOut /></span>
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
            <div className="language-switcher-wrapper">
              <LanguageSwitcher />
            </div>
            <span className="user-info">{getUsername()}</span>
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
