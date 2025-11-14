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
  LuLayers,
  LuRotateCcw,
  LuInfo,
  LuPlus,
  LuArrowLeft,
  LuPencil,
  LuTrash2,
  LuPlay,
  LuCircle,
  LuFilm,
  LuFileText,
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
        <button onClick={handleBackToList} className="btn btn-ghost meeting-details-back">
          <LuArrowLeft /> {t('meetings.backToList')}
        </button>

        <div className="meeting-details-card">
          <div className="meeting-details-hero">
            <div>
              <h1>{selectedMeeting.title}</h1>
              <div className="hero-chips">
                <span className="chip chip-type">{t(`meetings.type.${selectedMeeting.type}`)}</span>
                {selectedMeeting.subject && (
                  <span className="chip chip-subject">
                    <LuBookOpen /> {selectedMeeting.subject.name}
                  </span>
                )}
                <span className="chip chip-recurrence">
                  <LuRotateCcw /> {t(`meetings.recurrence.${selectedMeeting.is_permanent ? 'permanent' : selectedMeeting.recurrence}`)}
                </span>
              </div>
            </div>
            <div className="hero-status">
              {selectedMeeting.is_permanent && (
                <span className="permanent-badge">
                  <LuClock /> {t('meetings.recurrence.permanent')}
                </span>
              )}
              <span className={`status-badge ${selectedMeeting.is_permanent ? 'permanent' : getMeetingStatusInfo(selectedMeeting.status).className}`}>
                {selectedMeeting.is_permanent ? t('meetings.recurrence.permanent') : t(`meetings.status.${selectedMeeting.status}`)}
              </span>
            </div>
          </div>

          <div className="meeting-details-grid">
            <div className="detail-block">
              <LuCalendar className="detail-icon" />
              <div>
                <span className="detail-label">{t('meetings.details.scheduled')}</span>
                <div className="detail-value">{formatMeetingDate(selectedMeeting.scheduled_at)}</div>
              </div>
            </div>
            <div className="detail-block">
              <LuClock className="detail-icon" />
              <div>
                <span className="detail-label">{t('meetings.details.duration')}</span>
                <div className="detail-value">{formatDuration(selectedMeeting.duration)}</div>
              </div>
            </div>
            <div className="detail-block">
              <LuVideo className="detail-icon" />
              <div>
                <span className="detail-label">{t('meetings.details.videoRecording')}</span>
                <div className="detail-value">
                  {selectedMeeting.needs_video_record ? t('common.yes') : t('common.no')}
                </div>
              </div>
            </div>
            <div className="detail-block">
              <LuMic className="detail-icon" />
              <div>
                <span className="detail-label">{t('meetings.details.audioRecording')}</span>
                <div className="detail-value">
                  {selectedMeeting.needs_audio_record ? t('common.yes') : t('common.no')}
                </div>
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
            <div className="section-heading">
              <h3>{t('meetings.details.participants')}</h3>
              <span className="section-count">{selectedMeeting.participants.length}</span>
            </div>
            <div className="participants-grid">
              {selectedMeeting.participants.map((participant) => {
                const displayName = participant.user ?
                  (participant.user.first_name && participant.user.last_name ?
                    `${participant.user.first_name} ${participant.user.last_name}` :
                    participant.user.username) :
                  participant.user_id;
                return (
                  <div key={participant.user_id} className="participant-card">
                    <div className="participant-avatar">
                      {displayName.charAt(0).toUpperCase()}
                    </div>
                    <div className="participant-card-body">
                      <span className="participant-name">{displayName}</span>
                      <span className="participant-email">{participant.user?.email || '—'}</span>
                    </div>
                    <div className="participant-tags">
                      <span className={`role-badge role-${participant.role}`}>
                        {participant.role === 'speaker' ? <LuMic /> : <LuUsers />}
                        {participant.role === 'speaker' ? t('meetings.type.presentation') : t('meetings.details.participants')}
                      </span>
                      <span className={`status-badge ${participant.status}`}>
                        {t(`meetings.participantStatus.${participant.status}`)}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {selectedMeeting.departments.length > 0 && (
            <div className="details-section">
              <div className="section-heading">
                <h3>{t('meetings.details.departments')}</h3>
                <span className="section-count">{selectedMeeting.departments.length}</span>
              </div>
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
            <div className="meeting-details-actions">
              {(selectedMeeting.status !== 'cancelled' && (selectedMeeting.status !== 'completed' || selectedMeeting.is_permanent || selectedMeeting.recurrence === 'permanent')) && (
                <button className="btn btn-join" onClick={handleJoinMeeting}>
                  <LuPlay /> {t('meetings.details.joinMeeting')}
                </button>
              )}
              {(isMeetingPast(selectedMeeting) || (selectedMeeting.is_permanent && selectedMeeting.status === 'completed')) && (
                <button className="btn btn-secondary" onClick={handleViewRecordings}>
                  <LuFilm /> {t('meetings.card.viewRecordings')}
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
              const totalParticipants = meeting.participants.length;
              const onlineParticipants = meeting.active_participants_count || 0;
              const departmentsCount = meeting.departments ? meeting.departments.length : 0;
              const recurrenceKey = meeting.is_permanent ? 'permanent' : (meeting.recurrence || 'none');
              const recurrenceLabel = t(`meetings.recurrence.${recurrenceKey}`);

              return (
                <div
                  key={meeting.id}
                  className={`meeting-card ${isNow ? 'meeting-now' : ''} ${isPast ? 'meeting-past' : ''} ${meeting.is_permanent ? 'meeting-card-permanent' : ''}`}
                  onClick={() => handleViewDetails(meeting)}
                >
                  <div className="meeting-card-header">
                    <div className="meeting-card-title">
                      <h3 className="meeting-title">{meeting.title}</h3>
                      <div className="meeting-card-chips">
                        <span className="chip chip-type">
                          {t(`meetings.type.${meeting.type}`)}
                        </span>
                        {meeting.subject && (
                          <span className="chip chip-subject">
                            <LuBookOpen /> {meeting.subject.name}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="meeting-card-status">
                      {meeting.is_permanent && (
                        <span className="permanent-badge" title={t('meetings.permanentMeetingTooltip')}>
                          <LuClock /> {t('meetings.recurrence.permanent')}
                        </span>
                      )}
                      <span className={`status-badge ${meeting.is_permanent ? 'permanent' : statusInfo.className}`}>
                        {meeting.is_permanent ? t('meetings.recurrence.permanent') : t(`meetings.status.${meeting.status}`)}
                      </span>
                      <span className="recurrence-pill">
                        <LuRotateCcw /> {recurrenceLabel}
                      </span>
                    </div>
                  </div>

                  <div className="meeting-card-details-grid">
                    <div className="detail-block">
                      <LuCalendar className="detail-icon" />
                      <div>
                        <span className="detail-label">{t('meetings.card.scheduleLabel')}</span>
                        <div className="detail-value">{formatMeetingDate(meeting.scheduled_at)}</div>
                      </div>
                    </div>
                    <div className="detail-block">
                      <LuClock className="detail-icon" />
                      <div>
                        <span className="detail-label">{t('meetings.card.durationLabel')}</span>
                        <div className="detail-value">{formatDuration(meeting.duration)}</div>
                      </div>
                    </div>
                    <div className="detail-block">
                      <LuRotateCcw className="detail-icon" />
                      <div>
                        <span className="detail-label">{t('meetings.card.recurrenceLabel')}</span>
                        <div className="detail-value">{recurrenceLabel}</div>
                      </div>
                    </div>
                    <div className="detail-block">
                      <LuBookOpen className="detail-icon" />
                      <div>
                        <span className="detail-label">{t('meetings.card.subjectLabel')}</span>
                        <div className="detail-value">{meeting.subject?.name || t('common.loading')}</div>
                      </div>
                    </div>
                  </div>

                  <div className="meeting-card-stats">
                    <div className="stat-block">
                      <LuUsers className="stat-icon" />
                      <div>
                        <span className="detail-label">{t('meetings.card.participantsLabel')}</span>
                        <div className="detail-value">
                          {t('meetings.card.participantsSummary', {
                            total: totalParticipants,
                            online: onlineParticipants,
                          })}
                        </div>
                      </div>
                    </div>
                    {departmentsCount > 0 && (
                      <div className="stat-block">
                        <LuLayers className="stat-icon" />
                        <div>
                          <span className="detail-label">{t('meetings.details.departments')}</span>
                          <div className="detail-value">
                            {t('meetings.card.departmentsSummary', { count: departmentsCount })}
                          </div>
                        </div>
                      </div>
                    )}
                    {meeting.needs_video_record && (
                      <div className="stat-block">
                        <LuVideo className="stat-icon" />
                        <div>
                          <span className="detail-label">{t('meetings.card.videoRecording')}</span>
                          <div className="detail-value">{t('common.yes')}</div>
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="meeting-card-flags">
                    {meeting.is_recording && (
                      <span className="flag flag-recording">
                        <LuCircle /> {t('meetings.card.recording')}
                      </span>
                    )}
                    {meeting.is_transcribing && (
                      <span className="flag flag-transcription">
                        <LuFileText /> {t('meetings.card.transcription')}
                      </span>
                    )}
                    {isNow && (
                      <span className="flag flag-live">
                        <LuCircle className="pulse-icon" /> {t('meetings.meetingInProgress')}
                      </span>
                    )}
                  </div>

                  <div className="meeting-card-footer">
                    <div className="participants-preview">
                      {meeting.participants.slice(0, 3).map((participant) => {
                        const displayName = participant.user ?
                          (participant.user.first_name && participant.user.last_name ?
                            `${participant.user.first_name} ${participant.user.last_name}` :
                            participant.user.username) :
                          'U';
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
                      {onlineParticipants > 0 && (
                        <div className="online-indicator">
                          <LuCircle className="pulse-icon-small" />
                          <span>{t('meetings.card.onlineCount', { count: onlineParticipants })}</span>
                        </div>
                      )}
                    </div>
                    <div className="meeting-card-actions">
                      {isPast ? (
                        <button
                          className="btn-recordings"
                          onClick={(e) => {
                            e.stopPropagation();
                            navigate(`/meeting/${meeting.id}/recordings`);
                          }}
                          title={t('meetings.card.viewRecordings')}
                        >
                          <LuFilm />
                        </button>
                      ) : meeting.status !== 'cancelled' && (
                        <button
                          className="btn btn-join btn-sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            window.location.href = `/meeting/${meeting.id}`;
                          }}
                          title={t('meetings.details.joinMeeting')}
                        >
                          <LuPlay /> {t('meetings.details.joinMeeting')}
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
