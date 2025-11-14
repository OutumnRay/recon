import React, { useState, useEffect, useRef } from 'react';
import { NavLink, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuCalendar, LuSearch, LuFileText, LuSettings, LuLogOut, LuMenu, LuX, LuUser, LuChevronDown } from 'react-icons/lu';
import { LanguageSwitcher } from './LanguageSwitcher';
import { APP_VERSION } from '../config/version';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [hasFilePermission, setHasFilePermission] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    checkFilePermission();
  }, []);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
        setIsUserMenuOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
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

  const getUser = () => {
    const userStr = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        return {
          username: user.username || user.email || 'User',
          avatar: user.avatar || null,
          firstName: user.firstName || null,
          lastName: user.lastName || null,
        };
      } catch {
        return { username: 'User', avatar: null, firstName: null, lastName: null };
      }
    }
    return { username: 'User', avatar: null, firstName: null, lastName: null };
  };

  const user = getUser();
  const displayName = user.firstName && user.lastName
    ? `${user.firstName} ${user.lastName}`
    : user.username;

  const toggleUserMenu = () => {
    setIsUserMenuOpen(!isUserMenuOpen);
  };

  const handleProfileClick = () => {
    setIsUserMenuOpen(false);
    navigate('/dashboard/profile');
  };

  const handleLogout = () => {
    setIsUserMenuOpen(false);
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

    if (location.pathname.startsWith('/dashboard/profile')) {
      return { title: t('profile.title'), subtitle: t('profile.subtitle') };
    }

    if (location.pathname.startsWith('/dashboard/meetings')) {
      return { title: t('meetings.title'), subtitle: t('meetings.subtitle') };
    }

    return { title: 'Recontext', subtitle: t('nav.dashboard', { defaultValue: '' }) };
  })();

  const handleMobileMenuToggle = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const closeMobileMenu = () => {
    setIsMobileMenuOpen(false);
  };

  return (
    <div className="dashboard-layout">
      {isMobileMenuOpen && (
        <div className="sidebar-overlay active" onClick={closeMobileMenu}></div>
      )}
      <aside className={`sidebar ${isMobileMenuOpen ? 'mobile-open' : ''}`}>
        <div className="sidebar-header">
          <img src="/logo.png" alt="Recontext Logo" className="sidebar-logo" />
          <h1 className="sidebar-title">Recontext</h1>
        </div>

        <nav className="sidebar-nav">
          <NavLink
            to="/dashboard/meetings"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
            onClick={closeMobileMenu}
          >
            <span className="nav-icon"><LuCalendar /></span>
            <span className="nav-label">{t('nav.meetings')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/search"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
            onClick={closeMobileMenu}
          >
            <span className="nav-icon"><LuSearch /></span>
            <span className="nav-label">{t('nav.search')}</span>
          </NavLink>

          {hasFilePermission && (
            <NavLink
              to="/dashboard/documents"
              className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
              onClick={closeMobileMenu}
            >
              <span className="nav-icon"><LuFileText /></span>
              <span className="nav-label">{t('nav.documents')}</span>
            </NavLink>
          )}

          <NavLink
            to="/dashboard/management"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
            onClick={closeMobileMenu}
          >
            <span className="nav-icon"><LuSettings /></span>
            <span className="nav-label">{t('nav.management')}</span>
          </NavLink>

          <NavLink
            to="/dashboard/profile"
            className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}
            onClick={closeMobileMenu}
          >
            <span className="nav-icon"><LuUser /></span>
            <span className="nav-label">{t('nav.profile')}</span>
          </NavLink>
        </nav>

        <div className="sidebar-footer">
          <button onClick={handleLogout} className="logout-btn">
            <span className="nav-icon"><LuLogOut /></span>
            <span className="nav-label">{t('common.logout')}</span>
          </button>
          <div className="sidebar-version">
            v{APP_VERSION}
          </div>
        </div>
      </aside>

      <div className="main-content">
        <header className="top-header">
          <div className="header-left">
            <button
              className="sidebar-toggle"
              onClick={handleMobileMenuToggle}
              title={isMobileMenuOpen ? t('common.close') : t('common.menu')}
            >
              {isMobileMenuOpen ? <LuX /> : <LuMenu />}
            </button>
            <div className="page-context">
              <h2 className="page-current-title">{pageMeta.title}</h2>
              {pageMeta.subtitle && (
                <p className="page-current-description">{pageMeta.subtitle}</p>
              )}
            </div>
          </div>
          <div className="header-right">
            <div className="header-language-switcher">
              <LanguageSwitcher />
            </div>

            <div className="user-menu-container" ref={userMenuRef}>
              <button
                className="user-menu-trigger"
                onClick={toggleUserMenu}
                aria-label="User menu"
              >
                {user.avatar ? (
                  <img src={user.avatar} alt={displayName} className="user-avatar" />
                ) : (
                  <div className="user-avatar-placeholder">
                    <LuUser />
                  </div>
                )}
                <span className="user-display-name">{displayName}</span>
                <LuChevronDown className={`chevron-icon ${isUserMenuOpen ? 'open' : ''}`} />
              </button>

              {isUserMenuOpen && (
                <div className="user-menu-dropdown">
                  <div className="user-menu-header">
                    <div className="user-menu-info">
                      <p className="user-menu-name">{displayName}</p>
                      <p className="user-menu-username">@{user.username}</p>
                    </div>
                  </div>

                  <div className="user-menu-divider"></div>

                  <button
                    className="user-menu-item"
                    onClick={handleProfileClick}
                  >
                    <LuUser className="menu-item-icon" />
                    <span>{t('nav.profile')}</span>
                  </button>

                  <div className="user-menu-divider"></div>

                  <button
                    className="user-menu-item logout"
                    onClick={handleLogout}
                  >
                    <LuLogOut className="menu-item-icon" />
                    <span>{t('common.logout')}</span>
                  </button>
                </div>
              )}
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
