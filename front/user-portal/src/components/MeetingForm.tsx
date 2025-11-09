import React, { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import {
  createMeeting,
  updateMeeting,
  listMeetingSubjects,
} from '../services/meetings';
import type {
  MeetingWithDetails,
  MeetingType,
  MeetingRecurrence,
  CreateMeetingRequest,
  UpdateMeetingRequest,
  MeetingSubject,
} from '../types/meeting';
import { SearchableSelect } from './SearchableSelect';
import { MultiSelectWithSearch } from './MultiSelectWithSearch';
import { LuUsers, LuMic, LuVideo, LuClock } from 'react-icons/lu';
import './MeetingForm.css';

interface MeetingFormProps {
  meeting?: MeetingWithDetails;
  onSuccess?: (meeting: MeetingWithDetails) => void;
  onCancel?: () => void;
}

interface User {
  id: string;
  username: string;
  email: string;
  full_name: string;
}

interface Department {
  id: string;
  name: string;
  description: string;
}

export const MeetingForm: React.FC<MeetingFormProps> = ({
  meeting,
  onSuccess,
  onCancel,
}) => {
  const { t } = useTranslation();
  const isEditMode = !!meeting;

  // Form state
  const [title, setTitle] = useState(meeting?.title || '');

  // Convert UTC date from backend to local datetime-local format
  const getLocalDateTimeString = (utcDateString: string): string => {
    const date = new Date(utcDateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  };

  const [scheduledAt, setScheduledAt] = useState(
    meeting?.scheduled_at
      ? getLocalDateTimeString(meeting.scheduled_at)
      : ''
  );
  const [duration, setDuration] = useState(meeting?.duration || 60);
  const [recurrence, setRecurrence] = useState<MeetingRecurrence>(
    meeting?.recurrence || 'none'
  );
  const [type, setType] = useState<MeetingType>(meeting?.type || 'conference');
  const [subjectId, setSubjectId] = useState(meeting?.subject_id || '');
  const [needsVideoRecord, setNeedsVideoRecord] = useState(
    meeting?.needs_video_record || false
  );
  const [needsAudioRecord, setNeedsAudioRecord] = useState(
    meeting?.needs_audio_record || false
  );
  const [forceEndAtDuration, setForceEndAtDuration] = useState(
    meeting?.force_end_at_duration || false
  );
  const [additionalNotes, setAdditionalNotes] = useState(
    meeting?.additional_notes || ''
  );

  // Participants and departments
  const [speakerId, setSpeakerId] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [selectedDepartmentIds, setSelectedDepartmentIds] = useState<string[]>([]);

  // Dropdowns data
  const [subjects, setSubjects] = useState<MeetingSubject[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadingData, setLoadingData] = useState(true);

  // Fetch dropdown data
  useEffect(() => {
    fetchFormData();
  }, []);

  // Populate form if editing
  useEffect(() => {
    if (meeting) {
      const speaker = meeting.participants.find((p) => p.role === 'speaker');
      if (speaker) {
        setSpeakerId(speaker.user_id);
      }

      const participants = meeting.participants
        .filter((p) => p.role === 'participant')
        .map((p) => p.user_id);
      setSelectedUserIds(participants);

      const deptIds = meeting.departments.map((d) => d.id);
      setSelectedDepartmentIds(deptIds);
    }
  }, [meeting]);

  const fetchFormData = async () => {
    try {
      setLoadingData(true);

      // Fetch subjects
      const subjectsResponse = await listMeetingSubjects(1, 100);
      setSubjects(subjectsResponse.items);

      // Fetch users (TODO: Replace with actual API call)
      await fetchUsers();

      // Fetch departments (TODO: Replace with actual API call)
      await fetchDepartments();

      setError(null);
    } catch (err) {
      setError('Failed to load form data');
      console.error(err);
    } finally {
      setLoadingData(false);
    }
  };

  const fetchUsers = async () => {
    // TODO: Replace with actual API call to get users
    // For now, using placeholder
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/users', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setUsers(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch users:', err);
      // Set empty array if API doesn't exist yet
      setUsers([]);
    }
  };

  const fetchDepartments = async () => {
    // TODO: Replace with actual API call to get departments
    try {
      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch('/api/v1/departments', {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setDepartments(data.items || []);
      }
    } catch (err) {
      console.error('Failed to fetch departments:', err);
      // Set empty array if API doesn't exist yet
      setDepartments([]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // Validation
    if (!title.trim()) {
      setError(t('meetings.errors.titleRequired'));
      return;
    }

    if (!scheduledAt) {
      setError(t('meetings.errors.dateRequired'));
      return;
    }

    if (duration <= 0) {
      setError(t('meetings.errors.durationInvalid'));
      return;
    }

    if (!subjectId) {
      setError(t('meetings.errors.subjectRequired'));
      return;
    }

    if (type === 'presentation' && !speakerId) {
      setError(t('meetings.errors.speakerRequired'));
      return;
    }

    if (selectedUserIds.length === 0 && selectedDepartmentIds.length === 0) {
      setError(t('meetings.errors.participantsRequired'));
      return;
    }

    try {
      setLoading(true);

      const scheduledAtISO = new Date(scheduledAt).toISOString();

      if (isEditMode) {
        // Update meeting
        const updateRequest: UpdateMeetingRequest = {
          title,
          scheduled_at: scheduledAtISO,
          duration,
          recurrence,
          type,
          subject_id: subjectId,
          needs_video_record: needsVideoRecord,
          needs_audio_record: needsAudioRecord,
          force_end_at_duration: forceEndAtDuration,
          additional_notes: additionalNotes || undefined,
          speaker_id: type === 'presentation' ? speakerId : undefined,
          participant_ids: selectedUserIds,
          department_ids: selectedDepartmentIds,
        };

        const updatedMeeting = await updateMeeting(meeting!.id, updateRequest);
        onSuccess?.(updatedMeeting);
      } else {
        // Create meeting
        const createRequest: CreateMeetingRequest = {
          title,
          scheduled_at: scheduledAtISO,
          duration,
          recurrence,
          type,
          subject_id: subjectId,
          needs_video_record: needsVideoRecord,
          needs_audio_record: needsAudioRecord,
          force_end_at_duration: forceEndAtDuration,
          additional_notes: additionalNotes || undefined,
          speaker_id: type === 'presentation' ? speakerId : undefined,
          participant_ids: selectedUserIds,
          department_ids: selectedDepartmentIds,
        };

        const createdMeeting = await createMeeting(createRequest);
        onSuccess?.(createdMeeting);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : (isEditMode ? t('meetings.errors.updateFailed') : t('meetings.errors.createFailed'))
      );
    } finally {
      setLoading(false);
    }
  };

  if (loadingData) {
    return (
      <div className="meeting-form-container">
        <div className="loading-state">
          <div className="loading-spinner"></div>
          <p>{t('meetings.loadingForm')}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="meeting-form-container">
      <h2 className="form-title">
        {isEditMode ? t('meetings.form.editTitle') : t('meetings.form.createTitle')}
      </h2>

      {error && (
        <div className="error-message">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      <form onSubmit={handleSubmit} className="meeting-form">
        {/* Basic Information */}
        <div className="form-section">
          <h3 className="section-title">{t('meetings.form.basicInformation')}</h3>

          <div className="form-group">
            <label htmlFor="title" className="form-label required">
              {t('meetings.form.title')}
            </label>
            <input
              id="title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="form-input"
              placeholder={t('meetings.form.titlePlaceholder')}
              required
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="type" className="form-label required">
                {t('meetings.form.type')}
              </label>
              <div className="select-with-icon">
                {type === 'conference' ? <LuUsers className="select-icon" /> : <LuMic className="select-icon" />}
                <select
                  id="type"
                  value={type}
                  onChange={(e) => setType(e.target.value as MeetingType)}
                  className="form-select"
                  required
                >
                  <option value="conference">
                    {t('meetings.type.conference')}
                  </option>
                  <option value="presentation">
                    {t('meetings.type.presentation')}
                  </option>
                </select>
              </div>
            </div>

            <div className="form-group">
              <SearchableSelect
                id="subject"
                label={`${t('meetings.form.subject')} *`}
                value={subjectId}
                onChange={(value) => setSubjectId(value)}
                options={subjects.map(subject => ({
                  value: subject.id,
                  label: subject.name,
                }))}
                placeholder={t('meetings.form.selectSubject')}
                emptyPlaceholder={t('meetings.form.selectSubject')}
              />
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="scheduled-at" className="form-label required">
                {t('meetings.form.scheduledAt')}
              </label>
              <input
                id="scheduled-at"
                type="datetime-local"
                value={scheduledAt}
                onChange={(e) => setScheduledAt(e.target.value)}
                className="form-input"
                required
              />
            </div>

            <div className="form-group">
              <label htmlFor="duration" className="form-label required">
                {t('meetings.form.duration')}
              </label>
              <input
                id="duration"
                type="number"
                value={duration}
                onChange={(e) => setDuration(Number(e.target.value))}
                className="form-input"
                min="15"
                step="15"
                required
              />
            </div>
          </div>

          <div className="form-group">
            <label htmlFor="recurrence" className="form-label">
              {t('meetings.form.recurrence')}
            </label>
            <select
              id="recurrence"
              value={recurrence}
              onChange={(e) => setRecurrence(e.target.value as MeetingRecurrence)}
              className="form-select"
            >
              <option value="none">{t('meetings.recurrence.none')}</option>
              <option value="daily">{t('meetings.recurrence.daily')}</option>
              <option value="weekly">{t('meetings.recurrence.weekly')}</option>
              <option value="monthly">{t('meetings.recurrence.monthly')}</option>
            </select>
          </div>
        </div>

        {/* Participants */}
        <div className="form-section">
          <h3 className="section-title">{t('meetings.form.participants')}</h3>

          {type === 'presentation' && (
            <div className="form-group">
              <SearchableSelect
                id="speaker"
                label={`${t('meetings.form.speaker')} *`}
                value={speakerId}
                onChange={(value) => setSpeakerId(value)}
                options={users.map(user => ({
                  value: user.id,
                  label: `${user.full_name || user.username} (${user.email})`,
                }))}
                placeholder={t('meetings.form.selectSpeaker')}
                emptyPlaceholder={t('meetings.form.selectSpeaker')}
              />
            </div>
          )}

          <MultiSelectWithSearch
            label={t('meetings.form.individualParticipants')}
            options={users.map(user => ({
              id: user.id,
              name: user.full_name || user.username,
              email: user.email,
            }))}
            selectedIds={selectedUserIds}
            onChange={setSelectedUserIds}
            placeholder={t('meetings.form.searchUsers') || 'Поиск пользователей...'}
            emptyMessage={t('meetings.form.noUsersAvailable')}
            disabledIds={speakerId ? [speakerId] : []}
          />

          <MultiSelectWithSearch
            label={t('meetings.form.departments')}
            options={departments.map(dept => ({
              id: dept.id,
              name: dept.name,
              description: dept.description,
            }))}
            selectedIds={selectedDepartmentIds}
            onChange={setSelectedDepartmentIds}
            placeholder={t('meetings.form.searchDepartments') || 'Поиск подразделений...'}
            emptyMessage={t('meetings.form.noDepartmentsAvailable')}
          />
        </div>

        {/* Recording Options */}
        <div className="form-section">
          <h3 className="section-title">{t('meetings.form.recordingOptions')}</h3>

          <div className="form-group">
            <label className="checkbox-item">
              <input
                type="checkbox"
                checked={needsVideoRecord}
                onChange={(e) => setNeedsVideoRecord(e.target.checked)}
              />
              <span className="checkbox-label">
                <LuVideo className="checkbox-icon" /> {t('meetings.form.videoRecording')}
              </span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-item">
              <input
                type="checkbox"
                checked={needsAudioRecord}
                onChange={(e) => setNeedsAudioRecord(e.target.checked)}
              />
              <span className="checkbox-label">
                <LuMic className="checkbox-icon" /> {t('meetings.form.audioRecording')}
              </span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-item">
              <input
                type="checkbox"
                checked={forceEndAtDuration}
                onChange={(e) => setForceEndAtDuration(e.target.checked)}
              />
              <span className="checkbox-label">
                <LuClock className="checkbox-icon" /> {t('meetings.form.forceEndAtDuration')}
              </span>
            </label>
          </div>
        </div>

        {/* Additional Notes */}
        <div className="form-section">
          <h3 className="section-title">{t('meetings.form.additionalNotes')}</h3>

          <div className="form-group">
            <textarea
              value={additionalNotes}
              onChange={(e) => setAdditionalNotes(e.target.value)}
              className="form-textarea"
              placeholder={t('meetings.form.notesPlaceholder')}
              rows={5}
            />
          </div>
        </div>

        {/* Form Actions */}
        <div className="form-actions">
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="btn btn-secondary"
              disabled={loading}
            >
              {t('common.cancel')}
            </button>
          )}
          <button
            type="submit"
            className="btn btn-primary"
            disabled={loading}
          >
            {loading
              ? t('meetings.form.saving')
              : isEditMode
              ? t('meetings.form.updateButton')
              : t('meetings.form.createButton')}
          </button>
        </div>
      </form>
    </div>
  );
};

export default MeetingForm;
