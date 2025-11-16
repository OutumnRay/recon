import React, { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { LuUser, LuMail, LuCamera, LuSave, LuX, LuGlobe, LuBell } from 'react-icons/lu';
import './Profile.css';

interface User {
  id: string;
  username: string;
  email: string;
  firstName?: string;
  lastName?: string;
  phone?: string;
  bio?: string;
  avatar?: string;
  role: string;
  language: string;
  notification_preferences?: string;
}

export const Profile: React.FC = () => {
  const { t, i18n } = useTranslation();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [user, setUser] = useState<User | null>(null);
  const [editMode, setEditMode] = useState(false);
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    phone: '',
    bio: '',
    language: 'en',
    notificationPreferences: 'both',
  });

  useEffect(() => {
    loadUserData();
  }, []);

  const loadUserData = () => {
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (storedUser) {
      const userData = JSON.parse(storedUser);
      setUser(userData);
      setFormData({
        firstName: userData.firstName || '',
        lastName: userData.lastName || '',
        phone: userData.phone || '',
        bio: userData.bio || '',
        language: userData.language || 'en',
        notificationPreferences: userData.notification_preferences || 'both',
      });
      if (userData.avatar) {
        setAvatarPreview(userData.avatar);
      }
      // Set i18n language to match user preference
      if (userData.language && userData.language !== i18n.language) {
        i18n.changeLanguage(userData.language);
      }
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleAvatarClick = () => {
    if (editMode && fileInputRef.current) {
      fileInputRef.current.click();
    }
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate file type
    if (!file.type.startsWith('image/')) {
      setError(t('profile.errors.invalidImageType'));
      return;
    }

    // Validate file size (max 5MB)
    if (file.size > 5 * 1024 * 1024) {
      setError(t('profile.errors.imageSizeLimit'));
      return;
    }

    setSelectedFile(file);

    // Create preview
    const reader = new FileReader();
    reader.onloadend = () => {
      setAvatarPreview(reader.result as string);
    };
    reader.readAsDataURL(file);
    setError(null);
  };

  const uploadAvatar = async (): Promise<string | null> => {
    if (!selectedFile || !user) return null;

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('avatar', selectedFile);

      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${user.id}/avatar`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
        body: formData,
      });

      if (!response.ok) {
        const contentType = response.headers.get('content-type');
        let errorMessage = `Failed to upload avatar (${response.status})`;

        if (contentType?.includes('application/json')) {
          try {
            const errorData = await response.json();
            errorMessage = errorData.message || errorData.error || errorMessage;
          } catch {
            // If JSON parsing fails, use default message
          }
        } else {
          // Server returned HTML or other non-JSON response
          const text = await response.text();
          console.error('Server response:', text.substring(0, 200));
          errorMessage = `Server error: ${response.statusText || 'Unknown error'}`;
        }

        throw new Error(errorMessage);
      }

      const contentType = response.headers.get('content-type');
      if (!contentType?.includes('application/json')) {
        throw new Error('Server returned invalid response format');
      }

      const data = await response.json();
      return data.avatarUrl || null;
    } catch (err) {
      console.error('Avatar upload error:', err);
      throw err; // Re-throw to handle in handleSave
    } finally {
      setUploading(false);
    }
  };

  const handleSave = async () => {
    if (!user) return;

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      // Upload avatar if selected
      let avatarUrl = user.avatar;
      if (selectedFile) {
        try {
          const uploadedUrl = await uploadAvatar();
          if (uploadedUrl) {
            avatarUrl = uploadedUrl;
          } else {
            throw new Error(t('profile.errors.avatarUploadFailed'));
          }
        } catch (uploadErr) {
          // Fallback: use base64 preview if API upload fails
          console.warn('Avatar upload to server failed, using local preview:', uploadErr);
          if (avatarPreview && avatarPreview.startsWith('data:')) {
            avatarUrl = avatarPreview;
          } else {
            const uploadError = uploadErr instanceof Error ? uploadErr.message : t('profile.errors.avatarUploadFailed');
            setError(uploadError);
            setLoading(false);
            return;
          }
        }
      }

      // Update user profile
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${user.id}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          firstName: formData.firstName,
          lastName: formData.lastName,
          phone: formData.phone,
          bio: formData.bio,
          language: formData.language,
          notification_preferences: formData.notificationPreferences,
          avatar: avatarUrl,
        }),
      });

      if (!response.ok) {
        const contentType = response.headers.get('content-type');
        let errorMessage = `Failed to update profile (${response.status})`;

        if (contentType?.includes('application/json')) {
          try {
            const errorData = await response.json();
            errorMessage = errorData.message || errorData.error || errorMessage;
          } catch {
            // If JSON parsing fails, use default message
          }
        } else {
          const text = await response.text();
          console.error('Server response:', text.substring(0, 200));
          errorMessage = `Server error: ${response.statusText || 'Unknown error'}`;
        }

        throw new Error(errorMessage);
      }

      const contentType = response.headers.get('content-type');
      if (!contentType?.includes('application/json')) {
        throw new Error('Server returned invalid response format');
      }

      const updatedUser = await response.json();

      // Update stored user
      const storage = localStorage.getItem('user') ? localStorage : sessionStorage;
      storage.setItem('user', JSON.stringify(updatedUser));
      setUser(updatedUser);
      setSelectedFile(null);
      setEditMode(false);

      // Change language if it was updated
      if (formData.language !== i18n.language) {
        await i18n.changeLanguage(formData.language);
      }

      setSuccess(t('profile.saveSuccess'));

      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('profile.saveError'));
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    if (user) {
      setFormData({
        firstName: user.firstName || '',
        lastName: user.lastName || '',
        phone: user.phone || '',
        bio: user.bio || '',
        language: user.language || 'en',
        notificationPreferences: user.notification_preferences || 'both',
      });
      setAvatarPreview(user.avatar || null);
      setSelectedFile(null);
    }
    setEditMode(false);
    setError(null);
  };

  const removeAvatar = () => {
    setAvatarPreview(null);
    setSelectedFile(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  if (!user) {
    return (
      <div className="profile-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('common.loading')}</p>
        </div>
      </div>
    );
  }

  const displayName = user.firstName && user.lastName
    ? `${user.firstName} ${user.lastName}`
    : user.username;

  return (
    <div className="profile-container">
      <div className="profile-header">
        <h1 className="profile-title">
          <LuUser className="title-icon" />
          {t('profile.title')}
        </h1>
        {!editMode && (
          <button
            onClick={() => setEditMode(true)}
            className="btn btn-secondary"
          >
            {t('common.edit')}
          </button>
        )}
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

      <div className="profile-content">
        <div className="profile-sidebar">
          <div className="avatar-section">
            <div
              className={`avatar-container ${editMode ? 'editable' : ''}`}
              onClick={handleAvatarClick}
            >
              {avatarPreview ? (
                <img src={avatarPreview} alt={displayName} className="avatar-image" />
              ) : (
                <div className="avatar-placeholder">
                  <LuUser className="avatar-placeholder-icon" />
                </div>
              )}
              {editMode && (
                <div className="avatar-overlay">
                  <LuCamera className="camera-icon" />
                  <span className="avatar-overlay-text">{t('profile.changeAvatar')}</span>
                </div>
              )}
            </div>

            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              onChange={handleFileSelect}
              className="avatar-input"
            />

            {editMode && avatarPreview && (
              <button
                onClick={removeAvatar}
                className="btn btn-text btn-remove-avatar"
                type="button"
              >
                <LuX />
                {t('profile.removeAvatar')}
              </button>
            )}
          </div>

          <div className="user-info-card">
            <h2 className="user-name">{displayName}</h2>
            <p className="user-role">{user.role}</p>
          </div>
        </div>

        <div className="profile-main">
          <div className="profile-card">
            <div className="profile-section">
              <h2 className="section-title">{t('profile.accountInfo')}</h2>

              <div className="info-row">
                <div className="info-label">
                  <LuUser className="info-icon" />
                  {t('profile.username')}
                </div>
                <div className="info-value">{user.username}</div>
              </div>

              <div className="info-row">
                <div className="info-label">
                  <LuMail className="info-icon" />
                  {t('profile.email')}
                </div>
                <div className="info-value">{user.email}</div>
              </div>
            </div>

            <div className="profile-section">
              <h2 className="section-title">{t('profile.preferences')}</h2>

              <div className="form-group">
                <label htmlFor="language" className="form-label">
                  <LuGlobe className="info-icon" />
                  {t('profile.language')}
                </label>
                <select
                  id="language"
                  name="language"
                  value={formData.language}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-input"
                >
                  <option value="en">{t('languages.english')}</option>
                  <option value="ru">{t('languages.russian')}</option>
                </select>
                <p className="form-hint">{t('profile.languageDescription')}</p>
              </div>

              <div className="form-group">
                <label htmlFor="notificationPreferences" className="form-label">
                  <LuBell className="info-icon" />
                  {t('profile.notificationPreferences')}
                </label>
                <select
                  id="notificationPreferences"
                  name="notificationPreferences"
                  value={formData.notificationPreferences}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-input"
                >
                  <option value="tracks">{t('profile.notificationTracks')}</option>
                  <option value="rooms">{t('profile.notificationRooms')}</option>
                  <option value="both">{t('profile.notificationBoth')}</option>
                </select>
                <p className="form-hint">{t('profile.notificationDescription')}</p>
              </div>
            </div>

            <div className="profile-section">
              <h2 className="section-title">{t('profile.personalInfo')}</h2>

              <div className="form-group">
                <label htmlFor="firstName" className="form-label">
                  {t('profile.firstName')}
                </label>
                <input
                  type="text"
                  id="firstName"
                  name="firstName"
                  value={formData.firstName}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-input"
                  placeholder={t('profile.firstNamePlaceholder')}
                />
              </div>

              <div className="form-group">
                <label htmlFor="lastName" className="form-label">
                  {t('profile.lastName')}
                </label>
                <input
                  type="text"
                  id="lastName"
                  name="lastName"
                  value={formData.lastName}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-input"
                  placeholder={t('profile.lastNamePlaceholder')}
                />
              </div>

              <div className="form-group">
                <label htmlFor="phone" className="form-label">
                  {t('profile.phone')}
                </label>
                <input
                  type="tel"
                  id="phone"
                  name="phone"
                  value={formData.phone}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-input"
                  placeholder={t('profile.phonePlaceholder')}
                />
              </div>

              <div className="form-group">
                <label htmlFor="bio" className="form-label">
                  {t('profile.bio')}
                </label>
                <textarea
                  id="bio"
                  name="bio"
                  value={formData.bio}
                  onChange={handleInputChange}
                  disabled={!editMode}
                  className="form-textarea"
                  placeholder={t('profile.bioPlaceholder')}
                  rows={4}
                />
              </div>
            </div>

            {editMode && (
              <div className="profile-actions">
                <button
                  onClick={handleCancel}
                  className="btn btn-secondary"
                  disabled={loading}
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleSave}
                  disabled={loading || uploading}
                  className="btn btn-success"
                >
                  {loading || uploading ? (
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
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Profile;
