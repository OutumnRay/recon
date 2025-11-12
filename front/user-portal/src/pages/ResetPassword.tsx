import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LuKey, LuEye, LuEyeOff, LuArrowLeft, LuLoader, LuInfo, LuCheck } from 'react-icons/lu';
import './ResetPassword.css';

export const ResetPassword = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();

  const tokenIdParam = searchParams.get('token');
  const emailParam = searchParams.get('email');

  const [code, setCode] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [verifying, setVerifying] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [codeValid, setCodeValid] = useState(false);
  const [resetSuccess, setResetSuccess] = useState(false);

  useEffect(() => {
    if (!tokenIdParam) {
      navigate('/forgot-password');
    }
  }, [tokenIdParam, navigate]);

  const handleVerifyCode = async () => {
    if (code.length !== 6) {
      setError(t('resetPassword.errors.codeLength'));
      return;
    }

    setVerifying(true);
    setError(null);

    try {
      const response = await fetch('/api/v1/auth/password-reset/verify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token_id: tokenIdParam,
          code: code,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || t('resetPassword.errors.verifyFailed'));
      }

      if (data.valid) {
        setCodeValid(true);
        setError(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : t('resetPassword.errors.verifyFailed'));
      setCodeValid(false);
    } finally {
      setVerifying(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (newPassword.length < 8) {
      setError(t('resetPassword.errors.passwordTooShort'));
      return;
    }

    if (newPassword !== confirmPassword) {
      setError(t('resetPassword.errors.passwordMismatch'));
      return;
    }

    if (!codeValid) {
      setError(t('resetPassword.errors.codeNotVerified'));
      return;
    }

    setLoading(true);

    try {
      const response = await fetch('/api/v1/auth/password-reset/reset', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          token_id: tokenIdParam,
          code: code,
          new_password: newPassword,
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || t('resetPassword.errors.resetFailed'));
      }

      setResetSuccess(true);

      // Redirect to login after 3 seconds
      setTimeout(() => {
        navigate('/login');
      }, 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('resetPassword.errors.resetFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleBackToLogin = () => {
    navigate('/login');
  };

  if (resetSuccess) {
    return (
      <div className="reset-password-page">
        <div className="reset-password-container">
          <div className="success-card">
            <div className="success-icon">
              <LuCheck />
            </div>
            <h2 className="success-title">{t('resetPassword.successTitle')}</h2>
            <p className="success-message">{t('resetPassword.successMessage')}</p>
            <button onClick={handleBackToLogin} className="btn btn-primary btn-full">
              {t('common.login')}
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="reset-password-page">
      <div className="reset-password-container">
        <button onClick={handleBackToLogin} className="back-button">
          <LuArrowLeft /> {t('common.backToLogin')}
        </button>

        <div className="reset-password-card">
          <div className="card-header">
            <h1 className="card-title">{t('resetPassword.title')}</h1>
            <p className="card-description">
              {emailParam && t('resetPassword.description', { email: emailParam })}
            </p>
          </div>

          {error && (
            <div className="alert alert-error">
              <LuInfo />
              <span>{error}</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="reset-password-form">
            {/* Verification Code */}
            <div className="form-group">
              <label htmlFor="code" className="form-label">
                {t('resetPassword.codeLabel')}
              </label>
              <div className="code-input-wrapper">
                <input
                  id="code"
                  type="text"
                  value={code}
                  onChange={(e) => {
                    const value = e.target.value.replace(/\D/g, '').slice(0, 6);
                    setCode(value);
                    setCodeValid(false);
                  }}
                  className={`form-input code-input ${codeValid ? 'valid' : ''}`}
                  placeholder="000000"
                  maxLength={6}
                  required
                  disabled={loading || codeValid}
                  autoFocus
                />
                {!codeValid && (
                  <button
                    type="button"
                    onClick={handleVerifyCode}
                    className="btn btn-secondary btn-sm verify-btn"
                    disabled={code.length !== 6 || verifying}
                  >
                    {verifying ? (
                      <>
                        <LuLoader className="spinner" />
                        {t('resetPassword.verifying')}
                      </>
                    ) : (
                      t('resetPassword.verify')
                    )}
                  </button>
                )}
                {codeValid && (
                  <div className="verified-badge">
                    <LuCheck /> {t('resetPassword.verified')}
                  </div>
                )}
              </div>
              <p className="form-hint">{t('resetPassword.codeHint')}</p>
            </div>

            {/* New Password */}
            <div className="form-group">
              <label htmlFor="new-password" className="form-label">
                {t('resetPassword.newPasswordLabel')}
              </label>
              <div className="input-with-icon">
                <LuKey className="input-icon" />
                <input
                  id="new-password"
                  type={showPassword ? 'text' : 'password'}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  className="form-input"
                  placeholder={t('resetPassword.newPasswordPlaceholder')}
                  required
                  disabled={loading || !codeValid}
                  minLength={8}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="toggle-password"
                  disabled={!codeValid}
                >
                  {showPassword ? <LuEyeOff /> : <LuEye />}
                </button>
              </div>
            </div>

            {/* Confirm Password */}
            <div className="form-group">
              <label htmlFor="confirm-password" className="form-label">
                {t('resetPassword.confirmPasswordLabel')}
              </label>
              <div className="input-with-icon">
                <LuKey className="input-icon" />
                <input
                  id="confirm-password"
                  type={showConfirmPassword ? 'text' : 'password'}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="form-input"
                  placeholder={t('resetPassword.confirmPasswordPlaceholder')}
                  required
                  disabled={loading || !codeValid}
                  minLength={8}
                />
                <button
                  type="button"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  className="toggle-password"
                  disabled={!codeValid}
                >
                  {showConfirmPassword ? <LuEyeOff /> : <LuEye />}
                </button>
              </div>
            </div>

            <button
              type="submit"
              className="btn btn-primary btn-full"
              disabled={loading || !codeValid || !newPassword || !confirmPassword}
            >
              {loading ? (
                <>
                  <LuLoader className="spinner" />
                  {t('resetPassword.resetting')}
                </>
              ) : (
                t('resetPassword.resetButton')
              )}
            </button>
          </form>

          <div className="card-footer">
            <p className="footer-text">
              {t('resetPassword.didntReceive')}{' '}
              <button
                onClick={() => navigate('/forgot-password')}
                className="link-button"
              >
                {t('resetPassword.resendCode')}
              </button>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ResetPassword;
