import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { LanguageSwitcher } from '../components/LanguageSwitcher';
import './AnonymousJoin.css';

export default function AnonymousJoin() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const [displayName, setDisplayName] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    document.title = `Recontext - ${t('anonymousJoin.title')}`;
  }, [t]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!displayName.trim()) {
      setError(t('anonymousJoin.errors.nameRequired'));
      return;
    }

    if (displayName.trim().length > 255) {
      setError(t('anonymousJoin.errors.nameTooLong'));
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // Get auth token if user is logged in
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };

      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`/api/v1/meetings/${meetingId}/join-anonymous`, {
        method: 'POST',
        headers,
        body: JSON.stringify({
          displayName: displayName.trim(),
        }),
      });

      if (!response.ok) {
        // If 403 Forbidden, redirect to forbidden page
        if (response.status === 403) {
          navigate('/forbidden');
          return;
        }
        const errorData = await response.json();
        throw new Error(errorData.error || t('anonymousJoin.errors.joinFailed'));
      }

      const data = await response.json();

      // Check if user is authenticated and should be redirected to regular meeting
      if (data.redirect) {
        navigate(`/meeting/${data.meetingId}`);
        return;
      }

      // Redirect to meeting room with the token and meeting info
      navigate(`/meeting-room/${meetingId}`, {
        state: {
          token: data.token,
          url: data.url,
          roomName: data.roomName,
          participantName: data.participantName,
          meetingTitle: data.meetingTitle,
          scheduledAt: data.scheduledAt,
          duration: data.duration,
          forceEndAt: data.forceEndAt,
          isAnonymous: true,
        },
      });
    } catch (err: any) {
      console.error('Anonymous join error:', err);
      setError(err.message || t('anonymousJoin.errors.joinFailed'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="anonymous-join-page">
      <div className="anonymous-language-switcher">
        <LanguageSwitcher />
      </div>

      <div className="anonymous-join-card surface-card elevated">
        <div className="anonymous-join-header">
          <div className="anonymous-logo">
            <img src="/logo.png" alt="Recontext Logo" />
          </div>
          <h1>{t('anonymousJoin.title')}</h1>
          <p className="anonymous-join-description">
            {t('anonymousJoin.description')}
          </p>
        </div>

        {error && (
          <div className="alert alert-error" role="alert">
            <span>{error}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="anonymous-join-form">
          <div className="form-group">
            <label htmlFor="displayName" className="form-label">
              {t('anonymousJoin.nameLabel')}
            </label>
            <input
              type="text"
              id="displayName"
              value={displayName}
              onChange={(e) => {
                setDisplayName(e.target.value);
                if (error) {
                  setError(null);
                }
              }}
              placeholder={t('anonymousJoin.namePlaceholder')}
              maxLength={255}
              disabled={loading}
              autoFocus
              required
              className="form-input"
            />
            <p className="form-hint">
              {t('anonymousJoin.nameHint')}
            </p>
          </div>

          <button
            type="submit"
            className="btn btn-primary btn-lg join-button"
            disabled={loading || !displayName.trim()}
          >
            {loading ? t('anonymousJoin.joining') : t('anonymousJoin.joinButton')}
          </button>
        </form>

        <div className="anonymous-join-footer">
          <p>{t('anonymousJoin.footer')}</p>
        </div>
      </div>
    </div>
  );
}
