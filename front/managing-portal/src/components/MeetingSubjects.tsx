import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { meetingSubjectsApi } from '../services/meetingSubjects';
import type { MeetingSubject } from '../services/meetingSubjects';
import './MeetingSubjects.css';

export const MeetingSubjects: React.FC = (): ReactElement => {
  const { t } = useTranslation();
  const [subjects, setSubjects] = useState<MeetingSubject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchSubjects();
  }, []);

  const fetchSubjects = async () => {
    try {
      const response = await meetingSubjectsApi.getSubjects(1, 100);
      setSubjects(response.items || []);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('subjects.errors.loadFailed'));
    } finally {
      setLoading(false);
    }
  };


  const handleDeleteSubject = async (id: string) => {
    if (!confirm(t('subjects.confirmDelete'))) return;

    try {
      await meetingSubjectsApi.deleteSubject(id);
      fetchSubjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('subjects.errors.deleteFailed'));
    }
  };


  if (loading) {
    return (
      <div className="page-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('subjects.loading')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h1 className="page-title">{t('subjects.title')}</h1>
        <button onClick={() => window.location.href = '/subjects/new'} className="btn btn-primary">
          + {t('subjects.addSubject')}
        </button>
      </div>

      {error && (
        <div className="error-message">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {subjects.length === 0 && !loading && (
        <div className="empty-state">
          <p>{t('subjects.noSubjects')}</p>
        </div>
      )}

      {subjects.length > 0 && (
        <div className="table-container">
          <table className="data-table">
            <thead>
              <tr>
                <th>{t('subjects.table.name')}</th>
                <th>{t('subjects.table.description')}</th>
                <th>{t('subjects.table.status')}</th>
                <th>{t('subjects.table.actions')}</th>
              </tr>
            </thead>
            <tbody>
              {subjects.map((subject) => (
                <tr key={subject.id}>
                  <td>{subject.name}</td>
                  <td>{subject.description || '-'}</td>
                  <td>
                    <span className={`status-badge ${subject.is_active ? 'active' : 'inactive'}`}>
                      {subject.is_active ? t('subjects.status.active') : t('subjects.status.inactive')}
                    </span>
                  </td>
                  <td className="actions-cell">
                    <button
                      onClick={() => window.location.href = `/subjects/${subject.id}/edit`}
                      className="btn btn-sm btn-secondary"
                    >
                      {t('subjects.table.edit')}
                    </button>
                    <button
                      onClick={() => handleDeleteSubject(subject.id)}
                      className="btn btn-sm btn-danger"
                    >
                      {t('subjects.deleteSubject')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
};

export default MeetingSubjects;
