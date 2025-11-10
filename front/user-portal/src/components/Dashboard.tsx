import React, { useState, useEffect } from 'react';
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuCalendar, LuSearch, LuFileText, LuSettings, LuLogOut, LuMenu, LuX } from 'react-icons/lu';
import { LanguageSwitcher } from './LanguageSwitcher';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [hasFilePermission, setHasFilePermission] = useState(false);
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  useEffect(() => {
    checkFilePermission();
  }, []);

  const checkFilePermission = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/files/permission', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (response.ok) {
        const data = await response.json();
        setHasFilePermission(data.hasPermission || false);
      }
    } catch (err) {
      console.error('Failed to check file permission:', err);
    }
  };

  const getUsername = () => {
    const userStr = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        return user.username || user.email || 'User';
      } catch {
        return 'User';
      }
    }
    return 'User';
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    navigate('/');
  };

  const pageMeta = (() => {
    if (location.pathname.startsWith('/dashboard/documents')) {
      return { title: t('documents.title'), subtitle: t('documents.subtitle') };
    }

    if (location.pathname.startsWith('/dashboard/search')) {
      return { title: t('search.title'), subtitle: t('search.subtitle') };
    }

    if (location.pathname.startsWith('/dashboard/management')) {
      return { title: t('management.title'), subtitle: t('management.subtitle') };
    }

    if (location.pathname.startsWith('/dashboard/meetings')) {
      return { title: t('meetings.title'), subtitle: t('meetings.subtitle') };
    }

    return { title: 'Recontext', subtitle: t('nav.dashboard', { defaultValue: '' }) };
  })();

  return (
    <div className="dashboard-layout">
      <aside className={`sidebar ${isSidebarCollapsed ? 'collapsed' : ''}`}>
        <div className="sidebar-header">
          <img src="/logo.png" alt="Recontext Logo" className="sidebar-logo" />
          <h1 className="sidebar-title">Recontext</h1>
        </div>

        <nav className="sidebar-nav">
          <NavLink
            to="/dashboard/meetings"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon"><LuCalendar /></span>
            <span className="nav-label">{t('nav.meetings')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/search"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon"><LuSearch /></span>
            <span className="nav-label">{t('nav.search')}</span>
          </NavLink>

          {hasFilePermission && (
            <NavLink
              to="/dashboard/documents"
              className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
            >
              <span className="nav-icon"><LuFileText /></span>
              <span className="nav-label">{t('nav.documents')}</span>
            </NavLink>
          )}

          <NavLink
            to="/dashboard/management"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon"><LuSettings /></span>
            <span className="nav-label">{t('nav.management')}</span>
          </NavLink>
        </nav>

        <div className="sidebar-footer">
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon"><LuLogOut /></span>
            <span className="nav-label">{t('common.logout')}</span>
          </button>
        </div>
      </aside>

      <div className={`main-content ${isSidebarCollapsed ? 'expanded' : ''}`}>
        <header className="top-header">
          <div className="header-left">
            <button
              className="sidebar-toggle"
              onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
              title={isSidebarCollapsed ? t('common.expand') : t('common.collapse')}
            >
              {isSidebarCollapsed ? <LuMenu /> : <LuX />}
            </button>
            <div className="page-context">
              <h2 className="page-current-title">{pageMeta.title}</h2>
              {pageMeta.subtitle && (
                <p className="page-current-description">{pageMeta.subtitle}</p>
              )}
            </div>
          </div>
          <div className="header-right">
            <span className="user-info">{getUsername()}</span>
            <div className="header-language-switcher">
              <LanguageSwitcher />
            </div>
          </div>
        </header>
        <div className="content-area">
          <Outlet />
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
