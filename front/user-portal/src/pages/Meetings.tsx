import React, { useState, useEffect, useRef } from 'react';
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
  LuArrowRight,
  LuPencil,
  LuTrash2,
  LuPlay,
  LuFilm,
  LuCopy,
  LuCheck,
  LuList,
  LuX,
} from 'react-icons/lu';
import {
  listMyMeetings,
  formatMeetingDate,
  formatDuration,
  getMeetingStatusInfo,
  isMeetingUpcoming,
  isMeetingNow,
  isMeetingPast,
  deleteMeeting,
  cancelMeeting,
} from '../services/meetings';
import type {
  MeetingWithDetails,
  MeetingStatus,
} from '../types/meeting';
import MeetingForm from '../components/MeetingForm';
import './Meetings.css';

export const Meetings: React.FC = () => {
  const { t, i18n } = useTranslation();
  const navigate = useNavigate();
  const locale = i18n.language?.startsWith('ru') ? 'ru-RU' : 'en-US';
  const formatListDateTime = (value: string) =>
    new Date(value).toLocaleString(locale, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    });

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
  const [statusFilter, setStatusFilter] = useState<MeetingStatus | 'permanent' | ''>('');
  // Selected meeting for details view
  const [selectedMeeting, setSelectedMeeting] = useState<MeetingWithDetails | null>(null);

  // Copy link state
  const [copiedMeetingId, setCopiedMeetingId] = useState<string | null>(null);

  // Current user state for permission checks
  const [currentUserId, setCurrentUserId] = useState<string | null>(null);
  const [currentUserRole, setCurrentUserRole] = useState<string | null>(null);

  // View mode: list, details, create, or edit
  const [viewMode, setViewMode] = useState<'list' | 'details' | 'create' | 'edit'>('list');
  const statusSliderRef = useRef<HTMLDivElement>(null);

  const statusOptions: Array<{
    value: MeetingStatus | 'permanent' | '';
    label: string;
    icon: React.ReactNode;
  }> = [
    { value: '', label: t('meetings.filters.allStatuses'), icon: <LuList /> },
    { value: 'scheduled', label: t('meetings.status.scheduled'), icon: <LuCalendar /> },
    { value: 'in_progress', label: t('meetings.status.in_progress'), icon: <LuPlay /> },
    { value: 'permanent', label: t('meetings.filters.permanent'), icon: <LuRotateCcw /> },
    { value: 'completed', label: t('meetings.status.completed'), icon: <LuCheck /> },
    { value: 'cancelled', label: t('meetings.status.cancelled'), icon: <LuTrash2 /> },
  ];

  // Load current user data for permission checks
  useEffect(() => {
    const userStr = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (userStr) {
      try {
        const user = JSON.parse(userStr);
        setCurrentUserId(user.id);
        setCurrentUserRole(user.role);
      } catch (err) {
        console.error('Failed to parse user data:', err);
      }
    }
  }, []);

  // Fetch meetings when filters or page changes
  useEffect(() => {
    fetchMeetings();
  }, [page, statusFilter]);

  const fetchMeetings = async () => {
    try {
      setLoading(true);

      // Map filter to backend status parameter
      let apiStatus: MeetingStatus | undefined = undefined;
      if (statusFilter && statusFilter !== 'permanent') {
        apiStatus = statusFilter as MeetingStatus;
      }

      const response = await listMyMeetings({
        page,
        page_size: pageSize,
        status: apiStatus,
      });

      let filteredItems = response.items || [];

      // Apply client-side filtering for special cases
      if (statusFilter === 'permanent') {
        // permanent - only permanent meetings
        // Backend already excludes cancelled meetings by default
        filteredItems = filteredItems.filter(
          m => (m.is_permanent || m.recurrence === 'permanent')
        );
      } else if (statusFilter === 'in_progress') {
        // go - only if there are active participants
        filteredItems = filteredItems.filter(
          m => m.active_participants_count > 0
        );
      } else if (statusFilter === 'completed') {
        // completed - all non-permanent completed (not cancelled)
        filteredItems = filteredItems.filter(
          m => m.status === 'completed' && !m.is_permanent && m.recurrence !== 'permanent'
        );
      }

      setMeetings(filteredItems);
      setTotal(filteredItems.length);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : t('meetings.errors.loadFailed'));
      setMeetings([]);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusSelect = (value: MeetingStatus | 'permanent' | '') => {
    const nextValue = statusFilter === value ? '' : value;
    setStatusFilter(nextValue);
    setPage(1);
  };

  const handleSlideStatuses = (direction: 'left' | 'right') => {
    if (!statusSliderRef.current) return;
    const offset = direction === 'left' ? -200 : 200;
    statusSliderRef.current.scrollBy({ left: offset, behavior: 'smooth' });
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

  const handleCopyAnonymousLink = async (meetingId: string, event: React.MouseEvent) => {
    event.stopPropagation(); // Prevent row click

    const portalUrl = window.location.origin;
    const anonymousLink = `${portalUrl}/meeting/${meetingId}/join`;

    try {
      await navigator.clipboard.writeText(anonymousLink);
      setCopiedMeetingId(meetingId);
      setTimeout(() => setCopiedMeetingId(null), 2000);
    } catch (err) {
      console.error('Failed to copy link:', err);
    }
  };

  const handleViewRecordingsFromRow = (meetingId: string, event: React.MouseEvent) => {
    event.stopPropagation(); // Prevent row click
    navigate(`/meeting/${meetingId}/recordings`);
  };

  const handleJoinFromRow = (meetingId: string, event: React.MouseEvent) => {
    event.stopPropagation();
    window.location.href = `/meeting/${meetingId}`;
  };

  const handleRowKeyDown = (event: React.KeyboardEvent, meeting: MeetingWithDetails) => {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      handleViewDetails(meeting);
    }
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

  /**
   * Check if current user can cancel a meeting
   * - Admins can cancel any meeting
   * - Organization admins can cancel any meeting
   * - Meeting creators can cancel their own meetings
   * - Meeting must not already be cancelled
   */
  const canCancelMeeting = (meeting: MeetingWithDetails): boolean => {
    if (meeting.status === 'cancelled') return false;
    if (currentUserRole === 'admin') return true;
    if (currentUserRole === 'organization_admin') return true;
    if (currentUserId === meeting.created_by) return true;
    return false;
  };

  /**
   * Cancel a meeting from the list view
   */
  const handleCancelMeetingFromList = async (meeting: MeetingWithDetails, event: React.MouseEvent) => {
    event.stopPropagation();

    if (!window.confirm(t('meetings.confirmCancel', { title: meeting.title }))) {
      return;
    }

    try {
      await cancelMeeting(meeting.id);
      fetchMeetings();
    } catch (err) {
      alert(err instanceof Error ? err.message : t('meetings.errors.cancelFailed'));
    }
  };

  const totalPages = Math.ceil(total / pageSize);
  const hasActiveFilters = Boolean(statusFilter);

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
              {selectedMeeting.is_permanent ? (
                <span className="permanent-badge">
                  <LuClock /> {t('meetings.recurrence.permanent')}
                </span>
              ) : (
                <span className={`status-badge ${getMeetingStatusInfo(selectedMeeting.status).className}`}>
                  {t(`meetings.status.${selectedMeeting.status}`)}
                </span>
              )}
            </div>
          </div>

          <div className="meeting-details-grid">
            {!selectedMeeting.is_permanent && (
              <div className="detail-block">
                <LuCalendar className="detail-icon" />
                <div>
                  <span className="detail-label">{t('meetings.details.scheduled')}</span>
                  <div className="detail-value">{formatMeetingDate(selectedMeeting.scheduled_at)}</div>
                </div>
              </div>
            )}
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
                <span className="detail-label">{t('meetings.details.recording')}</span>
                <div className="detail-value">
                  {selectedMeeting.needs_record ? t('common.yes') : t('common.no')}
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
                const avatarUrl = participant.user?.avatar;
                return (
                  <div key={participant.user_id} className="participant-card">
                    <div className="participant-avatar">
                      {avatarUrl ? (
                        <img src={avatarUrl} alt={displayName} />
                      ) : (
                        displayName.charAt(0).toUpperCase()
                      )}
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
      <div className="status-slider-container">
        <button
          type="button"
          className="status-arrow"
          onClick={() => handleSlideStatuses('left')}
          aria-label={t('meetings.pagination.previous')}
        >
          <LuArrowLeft />
        </button>
        <div className="status-slider" ref={statusSliderRef}>
          {statusOptions.map((option) => (
            <button
              type="button"
              key={option.value || 'all'}
              className={`status-pill ${statusFilter === option.value ? 'active' : ''}`}
              onClick={() => handleStatusSelect(option.value)}
              aria-pressed={statusFilter === option.value}
            >
              {option.icon}
              <span>{option.label}</span>
            </button>
          ))}
        </div>
        <button
          type="button"
          className="status-arrow"
          onClick={() => handleSlideStatuses('right')}
          aria-label={t('meetings.pagination.next')}
        >
          <LuArrowRight />
        </button>
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
              const hasRecording = meeting.needs_record;
              // Show recordings button if there are any recording rooms for this meeting
              const hasRecordings = (meeting.recordings_count || 0) > 0;
              const canJoinFromRow = meeting.is_permanent || meeting.status === 'scheduled' || meeting.status === 'in_progress';

              return (
                <div
                  key={meeting.id}
                  className={`meeting-row ${isNow ? 'meeting-row-now' : ''} ${isPast ? 'meeting-row-past' : ''}`}
                  onClick={() => handleViewDetails(meeting)}
                  role="button"
                  tabIndex={0}
                  onKeyDown={event => handleRowKeyDown(event, meeting)}
                >
                  <div className="meeting-row-info">
                    <div className="meeting-row-primary">
                      <div className="meeting-row-title">
                        <span className="meeting-row-name">{meeting.title}</span>
                        {meeting.subject && (
                          <span className="meeting-row-chip">
                            <LuBookOpen /> {meeting.subject.name}
                          </span>
                        )}
                        {meeting.is_permanent && (
                          <span className="meeting-row-chip subtle">
                            <LuClock /> {t('meetings.recurrence.permanent')}
                          </span>
                        )}
                      </div>
                      <div className="meeting-row-subtitle">
                        <span>{t(`meetings.type.${meeting.type}`)}</span>
                        {hasRecording && (
                          <span className="meeting-row-icons">
                            {meeting.needs_record && <LuVideo />}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="meeting-row-meta">
                      {!meeting.is_permanent && (
                        <span className="meeting-row-date">
                          <LuCalendar /> {formatListDateTime(meeting.scheduled_at)}
                        </span>
                      )}
                      <span className="meeting-row-duration">
                        <LuClock /> {`${meeting.duration} min`}
                      </span>
                      <span className="meeting-row-participants">
                        <LuUsers /> {meeting.participants.length}
                      </span>
                      {!meeting.is_permanent && (
                        <span className={`status-badge ${statusInfo.className}`}>
                          {t(`meetings.status.${meeting.status}`)}
                        </span>
                      )}
                    </div>
                  </div>
                  <div className="meeting-row-actions">
                    {meeting.allow_anonymous && (
                      <button
                        className="meeting-row-action-btn"
                        onClick={(e) => handleCopyAnonymousLink(meeting.id, e)}
                        title={t('meetings.card.copyLink')}
                      >
                        {copiedMeetingId === meeting.id ? <LuCheck /> : <LuCopy />}
                      </button>
                    )}
                    {hasRecordings && (
                      <button
                        className="meeting-row-action-btn"
                        onClick={(e) => handleViewRecordingsFromRow(meeting.id, e)}
                        title={t('meetings.card.viewRecordings')}
                      >
                        <LuFilm />
                      </button>
                    )}
                    {canCancelMeeting(meeting) && (
                      <button
                        className="meeting-row-action-btn meeting-row-action-danger"
                        onClick={(e) => handleCancelMeetingFromList(meeting, e)}
                        title={t('meetings.card.cancelMeeting')}
                      >
                        <LuX />
                      </button>
                    )}
                  </div>
                  {canJoinFromRow && (
                    <button
                      className="meeting-row-cta"
                      onClick={(e) => handleJoinFromRow(meeting.id, e)}
                    >
                      <LuArrowRight />
                    </button>
                  )}
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
