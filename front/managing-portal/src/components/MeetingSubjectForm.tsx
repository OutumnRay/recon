import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { meetingSubjectsApi } from '../services/meetingSubjects';
import type { MeetingSubject } from '../services/meetingSubjects';
import './UserForm.css';

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

  useEffect(() => {
    if (isEditMode && id) {
      fetchSubject(id);
    }
  }, [id]);

  const fetchSubject = async (subjectId: string) => {
    try {
      setLoading(true);
      const subject: MeetingSubject = await meetingSubjectsApi.getSubject(subjectId);
      setName(subject.name);
      setDescription(subject.description || '');
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
        });
      } else {
        // Create subject
        await meetingSubjectsApi.createSubject({
          name,
          description,
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
