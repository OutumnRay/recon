import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { LuUser, LuGlobe, LuMail, LuSave } from 'react-icons/lu';
import './UserSettings.css';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  language: string;
}

export const UserSettings: React.FC = () => {
  const { t, i18n } = useTranslation();
  const [user, setUser] = useState<User | null>(null);
  const [language, setLanguage] = useState('en');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Load user data
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (storedUser) {
      const userData = JSON.parse(storedUser);
      setUser(userData);
      setLanguage(userData.language || 'en');
    }
  }, []);

  const handleSave = async () => {
    if (!user) return;

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${user.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          language,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to update settings');
      }

      const updatedUser = await response.json();

      // Update stored user
      const storage = localStorage.getItem('user') ? localStorage : sessionStorage;
      storage.setItem('user', JSON.stringify(updatedUser));
      setUser(updatedUser);

      // Update language
      await i18n.changeLanguage(language);
      localStorage.setItem('i18nextLng', language);

      setSuccess(t('settings.saveSuccess'));
    } catch (err) {
      setError(err instanceof Error ? err.message : t('settings.saveError'));
    } finally {
      setLoading(false);
    }
  };

  if (!user) {
    return (
      <div className="user-settings-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('common.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-settings-container">
      <div className="settings-header">
        <h1 className="settings-title">
          <LuUser className="title-icon" />
          {t('settings.title')}
        </h1>
      </div>

      {success && (
        <div className="alert alert-success">
          {success}
        </div>
      )}

      {error && (
        <div className="alert alert-error">
          {error}
        </div>
      )}

      <div className="settings-card">
        <div className="settings-section">
          <h2 className="section-title">{t('settings.accountInfo')}</h2>

          <div className="info-row">
            <div className="info-label">
              <LuUser className="info-icon" />
              {t('settings.username')}
            </div>
            <div className="info-value">{user.username}</div>
          </div>

          <div className="info-row">
            <div className="info-label">
              <LuMail className="info-icon" />
              {t('settings.email')}
            </div>
            <div className="info-value">{user.email}</div>
          </div>
        </div>

        <div className="settings-section">
          <h2 className="section-title">{t('settings.preferences')}</h2>

          <div className="form-group">
            <label htmlFor="language" className="form-label">
              <LuGlobe className="info-icon" />
              {t('settings.language')}
            </label>
            <select
              id="language"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
              className="form-select"
            >
              <option value="en">{t('languages.english')}</option>
              <option value="ru">{t('languages.russian')}</option>
            </select>
            <p className="form-hint">{t('settings.languageHint')}</p>
          </div>
        </div>

        <div className="settings-actions">
          <button
            onClick={handleSave}
            disabled={loading}
            className="btn btn-primary"
          >
            {loading ? (
              <>
                <div className="btn-spinner"></div>
                {t('common.saving')}
              </>
            ) : (
              <>
                <LuSave />
                {t('common.save')}
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};

export default UserSettings;
