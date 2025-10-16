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
    <div className="dashboard-container">
      <header className="dashboard-header">
        <div className="header-left">
          <img src="/logo.png" alt="Recontext Logo" className="header-logo" />
          <h1 className="header-title">{t('login.title')}</h1>
        </div>
        <div className="header-right">
          <LanguageSwitcher />
          <button onClick={handleLogout} className="btn btn-secondary">
            {t('common.logout')}
          </button>
        </div>
      </header>

      <nav className="dashboard-nav">
        <NavLink to="/dashboard/meetings" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          {t('nav.meetings')}
        </NavLink>
        <NavLink to="/dashboard/search" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          {t('nav.search')}
        </NavLink>
        <NavLink to="/dashboard/documents" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          {t('nav.documents')}
        </NavLink>
        <NavLink to="/dashboard/management" className={({ isActive }) => isActive ? 'nav-link active' : 'nav-link'}>
          {t('nav.management')}
        </NavLink>
      </nav>

      <main className="dashboard-main">
        <Outlet />
      </main>
    </div>
  );
};

export default Dashboard;
