import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  listMyMeetings,
  formatMeetingDate,
  formatDuration,
  getMeetingStatusInfo,
  getMeetingTypeInfo,
  isMeetingUpcoming,
  isMeetingNow,
  isMeetingPast,
  listMeetingSubjects,
  deleteMeeting,
} from '../services/meetings';
import type {
  MeetingWithDetails,
  MeetingStatus,
  MeetingType,
  MeetingSubject,
} from '../types/meeting';
import MeetingForm from '../components/MeetingForm';
import './Meetings.css';

export const Meetings: React.FC = () => {
  const { t } = useTranslation();

  // State for meetings list
  const [meetings, setMeetings] = useState<MeetingWithDetails[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Pagination state
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [total, setTotal] = useState(0);

  // Filter state
  const [statusFilter, setStatusFilter] = useState<MeetingStatus | ''>('');
  const [typeFilter, setTypeFilter] = useState<MeetingType | ''>('');
  const [subjectFilter, setSubjectFilter] = useState<string>('');
  const [dateFromFilter, setDateFromFilter] = useState<string>('');
  const [dateToFilter, setDateToFilter] = useState<string>('');

  // Subjects for dropdown
  const [subjects, setSubjects] = useState<MeetingSubject[]>([]);

  // Selected meeting for details view
  const [selectedMeeting, setSelectedMeeting] = useState<MeetingWithDetails | null>(null);

  // View mode: list, details, create, or edit
  const [viewMode, setViewMode] = useState<'list' | 'details' | 'create' | 'edit'>('list');

  // Fetch subjects on mount
  useEffect(() => {
    fetchSubjects();
  }, []);

  // Fetch meetings when filters or page changes
  useEffect(() => {
    fetchMeetings();
  }, [page, statusFilter, typeFilter, subjectFilter, dateFromFilter, dateToFilter]);

  const fetchSubjects = async () => {
    try {
      const response = await listMeetingSubjects(1, 100);
      setSubjects(response.items);
    } catch (err) {
      console.error('Failed to fetch subjects:', err);
    }
  };

  const fetchMeetings = async () => {
    try {
      setLoading(true);
      const response = await listMyMeetings({
        page,
        page_size: pageSize,
        status: statusFilter || undefined,
        type: typeFilter || undefined,
        subject_id: subjectFilter || undefined,
        date_from: dateFromFilter || undefined,
        date_to: dateToFilter || undefined,
      });

      setMeetings(response.items || []);
      setTotal(response.total || 0);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load meetings');
      setMeetings([]);
    } finally {
      setLoading(false);
    }
  };

  const handleResetFilters = () => {
    setStatusFilter('');
    setTypeFilter('');
    setSubjectFilter('');
    setDateFromFilter('');
    setDateToFilter('');
    setPage(1);
  };

  const handleViewDetails = (meeting: MeetingWithDetails) => {
    setSelectedMeeting(meeting);
    setViewMode('details');
  };

  const handleBackToList = () => {
    setSelectedMeeting(null);
    setViewMode('list');
    fetchMeetings();
  };

  const handleCreateMeeting = () => {
    setSelectedMeeting(null);
    setViewMode('create');
  };

  const handleEditMeeting = () => {
    setViewMode('edit');
  };

  const handleFormSuccess = (meeting: MeetingWithDetails) => {
    setSelectedMeeting(meeting);
    setViewMode('details');
    fetchMeetings();
  };

  const handleFormCancel = () => {
    if (selectedMeeting) {
      setViewMode('details');
    } else {
      setViewMode('list');
    }
  };

  const handleDeleteMeeting = async () => {
    if (!selectedMeeting) return;

    if (!window.confirm(`Are you sure you want to cancel "${selectedMeeting.title}"?`)) {
      return;
    }

    try {
      await deleteMeeting(selectedMeeting.id);
      setViewMode('list');
      setSelectedMeeting(null);
      fetchMeetings();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to delete meeting');
    }
  };

  const totalPages = Math.ceil(total / pageSize);
  const hasActiveFilters = statusFilter || typeFilter || subjectFilter || dateFromFilter || dateToFilter;

  // Render create form
  if (viewMode === 'create') {
    return (
      <div className="page-container">
        <MeetingForm
          onSuccess={handleFormSuccess}
          onCancel={handleFormCancel}
        />
      </div>
    );
  }

  // Render edit form
  if (viewMode === 'edit' && selectedMeeting) {
    return (
      <div className="page-container">
        <MeetingForm
          meeting={selectedMeeting}
          onSuccess={handleFormSuccess}
          onCancel={handleFormCancel}
        />
      </div>
    );
  }

  // Render details view
  if (viewMode === 'details' && selectedMeeting) {
    return (
      <div className="page-container">
        <div className="meeting-details-header">
          <button onClick={handleBackToList} className="back-btn">
            ← Back to Meetings
          </button>
          <h1 className="page-title">{selectedMeeting.title}</h1>
        </div>

        <div className="meeting-details-card">
          <div className="details-section">
            <h3>Meeting Information</h3>
            <div className="details-grid">
              <div className="detail-item">
                <span className="detail-label">Type:</span>
                <span className="detail-value">
                  {getMeetingTypeInfo(selectedMeeting.type).icon}{' '}
                  {getMeetingTypeInfo(selectedMeeting.type).label}
                </span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Status:</span>
                <span className={`status-badge ${getMeetingStatusInfo(selectedMeeting.status).className}`}>
                  {getMeetingStatusInfo(selectedMeeting.status).label}
                </span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Scheduled:</span>
                <span className="detail-value">{formatMeetingDate(selectedMeeting.scheduled_at)}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Duration:</span>
                <span className="detail-value">{formatDuration(selectedMeeting.duration)}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Subject:</span>
                <span className="detail-value">{selectedMeeting.subject?.name || 'N/A'}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Recurrence:</span>
                <span className="detail-value">{selectedMeeting.recurrence}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Video Recording:</span>
                <span className="detail-value">{selectedMeeting.needs_video_record ? '✓ Yes' : '✗ No'}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Audio Recording:</span>
                <span className="detail-value">{selectedMeeting.needs_audio_record ? '✓ Yes' : '✗ No'}</span>
              </div>
            </div>
          </div>

          {selectedMeeting.additional_notes && (
            <div className="details-section">
              <h3>Additional Notes</h3>
              <p className="notes-text">{selectedMeeting.additional_notes}</p>
            </div>
          )}

          <div className="details-section">
            <h3>Participants ({selectedMeeting.participants.length})</h3>
            <div className="participants-list">
              {selectedMeeting.participants.map((participant) => (
                <div key={participant.user_id} className="participant-item">
                  <div className="participant-info">
                    <span className="participant-name">{participant.user.full_name}</span>
                    <span className="participant-email">{participant.user.email}</span>
                  </div>
                  <div className="participant-badges">
                    <span className={`role-badge role-${participant.role}`}>
                      {participant.role === 'speaker' ? '🎤 Speaker' : '👤 Participant'}
                    </span>
                    <span className={`status-badge ${participant.status}`}>
                      {participant.status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {selectedMeeting.departments.length > 0 && (
            <div className="details-section">
              <h3>Departments ({selectedMeeting.departments.length})</h3>
              <div className="departments-list">
                {selectedMeeting.departments.map((dept) => (
                  <div key={dept.id} className="department-tag">
                    {dept.name}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="details-section">
            <h3>Meeting Actions</h3>
            <div className="meeting-actions">
              {isMeetingNow(selectedMeeting) && (
                <button className="action-btn join-btn">
                  🎥 Join Meeting
                </button>
              )}
              {isMeetingUpcoming(selectedMeeting) && (
                <button className="action-btn calendar-btn">
                  📅 Add to Calendar
                </button>
              )}
              <button className="action-btn edit-btn" onClick={handleEditMeeting}>
                ✏️ Edit Meeting
              </button>
              <button className="action-btn delete-btn" onClick={handleDeleteMeeting}>
                🗑️ Cancel Meeting
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Render list view
  return (
    <div className="page-container">
      <h1 className="page-title">{t('nav.meetings')}</h1>
      <p className="page-subtitle">View and manage your scheduled video meetings</p>

      {/* Filters Section */}
      <div className="filters-section">
        <div className="filters-grid">
          <div className="filter-group">
            <label htmlFor="status-filter">Status</label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value as MeetingStatus | '');
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">All Statuses</option>
              <option value="scheduled">Scheduled</option>
              <option value="in_progress">In Progress</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="type-filter">Type</label>
            <select
              id="type-filter"
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value as MeetingType | '');
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">All Types</option>
              <option value="presentation">🎤 Presentation</option>
              <option value="conference">👥 Conference</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="subject-filter">Subject</label>
            <select
              id="subject-filter"
              value={subjectFilter}
              onChange={(e) => {
                setSubjectFilter(e.target.value);
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">All Subjects</option>
              {subjects.map((subject) => (
                <option key={subject.id} value={subject.id}>
                  {subject.name}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="date-from-filter">From Date</label>
            <input
              id="date-from-filter"
              type="datetime-local"
              value={dateFromFilter}
              onChange={(e) => {
                setDateFromFilter(e.target.value);
                setPage(1);
              }}
              className="filter-input"
            />
          </div>

          <div className="filter-group">
            <label htmlFor="date-to-filter">To Date</label>
            <input
              id="date-to-filter"
              type="datetime-local"
              value={dateToFilter}
              onChange={(e) => {
                setDateToFilter(e.target.value);
                setPage(1);
              }}
              className="filter-input"
            />
          </div>
        </div>

        {hasActiveFilters && (
          <button onClick={handleResetFilters} className="reset-filters-btn">
            🔄 Reset Filters
          </button>
        )}
      </div>

      {/* Error Message */}
      {error && (
        <div className="error-message">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {/* Loading State */}
      {loading && meetings.length === 0 && (
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>Loading meetings...</p>
        </div>
      )}

      {/* Empty State */}
      {!loading && meetings.length === 0 && !error && (
        <div className="empty-state">
          <h2 className="empty-title">No meetings found</h2>
          <p className="empty-description">
            {hasActiveFilters
              ? 'Try adjusting your filters to see more meetings.'
              : 'You have no scheduled meetings yet. Create a new meeting to get started.'}
          </p>
          <button className="create-meeting-btn" onClick={handleCreateMeeting}>
            ➕ Create New Meeting
          </button>
        </div>
      )}

      {/* Meetings List */}
      {meetings.length > 0 && (
        <>
          <div className="meetings-list">
            {meetings.map((meeting) => {
              const statusInfo = getMeetingStatusInfo(meeting.status);
              const typeInfo = getMeetingTypeInfo(meeting.type);
              const isNow = isMeetingNow(meeting);
              const isPast = isMeetingPast(meeting);

              return (
                <div
                  key={meeting.id}
                  className={`meeting-card ${isNow ? 'meeting-now' : ''} ${isPast ? 'meeting-past' : ''}`}
                  onClick={() => handleViewDetails(meeting)}
                >
                  <div className="meeting-card-header">
                    <div className="meeting-title-section">
                      <h3 className="meeting-title">{meeting.title}</h3>
                      <div className="meeting-meta">
                        <span className="meeting-type">
                          {typeInfo.icon} {typeInfo.label}
                        </span>
                        {meeting.subject && (
                          <span className="meeting-subject">
                            📚 {meeting.subject.name}
                          </span>
                        )}
                      </div>
                    </div>
                    <span className={`status-badge ${statusInfo.className}`}>
                      {statusInfo.label}
                    </span>
                  </div>

                  <div className="meeting-card-body">
                    <div className="meeting-info-grid">
                      <div className="info-item">
                        <span className="info-icon">📅</span>
                        <span className="info-text">{formatMeetingDate(meeting.scheduled_at)}</span>
                      </div>
                      <div className="info-item">
                        <span className="info-icon">⏱️</span>
                        <span className="info-text">{formatDuration(meeting.duration)}</span>
                      </div>
                      <div className="info-item">
                        <span className="info-icon">👥</span>
                        <span className="info-text">{meeting.participants.length} participants</span>
                      </div>
                      {meeting.needs_video_record && (
                        <div className="info-item">
                          <span className="info-icon">🎥</span>
                          <span className="info-text">Video Recording</span>
                        </div>
                      )}
                    </div>

                    {isNow && (
                      <div className="meeting-now-badge">
                        🔴 Meeting in progress
                      </div>
                    )}
                  </div>

                  <div className="meeting-card-footer">
                    <div className="participants-preview">
                      {meeting.participants.slice(0, 3).map((participant) => (
                        <div key={participant.user_id} className="participant-avatar" title={participant.user.full_name}>
                          {participant.user.full_name.charAt(0).toUpperCase()}
                        </div>
                      ))}
                      {meeting.participants.length > 3 && (
                        <div className="participant-avatar more">
                          +{meeting.participants.length - 3}
                        </div>
                      )}
                    </div>
                    <button className="view-details-btn">
                      View Details →
                    </button>
                  </div>
                </div>
              );
            })}
          </div>

          {/* Pagination */}
          {totalPages > 1 && (
            <div className="pagination">
              <button
                onClick={() => setPage(page - 1)}
                disabled={page === 1}
                className="pagination-btn"
              >
                Previous
              </button>
              <span className="pagination-info">
                Page {page} of {totalPages} ({total} total meetings)
              </span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
                className="pagination-btn"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}

      {/* Floating Action Button */}
      <button className="fab" title="Create new meeting" onClick={handleCreateMeeting}>
        ➕
      </button>
    </div>
  );
};

export default Meetings;
