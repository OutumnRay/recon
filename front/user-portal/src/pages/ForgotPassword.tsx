import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuMail, LuArrowLeft, LuLoader, LuInfo } from 'react-icons/lu';
import './ForgotPassword.css';

export const ForgotPassword = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [tokenId, setTokenId] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);

    try {
      const response = await fetch('/api/v1/auth/password-reset/request', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || t('forgotPassword.errors.requestFailed'));
      }

      setTokenId(data.token_id);
      setSuccess(true);

      // Redirect to reset page after 2 seconds
      setTimeout(() => {
        navigate(`/reset-password?token=${data.token_id}&email=${encodeURIComponent(email)}`);
      }, 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('forgotPassword.errors.requestFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleBackToLogin = () => {
    navigate('/login');
  };

  if (success) {
    return (
      <div className="forgot-password-page">
        <div className="forgot-password-container">
          <div className="success-card">
            <div className="success-icon">✉️</div>
            <h2 className="success-title">{t('forgotPassword.successTitle')}</h2>
            <p className="success-message">{t('forgotPassword.successMessage', { email })}</p>
            <div className="success-instructions">
              <p>{t('forgotPassword.checkEmail')}</p>
              <p className="code-validity">{t('forgotPassword.codeValidity')}</p>
            </div>
            <button
              onClick={() => navigate(`/reset-password?token=${tokenId}&email=${encodeURIComponent(email)}`)}
              className="btn btn-primary btn-full"
            >
              {t('forgotPassword.enterCode')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="forgot-password-page">
      <div className="forgot-password-container">
        <button onClick={handleBackToLogin} className="back-button">
          <LuArrowLeft /> {t('common.backToLogin')}
        </button>

        <div className="forgot-password-card">
          <div className="card-header">
            <div className="lock-icon">🔐</div>
            <h1 className="card-title">{t('forgotPassword.title')}</h1>
            <p className="card-description">{t('forgotPassword.description')}</p>
          </div>

          {error && (
            <div className="alert alert-error">
              <LuInfo />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="forgot-password-form">
            <div className="form-group">
              <label htmlFor="email" className="form-label">
                {t('forgotPassword.emailLabel')}
              </label>
              <div className="input-with-icon">
                <LuMail className="input-icon" />
                <input
                  id="email"
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="form-input"
                  placeholder={t('forgotPassword.emailPlaceholder')}
                  required
                  disabled={loading}
                  autoFocus
                />
              </div>
            </div>

            <button
              type="submit"
              className="btn btn-primary btn-full"
              disabled={loading}
            >
              {loading ? (
                <>
                  <LuLoader className="spinner" />
                  {t('forgotPassword.sending')}
                </>
              ) : (
                t('forgotPassword.sendCode')
              )}
            </button>
          </form>

          <div className="card-footer">
            <p className="footer-text">
              {t('forgotPassword.rememberPassword')}{' '}
              <button onClick={handleBackToLogin} className="link-button">
                {t('common.login')}
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ForgotPassword;
