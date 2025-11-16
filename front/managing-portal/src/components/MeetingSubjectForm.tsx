import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { meetingSubjectsApi } from '../services/meetingSubjects';
import type { MeetingSubject } from '../services/meetingSubjects';
import './UserForm.css';

interface Organization {
  id: string;
  name: string;
  description?: string;
}

export const MeetingSubjectForm: React.FC = () => {
  const { t } = useTranslation();

  // Extract subject ID from URL path
  const pathParts = window.location.pathname.split('/');
  const id = pathParts.includes('edit') ? pathParts[pathParts.length - 2] : null;
  const isEditMode = !!id;

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [organizationId, setOrganizationId] = useState('');
  const [organizations, setOrganizations] = useState<Organization[]>([]);

  useEffect(() => {
    loadOrganizations();
    if (isEditMode && id) {
      fetchSubject(id);
    }
  }, [id]);

  const loadOrganizations = async () => {
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/organizations', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setOrganizations(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch organizations:', err);
    }
  };

  const fetchSubject = async (subjectId: string) => {
    try {
      setLoading(true);
      const subject: MeetingSubject = await meetingSubjectsApi.getSubject(subjectId);
      setName(subject.name);
      setDescription(subject.description || '');
      setOrganizationId(subject.organization_id || '');
    } catch (err) {
      setError(err instanceof Error ? err.message : t('subjects.errors.loadFailed'));
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    try {
      setLoading(true);

      if (isEditMode) {
        // Update subject
        await meetingSubjectsApi.updateSubject(id!, {
          name,
          description,
          organization_id: organizationId || null,
        });
      } else {
        // Create subject
        await meetingSubjectsApi.createSubject({
          name,
          description,
          organization_id: organizationId || null,
        });
      }

      window.location.href = '/subjects';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  };

  if (loading && isEditMode) {
    return (
      <div className="user-form-page">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('subjects.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="user-form-page">
      <div className="form-header">
        <h1>{isEditMode ? t('subjects.form.editSubjectTitle', { name }) : t('subjects.addSubject')}</h1>
        <button onClick={() => window.location.href = '/subjects'} className="btn btn-secondary">
          {t('subjects.form.cancel')}
        </button>
      </div>

      {error && (
        <div className="error-banner">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="user-form">
        <div className="form-section">
          <h2>{t('subjects.form.subjectInfo')}</h2>

          <div className="form-group">
            <label htmlFor="name">{t('subjects.form.name')} *</label>
            <input
              type="text"
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              placeholder={t('subjects.form.namePlaceholder')}
            />
          </div>

          <div className="form-group">
            <label htmlFor="description">{t('subjects.form.description')}</label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={4}
              placeholder={t('subjects.form.descriptionPlaceholder')}
            />
          </div>

          <div className="form-group">
            <label htmlFor="organization">{t('subjects.form.organization', 'Organization')}</label>
            <select
              id="organization"
              value={organizationId}
              onChange={(e) => setOrganizationId(e.target.value)}
            >
              <option value="">{t('subjects.form.selectOrganization', 'Select an organization...')}</option>
              {organizations.map((org) => (
                <option key={org.id} value={org.id}>
                  {org.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-actions">
          <button type="button" onClick={() => window.location.href = '/subjects'} className="btn btn-secondary">
            {t('subjects.form.cancel')}
          </button>
          <button type="submit" disabled={loading} className="btn btn-primary">
            {loading ? t('subjects.form.saving') : (isEditMode ? t('subjects.form.update') : t('subjects.form.save'))}
          </button>
        </div>
      </form>
    </div>
  );
};

export default MeetingSubjectForm;
