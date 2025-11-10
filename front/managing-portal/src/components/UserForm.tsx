import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { SearchableSelect } from './SearchableSelect';
import './UserForm.css';

interface UserPermissions {
  can_schedule_meetings: boolean;
  can_manage_department: boolean;
  can_approve_recordings: boolean;
}

interface User {
  id: string;
  username: string;
  email: string;
  role: 'admin' | 'user' | 'operator';
  department_id?: string | null;
  groups?: string[];
  permissions: UserPermissions;
  language: string;
  is_active?: boolean;
}

interface Group {
  id: string;
  name: string;
  description?: string;
}

interface Department {
  id: string;
  name: string;
  path?: string;
}

export const UserForm: React.FC = () => {
  const { t } = useTranslation();

  // Extract user ID from URL path
  const pathParts = window.location.pathname.split('/');
  const id = pathParts.includes('edit') ? pathParts[pathParts.length - 2] : null;
  const isEditMode = !!id;

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState<'admin' | 'user' | 'operator'>('user');
  const [language, setLanguage] = useState('en');
  const [departmentId, setDepartmentId] = useState('');
  const [selectedGroups, setSelectedGroups] = useState<string[]>([]);
  const [permissions, setPermissions] = useState<UserPermissions>({
    can_schedule_meetings: false,
    can_manage_department: false,
    can_approve_recordings: false,
  });
  const [isActive, setIsActive] = useState(true);
  const [showPasswordReset, setShowPasswordReset] = useState(false);

  const [groups, setGroups] = useState<Group[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);

  // Function to generate a random password
  const generatePassword = () => {
    const length = 12;
    const charset = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*';
    let newPassword = '';

    // Ensure at least one uppercase, one lowercase, one number, and one special char
    newPassword += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'[Math.floor(Math.random() * 26)];
    newPassword += 'abcdefghijklmnopqrstuvwxyz'[Math.floor(Math.random() * 26)];
    newPassword += '0123456789'[Math.floor(Math.random() * 10)];
    newPassword += '!@#$%^&*'[Math.floor(Math.random() * 8)];

    // Fill the rest randomly
    for (let i = 4; i < length; i++) {
      newPassword += charset[Math.floor(Math.random() * charset.length)];
    }

    // Shuffle the password
    newPassword = newPassword.split('').sort(() => Math.random() - 0.5).join('');

    setPassword(newPassword);
  };

  useEffect(() => {
    fetchGroups();
    fetchDepartments();

    if (isEditMode && id) {
      fetchUser(id);
    }
  }, [id]);

  const fetchUser = async (userId: string) => {
    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/v1/users/${userId}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch user');
      }

      const user: User = await response.json();
      setUsername(user.username);
      setEmail(user.email);
      setRole(user.role);
      setLanguage(user.language || 'en');
      setDepartmentId(user.department_id || '');
      setSelectedGroups(user.groups || []);
      setPermissions(user.permissions);
      setIsActive(user.is_active ?? true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load user');
    } finally {
      setLoading(false);
    }
  };

  const fetchGroups = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/groups', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setGroups(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch groups:', err);
    }
  };

  const fetchDepartments = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/departments', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setDepartments(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch departments:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!isEditMode && !password) {
      setError(t('users.form.passwordRequired'));
      return;
    }

    if (isEditMode && showPasswordReset && !password) {
      setError(t('users.form.passwordRequired'));
      return;
    }

    if (isEditMode && password && password.length < 8) {
      setError(t('users.form.passwordTooShort'));
      return;
    }

    try {
      setLoading(true);
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');

      if (isEditMode) {
        // Update user
        const updateData: any = {
          email,
          role,
          department_id: departmentId || null,
          groups: selectedGroups,
          permissions,
          language,
          is_active: isActive,
        };

        // Include password only if it's being reset
        if (showPasswordReset && password) {
          updateData.password = password;
        }

        const response = await fetch(`/api/v1/users/${id}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(updateData),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.message || 'Failed to update user');
        }
      } else {
        // Create user
        const registerResponse = await fetch('/api/v1/auth/register', {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            username,
            email,
            password,
            language,
          }),
        });

        if (!registerResponse.ok) {
          const errorData = await registerResponse.json();
          throw new Error(errorData.message || 'Failed to create user');
        }

        // Get the created user to get their ID
        const createdUser = await registerResponse.json();

        // Update user with additional fields (role, department, groups, permissions)
        const updateResponse = await fetch(`/api/v1/users/${createdUser.id}`, {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            role,
            department_id: departmentId || null,
            groups: selectedGroups,
            permissions,
            language,
            is_active: isActive,
          }),
        });

        if (!updateResponse.ok) {
          throw new Error('User created but failed to update additional fields');
        }
      }

      window.location.href = '/users';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  const toggleGroup = (groupId: string) => {
    setSelectedGroups(prev =>
      prev.includes(groupId)
        ? prev.filter(id => id !== groupId)
        : [...prev, groupId]
    );
  };

  if (loading && isEditMode) {
    return (
      <div className="user-form-page">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('users.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-form-page">
      <div className="form-header">
        <h1>{isEditMode ? t('users.form.editTitle') : t('users.form.createTitle')}</h1>
        <button onClick={() => window.location.href = '/users'} className="btn btn-secondary">
          {t('users.form.cancel')}
        </button>
      </div>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="user-form">
        <div className="form-section">
          <h2>{t('users.form.accountInfo')}</h2>

          <div className="form-group">
            <label htmlFor="username">
              {t('users.form.username')} {!isEditMode && t('users.form.required')}
            </label>
            <input
              type="text"
              id="username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={isEditMode}
              required={!isEditMode}
              placeholder={t('users.form.usernamePlaceholder')}
            />
            {isEditMode && <p className="form-hint">{t('users.form.usernameHint')}</p>}
          </div>

          <div className="form-group">
            <label htmlFor="email">
              {t('users.form.email')} {t('users.form.required')}
            </label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder={t('users.form.emailPlaceholder')}
            />
          </div>

          {!isEditMode ? (
            <div className="form-group">
              <label htmlFor="password">
                {t('users.form.password')} {t('users.form.required')}
              </label>
              <div style={{ display: 'flex', gap: '8px' }}>
                <input
                  type="text"
                  id="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  minLength={8}
                  placeholder={t('users.form.passwordPlaceholder')}
                  style={{ flex: 1 }}
                />
                <button
                  type="button"
                  onClick={generatePassword}
                  className="btn btn-secondary"
                  style={{ whiteSpace: 'nowrap' }}
                >
                  {t('users.form.generatePassword')}
                </button>
              </div>
            </div>
          ) : (
            <div className="form-group">
              <label>{t('users.form.resetPassword')}</label>
              {!showPasswordReset ? (
                <button
                  type="button"
                  onClick={() => {
                    setShowPasswordReset(true);
                    setPassword('');
                  }}
                  className="btn btn-secondary"
                  style={{ width: '100%' }}
                >
                  {t('users.form.setNewPassword')}
                </button>
              ) : (
                <>
                  <div style={{ display: 'flex', gap: '8px', marginBottom: '8px' }}>
                    <input
                      type="text"
                      id="password"
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      minLength={8}
                      placeholder={t('users.form.passwordPlaceholder')}
                      style={{ flex: 1 }}
                    />
                    <button
                      type="button"
                      onClick={generatePassword}
                      className="btn btn-secondary"
                      style={{ whiteSpace: 'nowrap' }}
                    >
                      {t('users.form.generatePassword')}
                    </button>
                  </div>
                  <button
                    type="button"
                    onClick={() => {
                      setShowPasswordReset(false);
                      setPassword('');
                    }}
                    className="btn btn-ghost"
                    style={{ width: '100%' }}
                  >
                    {t('users.form.cancelPasswordReset')}
                  </button>
                </>
              )}
              <p className="form-hint">{t('users.form.passwordResetHint')}</p>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="role">
              {t('users.form.role')} {t('users.form.required')}
            </label>
            <select
              id="role"
              value={role}
              onChange={(e) => setRole(e.target.value as any)}
            >
              <option value="user">{t('users.roles.user')}</option>
              <option value="operator">{t('users.roles.operator')}</option>
              <option value="admin">{t('users.roles.admin')}</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="language">
              {t('users.form.language')} {t('users.form.required')}
            </label>
            <select
              id="language"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
            >
              <option value="en">{t('users.languages.en')}</option>
              <option value="ru">{t('users.languages.ru')}</option>
            </select>
          </div>

          {isEditMode && (
            <div className="form-group">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={isActive}
                  onChange={(e) => setIsActive(e.target.checked)}
                />
                <span>{t('users.form.isActive')}</span>
              </label>
            </div>
          )}
        </div>

        <div className="form-section">
          <h2>{t('users.form.organization')}</h2>

          <div className="form-group">
            <SearchableSelect
              id="department"
              label={t('users.form.department')}
              value={departmentId}
              onChange={setDepartmentId}
              options={departments.map(dept => ({
                value: dept.id,
                label: dept.path || dept.name,
              }))}
              placeholder={t('departments.form.selectParent')}
              emptyPlaceholder={t('departments.form.noParent')}
            />
          </div>

          <div className="form-group">
            <label>{t('users.form.groups')}</label>
            {groups.length === 0 ? (
              <p className="form-hint">{t('users.form.noGroupsAvailable')}</p>
            ) : (
              <div className="checkbox-list">
                {groups.map((group) => (
                  <label key={group.id} className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={selectedGroups.includes(group.id)}
                      onChange={() => toggleGroup(group.id)}
                    />
                    <span>
                      {group.name}
                      {group.description && (
                        <small> - {group.description}</small>
                      )}
                    </span>
                  </label>
                ))}
              </div>
            )}
          </div>
        </div>

        <div className="form-section">
          <h2>{t('users.form.permissions')}</h2>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_schedule_meetings}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_schedule_meetings: e.target.checked
                })}
              />
              <span>{t('users.form.canScheduleMeetings')}</span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_manage_department}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_manage_department: e.target.checked
                })}
              />
              <span>{t('users.form.canManageDepartment')}</span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-label">
              <input
                type="checkbox"
                checked={permissions.can_approve_recordings}
                onChange={(e) => setPermissions({
                  ...permissions,
                  can_approve_recordings: e.target.checked
                })}
              />
              <span>{t('users.form.canApproveRecordings')}</span>
            </label>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" onClick={() => window.location.href = '/users'} className="btn btn-secondary">
            {t('users.form.cancel')}
          </button>
          <button type="submit" disabled={loading} className="btn btn-primary">
            {loading ? t('users.form.saving') : (isEditMode ? t('users.form.update') : t('users.form.save'))}
          </button>
        </div>
      </form>
    </div>
  );
};

export default UserForm;
