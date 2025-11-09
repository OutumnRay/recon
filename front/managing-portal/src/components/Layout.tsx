import React from 'react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { LuLayoutDashboard, LuUsers, LuBuilding2, LuBuilding, LuBookmark, LuVideo, LuLogOut } from 'react-icons/lu';
import { LanguageSwitcher } from './LanguageSwitcher';
import './Layout.css';

interface LayoutProps {
  children: React.ReactNode;
  currentPath: string;
}

export const Layout: React.FC<LayoutProps> = ({ children, currentPath }): ReactElement => {
  const { t } = useTranslation();

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
    // For rooms, also mark as active if on a room details page
    if (path === '/rooms' && currentPath.startsWith('/rooms')) {
      return 'active';
    }
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
            <span className="nav-label">{t('nav.dashboard')}</span>
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
            <span className="nav-label">{t('nav.users')}</span>
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
            <span className="nav-label">{t('nav.groups')}</span>
          </a>

          <a
            href="/departments"
            className={`nav-item ${isActive('/departments')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/departments';
            }}
          >
            <span className="nav-icon"><LuBuilding /></span>
            <span className="nav-label">{t('nav.departments')}</span>
          </a>

          <a
            href="/subjects"
            className={`nav-item ${isActive('/subjects')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/subjects';
            }}
          >
            <span className="nav-icon"><LuBookmark /></span>
            <span className="nav-label">{t('nav.subjects')}</span>
          </a>

          <a
            href="/rooms"
            className={`nav-item ${isActive('/rooms')}`}
            onClick={(e) => {
              e.preventDefault();
              window.location.href = '/rooms';
            }}
          >
            <span className="nav-icon"><LuVideo /></span>
            <span className="nav-label">{t('nav.rooms')}</span>
          </a>
        </nav>

        <div className="sidebar-footer">
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon"><LuLogOut /></span>
            <span className="nav-label">{t('common.logout')}</span>
          </button>
        </div>
      </aside>

      <div className="main-content">
        <header className="top-header">
          <div className="header-left">
            <h2 className="page-current-title">
              {currentPath === '/dashboard' && t('dashboard.title')}
              {currentPath === '/users' && t('users.title')}
              {currentPath === '/groups' && t('groups.title')}
              {currentPath === '/departments' && t('departments.title')}
              {currentPath === '/subjects' && t('subjects.title')}
              {currentPath === '/rooms' && t('rooms.title')}
              {currentPath.startsWith('/rooms/') && t('rooms.viewDetails')}
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
