import React, { useState } from 'react';
import type { FormEvent } from 'react';
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
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

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

      // Redirect to dashboard after short delay
      setTimeout(() => {
        window.location.href = '/dashboard';
      }, 1000);

    } catch (err) {
      setError(err instanceof Error ? err.message : t('login.errors.networkError'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="login-container">
      <div className="language-switcher-wrapper">
        <LanguageSwitcher />
      </div>
      <div className="login-card">
        <div className="login-header">
          <div className="login-logo">
            <img src="/logo.png" alt="Recontext Logo" />
          </div>
          <h1 className="login-title">{t('login.title')}</h1>
          <p className="login-subtitle">
            {t('login.subtitle')}
          </p>
        </div>

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
              disabled={loading}
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
              disabled={loading}
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
                disabled={loading}
              />
              <span>{t('login.rememberMe')}</span>
            </label>
            <a href="#forgot-password" className="form-link">
              {t('login.forgotPassword')}
            </a>
          </div>

          <button
            type="submit"
            className={`btn btn-primary ${loading ? 'btn-loading' : ''}`}
            disabled={loading || !username || !password}
          >
            {loading ? t('login.loggingIn') : t('login.loginButton')}
          </button>
        </form>

        <div className="login-footer">
          <p>
            {t('login.defaultCredentials')} <strong>user / user123</strong>
          </p>
          <p style={{ marginTop: '8px' }}>
            {t('login.needHelp')}{' '}
            <a href="#support" className="login-footer-link">
              {t('login.contactSupport')}
            </a>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;
