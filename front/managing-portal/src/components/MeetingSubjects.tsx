import React, { useEffect, useState } from 'react';
import type { ReactElement } from 'react';
import { useTranslation } from 'react-i18next';
import { meetingSubjectsApi } from '../services/meetingSubjects';
import type { MeetingSubject, CreateMeetingSubjectRequest, UpdateMeetingSubjectRequest } from '../services/meetingSubjects';
import './MeetingSubjects.css';

export const MeetingSubjects: React.FC = (): ReactElement => {
  const { t } = useTranslation();
  const [subjects, setSubjects] = useState<MeetingSubject[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedSubject, setSelectedSubject] = useState<MeetingSubject | null>(null);
  const [formData, setFormData] = useState<CreateMeetingSubjectRequest>({
    name: '',
    description: ''
  });

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

  const handleAddSubject = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await meetingSubjectsApi.createSubject(formData);
      setShowAddModal(false);
      setFormData({ name: '', description: '' });
      fetchSubjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('subjects.errors.createFailed'));
    }
  };

  const handleUpdateSubject = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedSubject) return;

    try {
      const updateData: UpdateMeetingSubjectRequest = {
        name: formData.name,
        description: formData.description
      };
      await meetingSubjectsApi.updateSubject(selectedSubject.id, updateData);
      setShowEditModal(false);
      setSelectedSubject(null);
      setFormData({ name: '', description: '' });
      fetchSubjects();
    } catch (err) {
      setError(err instanceof Error ? err.message : t('subjects.errors.updateFailed'));
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

  const openEditModal = (subject: MeetingSubject) => {
    setSelectedSubject(subject);
    setFormData({
      name: subject.name,
      description: subject.description || ''
    });
    setShowEditModal(true);
  };

  const openAddModal = () => {
    setFormData({ name: '', description: '' });
    setShowAddModal(true);
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
        <button onClick={openAddModal} className="btn btn-primary">
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
                      onClick={() => openEditModal(subject)}
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

      {/* Add Subject Modal */}
      {showAddModal && (
        <div className="modal-overlay" onClick={() => setShowAddModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{t('subjects.addSubject')}</h2>
              <button onClick={() => setShowAddModal(false)} className="modal-close">×</button>
            </div>
            <form onSubmit={handleAddSubject} className="modal-form">
              <div className="form-group">
                <label>{t('subjects.form.name')}</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  className="form-input"
                />
              </div>
              <div className="form-group">
                <label>{t('subjects.form.description')}</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="form-textarea"
                  rows={3}
                />
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowAddModal(false)} className="btn btn-secondary">
                  {t('subjects.form.cancel')}
                </button>
                <button type="submit" className="btn btn-primary">
                  {t('subjects.form.save')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Subject Modal */}
      {showEditModal && selectedSubject && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>{t('subjects.form.editSubjectTitle', { name: selectedSubject.name })}</h2>
              <button onClick={() => setShowEditModal(false)} className="modal-close">×</button>
            </div>
            <form onSubmit={handleUpdateSubject} className="modal-form">
              <div className="form-group">
                <label>{t('subjects.form.name')}</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  required
                  className="form-input"
                />
              </div>
              <div className="form-group">
                <label>{t('subjects.form.description')}</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  className="form-textarea"
                  rows={3}
                />
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowEditModal(false)} className="btn btn-secondary">
                  {t('subjects.form.cancel')}
                </button>
                <button type="submit" className="btn btn-primary">
                  {t('subjects.form.update')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default MeetingSubjects;
