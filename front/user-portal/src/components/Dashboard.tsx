import React from 'react';
import { NavLink, Outlet, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LanguageSwitcher } from './LanguageSwitcher';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
    navigate('/');
  };

  return (
    <div className="dashboard-layout">
      <aside className="sidebar">
        <div className="sidebar-header">
          <img src="/logo.png" alt="Recontext Logo" className="sidebar-logo" />
          <h1 className="sidebar-title">{t('login.title')}</h1>
        </div>

        <nav className="sidebar-nav">
          <NavLink
            to="/dashboard/meetings"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon">📅</span>
            <span className="nav-label">{t('nav.meetings')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/search"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon">🔍</span>
            <span className="nav-label">{t('nav.search')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/documents"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon">📄</span>
            <span className="nav-label">{t('nav.documents')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/management"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
          >
            <span className="nav-icon">⚙️</span>
            <span className="nav-label">{t('nav.management')}</span>
          </NavLink>
        </nav>

        <div className="sidebar-footer">
          <div className="language-switcher-wrapper">
            <LanguageSwitcher />
          </div>
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon">🚪</span>
            <span className="nav-label">{t('common.logout')}</span>
          </button>
        </div>
      </aside>

      <div className="main-content">
        <header className="top-header">
          <div className="header-left">
            <h2 className="page-current-title">{t('login.title')}</h2>
          </div>
          <div className="header-right">
            <span className="user-info">User</span>
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
