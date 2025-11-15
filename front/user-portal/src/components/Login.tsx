import React, { useState } from 'react';
import type { FormEvent } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LanguageSwitcher } from './LanguageSwitcher';
import './Login.css';

interface LoginResponse {
  token: string;
  expiresAt: string;
  user: {
    id: string;
    username: string;
    email: string;
    role: string;
    language: string;
  };
}

interface ErrorResponse {
  error: string;
  message: string;
  code: number;
}

export const Login: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();

  // Get the page user was trying to access before being redirected to login
  const from = (location.state as any)?.from || '/dashboard/meetings';

  const clearStoredAuth = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user');
  };

  // Set page title
  React.useEffect(() => {
    document.title = `Recontext - ${t('login.title')}`;
  }, [t]);

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [isAutoLoggingIn, setIsAutoLoggingIn] = useState(() => {
    if (typeof window === 'undefined') return false;
    return Boolean(localStorage.getItem('token') || sessionStorage.getItem('token'));
  });

  // Attempt automatic login when a valid token is already stored
  React.useEffect(() => {
    let isCancelled = false;
    const storedToken = localStorage.getItem('token') || sessionStorage.getItem('token');

    if (!storedToken) {
      setIsAutoLoggingIn(false);
      return;
    }

    const attemptAutoLogin = async () => {
      try {
        const response = await fetch('/api/v1/meetings?limit=1', {
          headers: {
            Authorization: `Bearer ${storedToken}`,
          },
        });

        if (!isCancelled && response.ok) {
          navigate(from, { replace: true });
          return;
        }
      } catch (autoErr) {
        console.warn('[Login] Auto-login failed:', autoErr);
      }

      if (!isCancelled) {
        clearStoredAuth();
        setIsAutoLoggingIn(false);
      }
    };

    attemptAutoLogin();

    return () => {
      isCancelled = true;
    };
  }, [from, navigate]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    setLoading(true);

    try {
      const response = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ username, password }),
      });

      const data = await response.json();

      if (!response.ok) {
        const errorData = data as ErrorResponse;
        throw new Error(errorData.error || t('login.errors.loginFailed'));
      }

      const loginData = data as LoginResponse;

      // Apply user's preferred language if available
      if (loginData.user.language) {
        await i18n.changeLanguage(loginData.user.language);
        localStorage.setItem('i18nextLng', loginData.user.language);
      }

      // Store token
      if (rememberMe) {
        localStorage.setItem('token', loginData.token);
        localStorage.setItem('user', JSON.stringify(loginData.user));
      } else {
        sessionStorage.setItem('token', loginData.token);
        sessionStorage.setItem('user', JSON.stringify(loginData.user));
      }

      setSuccess(t('login.loginSuccess'));

      // Redirect to the page user was trying to access, or dashboard
      setTimeout(() => {
        window.location.href = from;
      }, 1000);

    } catch (err) {
      setError(err instanceof Error ? err.message : t('login.errors.networkError'));
    } finally {
      setLoading(false);
    }
  };

  const formDisabled = loading || isAutoLoggingIn;

  return (
    <div className="login-container">
      <div className="language-switcher-wrapper">
        <LanguageSwitcher />
      </div>
      <div className="login-card surface-card elevated">
        <div className="login-header">
          <div className="login-logo">
            <img src="/logo.png" alt="Recontext Logo" />
          </div>
          <h1 className="login-title">{t('login.title')}</h1>
          <p className="login-subtitle">
            {t('login.subtitle')}
          </p>
        </div>

        {isAutoLoggingIn && (
          <div className="alert alert-info">
            <span>{t('login.loggingIn')}</span>
          </div>
        )}

        {error && (
          <div className="alert alert-error">
            <span>{error}</span>
          </div>
        )}

        {success && (
          <div className="alert alert-success">
            <span>{success}</span>
          </div>
        )}

        <form className="login-form" onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="username" className="form-label">
              {t('login.usernameLabel')}
            </label>
            <input
              id="username"
              type="text"
              className="form-input"
              placeholder={t('login.usernamePlaceholder')}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={formDisabled}
              required
              autoComplete="username"
              autoFocus
            />
          </div>

          <div className="form-group">
            <label htmlFor="password" className="form-label">
              {t('login.passwordLabel')}
            </label>
            <input
              id="password"
              type="password"
              className="form-input"
              placeholder={t('login.passwordPlaceholder')}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={formDisabled}
              required
              autoComplete="current-password"
            />
          </div>

          <div className="form-options">
            <label className="form-checkbox">
              <input
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                disabled={formDisabled}
              />
              <span>{t('login.rememberMe')}</span>
            </label>
            <button
              type="button"
              onClick={() => navigate('/forgot-password')}
              className="form-link"
              disabled={formDisabled}
            >
              {t('login.forgotPassword')}
            </button>
          </div>

          <button
            type="submit"
            className={`btn btn-primary ${(loading || isAutoLoggingIn) ? 'btn-loading' : ''}`}
            disabled={formDisabled || !username || !password}
          >
            {loading ? t('login.loggingIn') : t('login.loginButton')}
          </button>
        </form>

        <div className="login-footer">
                    <p className="login-footer-note">
            {t('login.needHelp')}{' '}
            <a href="mailto:support@recontext.ru" className="login-footer-link">
              {t('login.contactSupport')}
            </a>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
