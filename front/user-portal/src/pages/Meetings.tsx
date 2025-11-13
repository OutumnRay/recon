import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import {
  LuCalendar,
  LuClock,
  LuUsers,
  LuVideo,
  LuMic,
  LuBookOpen,
  LuRotateCcw,
  LuInfo,
  LuPlus,
  LuArrowLeft,
  LuPencil,
  LuTrash2,
  LuPlay,
  LuCircle,
  LuCheck,
  LuX,
  LuFilm,
} from 'react-icons/lu';
import {
  listMyMeetings,
  formatMeetingDate,
  formatDuration,
  getMeetingStatusInfo,
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
import { DateTimePicker } from '../components/DateTimePicker';
import './Meetings.css';

export const Meetings: React.FC = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  // Set page title
  useEffect(() => {
    document.title = `Recontext - ${t('nav.meetings')}`;
  }, [t]);

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

      // Convert local datetime-local values to UTC ISO format for API
      const dateFromUTC = dateFromFilter ? new Date(dateFromFilter).toISOString() : undefined;
      const dateToUTC = dateToFilter ? new Date(dateToFilter).toISOString() : undefined;

      const response = await listMyMeetings({
        page,
        page_size: pageSize,
        status: statusFilter || undefined,
        type: typeFilter || undefined,
        subject_id: subjectFilter || undefined,
        date_from: dateFromUTC,
        date_to: dateToUTC,
      });

      setMeetings(response.items || []);
      setTotal(response.total || 0);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('meetings.errors.loadFailed'));
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

  const handleJoinMeeting = () => {
    if (!selectedMeeting) return;
    // Navigate to meeting room page - token will be fetched there
    window.location.href = `/meeting/${selectedMeeting.id}`;
  };

  const handleViewRecordings = () => {
    if (!selectedMeeting) return;
    navigate(`/meeting/${selectedMeeting.id}/recordings`);
  };

  const handleDeleteMeeting = async () => {
    if (!selectedMeeting) return;

    if (!window.confirm(t('meetings.confirmDelete', { title: selectedMeeting.title }))) {
      return;
    }

    try {
      await deleteMeeting(selectedMeeting.id);
      setViewMode('list');
      setSelectedMeeting(null);
      fetchMeetings();
    } catch (err) {
      alert(err instanceof Error ? err.message : t('meetings.errors.deleteFailed'));
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
          <button onClick={handleBackToList} className="btn btn-ghost">
            <LuArrowLeft /> {t('meetings.backToList')}
          </button>
          <h1 className="page-title">{selectedMeeting.title}</h1>
        </div>

        <div className="meeting-details-card">
          <div className="details-section">
            <h3>{t('meetings.details.information')}</h3>
            <div className="details-grid">
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.type')}:</span>
                <span className="detail-value">
                  {t(`meetings.type.${selectedMeeting.type}`)}
                </span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.status')}:</span>
                <span className={`status-badge ${getMeetingStatusInfo(selectedMeeting.status).className}`}>
                  {t(`meetings.status.${selectedMeeting.status}`)}
                </span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.scheduled')}:</span>
                <span className="detail-value">{formatMeetingDate(selectedMeeting.scheduled_at)}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.duration')}:</span>
                <span className="detail-value">{formatDuration(selectedMeeting.duration)}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.subject')}:</span>
                <span className="detail-value">{selectedMeeting.subject?.name || 'N/A'}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.recurrence')}:</span>
                <span className="detail-value">{t(`meetings.recurrence.${selectedMeeting.recurrence}`)}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.videoRecording')}:</span>
                <span className="detail-value">
                  {selectedMeeting.needs_video_record ? (
                    <><LuCheck className="icon-success" /> {t('common.yes')}</>
                  ) : (
                    <><LuX className="icon-muted" /> {t('common.no')}</>
                  )}
                </span>
              </div>
              <div className="detail-item">
                <span className="detail-label">{t('meetings.details.audioRecording')}:</span>
                <span className="detail-value">
                  {selectedMeeting.needs_audio_record ? (
                    <><LuCheck className="icon-success" /> {t('common.yes')}</>
                  ) : (
                    <><LuX className="icon-muted" /> {t('common.no')}</>
                  )}
                </span>
              </div>
            </div>
          </div>

          {selectedMeeting.additional_notes && (
            <div className="details-section">
              <h3>{t('meetings.details.additionalNotes')}</h3>
              <p className="notes-text">{selectedMeeting.additional_notes}</p>
            </div>
          )}

          <div className="details-section">
            <h3>{t('meetings.details.participants')} ({selectedMeeting.participants.length})</h3>
            <div className="participants-list">
              {selectedMeeting.participants.map((participant) => (
                <div key={participant.user_id} className="participant-item">
                  <div className="participant-info">
                    <span className="participant-name">
                      {participant.user?.full_name || participant.user?.username || participant.user_id}
                    </span>
                    <span className="participant-email">{participant.user?.email || ''}</span>
                  </div>
                  <div className="participant-badges">
                    <span className={`role-badge role-${participant.role}`}>
                      {participant.role === 'speaker' ? (
                        <><LuMic /> {t('meetings.type.presentation')}</>
                      ) : (
                        <><LuUsers /> {t('meetings.details.participants')}</>
                      )}
                    </span>
                    <span className={`status-badge ${participant.status}`}>
                      {t(`meetings.participantStatus.${participant.status}`)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {selectedMeeting.departments.length > 0 && (
            <div className="details-section">
              <h3>{t('meetings.details.departments')} ({selectedMeeting.departments.length})</h3>
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
            <h3>{t('meetings.details.actions')}</h3>
            <div className="meeting-actions">
              {selectedMeeting.status !== 'cancelled' && (
                <button className="btn btn-primary" onClick={handleJoinMeeting}>
                  <LuPlay /> {t('meetings.details.joinMeeting')}
                </button>
              )}
              {isMeetingPast(selectedMeeting) && (
                <button className="btn btn-secondary" onClick={handleViewRecordings}>
                  <LuFilm /> Записи встречи
                </button>
              )}
              {isMeetingUpcoming(selectedMeeting) && (
                <button className="btn btn-secondary">
                  <LuCalendar /> {t('meetings.details.addToCalendar')}
                </button>
              )}
              <button className="btn btn-secondary" onClick={handleEditMeeting}>
                <LuPencil /> {t('meetings.details.editMeeting')}
              </button>
              <button className="btn btn-danger" onClick={handleDeleteMeeting}>
                <LuTrash2 /> {t('meetings.details.cancelMeeting')}
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
      {/* Filters Section */}
      <div className="filters-section">
        <div className="filters-grid">
          <div className="filter-group">
            <label htmlFor="status-filter">{t('meetings.filters.status')}</label>
            <select
              id="status-filter"
              value={statusFilter}
              onChange={(e) => {
                setStatusFilter(e.target.value as MeetingStatus | '');
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">{t('meetings.filters.allStatuses')}</option>
              <option value="scheduled">{t('meetings.status.scheduled')}</option>
              <option value="in_progress">{t('meetings.status.in_progress')}</option>
              <option value="completed">{t('meetings.status.completed')}</option>
              <option value="cancelled">{t('meetings.status.cancelled')}</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="type-filter">{t('meetings.filters.type')}</label>
            <select
              id="type-filter"
              value={typeFilter}
              onChange={(e) => {
                setTypeFilter(e.target.value as MeetingType | '');
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">{t('meetings.filters.allTypes')}</option>
              <option value="presentation">{t('meetings.type.presentation')}</option>
              <option value="conference">{t('meetings.type.conference')}</option>
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="subject-filter">{t('meetings.filters.subject')}</label>
            <select
              id="subject-filter"
              value={subjectFilter}
              onChange={(e) => {
                setSubjectFilter(e.target.value);
                setPage(1);
              }}
              className="filter-select"
            >
              <option value="">{t('meetings.filters.allSubjects')}</option>
              {subjects.map((subject) => (
                <option key={subject.id} value={subject.id}>
                  {subject.name}
                </option>
              ))}
            </select>
          </div>

          <div className="filter-group">
            <label htmlFor="date-from-filter">{t('meetings.filters.dateFrom')}</label>
            <DateTimePicker
              id="date-from-filter"
              type="datetime-local"
              value={dateFromFilter}
              onChange={(value) => {
                setDateFromFilter(value);
                setPage(1);
              }}
            />
          </div>

          <div className="filter-group">
            <label htmlFor="date-to-filter">{t('meetings.filters.dateTo')}</label>
            <DateTimePicker
              id="date-to-filter"
              type="datetime-local"
              value={dateToFilter}
              onChange={(value) => {
                setDateToFilter(value);
                setPage(1);
              }}
            />
          </div>
        </div>

        {hasActiveFilters && (
          <button onClick={handleResetFilters} className="btn btn-ghost">
            <LuRotateCcw /> {t('meetings.filters.resetFilters')}
          </button>
        )}
      </div>

      {/* Error Message */}
      {error && (
        <div className="alert alert-error">
          <LuInfo />
          <span>{error}</span>
        </div>
      )}

      {/* Loading State */}
      {loading && meetings.length === 0 && (
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('meetings.loadingMeetings')}</p>
        </div>
      )}

      {/* Empty State */}
      {!loading && meetings.length === 0 && !error && (
        <div className="empty-state">
          <h2 className="empty-title">{t('meetings.noMeetings')}</h2>
          <p className="empty-description">
            {hasActiveFilters
              ? t('meetings.noMeetingsFiltered')
              : t('meetings.noMeetingsDescription')}
          </p>
          <button className="btn btn-primary" onClick={handleCreateMeeting}>
            <LuPlus /> {t('meetings.createNew')}
          </button>
        </div>
      )}

      {/* Meetings List */}
      {meetings.length > 0 && (
        <>
          <div className="meetings-list">
            {meetings.map((meeting) => {
              const statusInfo = getMeetingStatusInfo(meeting.status);
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
                          {t(`meetings.type.${meeting.type}`)}
                        </span>
                        {meeting.subject && (
                          <span className="meeting-subject">
                            <LuBookOpen /> {meeting.subject.name}
                          </span>
                        )}
                      </div>
                    </div>
                    <span className={`status-badge ${statusInfo.className}`}>
                      {t(`meetings.status.${meeting.status}`)}
                    </span>
                  </div>

                  <div className="meeting-card-body">
                    <div className="meeting-info-grid">
                      <div className="info-item">
                        <LuCalendar className="info-icon" />
                        <span className="info-text">{formatMeetingDate(meeting.scheduled_at)}</span>
                      </div>
                      <div className="info-item">
                        <LuClock className="info-icon" />
                        <span className="info-text">{formatDuration(meeting.duration)}</span>
                      </div>
                      <div className="info-item">
                        <LuUsers className="info-icon" />
                        <span className="info-text">{t('meetings.card.participants', { count: meeting.participants.length })}</span>
                      </div>
                      {meeting.needs_video_record && (
                        <div className="info-item">
                          <LuVideo className="info-icon" />
                          <span className="info-text">{t('meetings.card.videoRecording')}</span>
                        </div>
                      )}
                    </div>

                    {isNow && (
                      <div className="meeting-now-badge">
                        <LuCircle className="pulse-icon" /> {t('meetings.meetingInProgress')}
                      </div>
                    )}
                  </div>

                  <div className="meeting-card-footer">
                    <div className="participants-preview">
                      {meeting.participants.slice(0, 3).map((participant) => {
                        const displayName = participant.user?.full_name || participant.user?.username || 'U';
                        return (
                          <div key={participant.user_id} className="participant-avatar" title={displayName}>
                            {displayName.charAt(0).toUpperCase()}
                          </div>
                        );
                      })}
                      {meeting.participants.length > 3 && (
                        <div className="participant-avatar more">
                          +{meeting.participants.length - 3}
                        </div>
                      )}
                    </div>
                    <div className="meeting-card-actions">
                      {isPast && (
                        <button
                          className="btn-recordings"
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/meeting/${meeting.id}/recordings`);
                          }}
                          title="Просмотр записей"
                        >
                          <LuFilm />
                        </button>
                      )}
                      <button className="view-details-btn">
                        {t('meetings.viewDetails')} →
                      </button>
                    </div>
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
                {t('meetings.pagination.previous')}
              </button>
              <span className="pagination-info">
                {t('meetings.pagination.info', { page, total: totalPages, count: total })}
              </span>
              <button
                onClick={() => setPage(page + 1)}
                disabled={page >= totalPages}
                className="pagination-btn"
              >
                {t('meetings.pagination.next')}
              </button>
            </div>
          )}
        </>
      )}

      {/* Floating Action Button */}
      <button className="fab" title={t('meetings.createNew')} onClick={handleCreateMeeting}>
        <LuPlus />
      </button>
    </div>
  );
};

export default Meetings;
