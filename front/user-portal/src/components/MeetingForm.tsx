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
import { DateTimePicker } from './DateTimePicker';
import { LuMic, LuVideo, LuClock, LuInfo, LuLoader } from 'react-icons/lu';
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
  const [needsTranscription, setNeedsTranscription] = useState(
    meeting?.needs_transcription || false
  );
  const [forceEndAtDuration, setForceEndAtDuration] = useState(
    meeting?.force_end_at_duration || false
  );
  const [isPermanent, setIsPermanent] = useState(
    meeting?.is_permanent || false
  );
  const [allowAnonymous, setAllowAnonymous] = useState(
    meeting?.allow_anonymous || false
  );
  const [additionalNotes, setAdditionalNotes] = useState(
    meeting?.additional_notes || ''
  );

  // Auto-enable audio recording when video is enabled
  useEffect(() => {
    if (needsVideoRecord && !needsAudioRecord) {
      setNeedsAudioRecord(true);
    }
  }, [needsVideoRecord]);

  // Participants and departments
  const [speakerId, setSpeakerId] = useState('');
  const [selectedUserIds, setSelectedUserIds] = useState<string[]>([]);
  const [selectedDepartmentIds, setSelectedDepartmentIds] = useState<string[]>([]);

  // Current user ID (creator/organizer)
  const [currentUserId, setCurrentUserId] = useState<string>('');

  // Dropdowns data
  const [subjects, setSubjects] = useState<MeetingSubject[]>([]);
  const [users, setUsers] = useState<User[]>([]);
  const [departments, setDepartments] = useState<Department[]>([]);

  // UI state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loadingData, setLoadingData] = useState(true);

  // Get current user on mount
  useEffect(() => {
    const storedUser = localStorage.getItem('user') || sessionStorage.getItem('user');
    if (storedUser) {
      try {
        const user = JSON.parse(storedUser);
        setCurrentUserId(user.id || '');
      } catch (err) {
        console.error('Failed to parse user data:', err);
      }
    }
  }, []);

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

  // Automatically add current user to participants when creating new meeting
  useEffect(() => {
    if (!isEditMode && currentUserId && !selectedUserIds.includes(currentUserId)) {
      setSelectedUserIds((prev) => [...prev, currentUserId]);
    }
  }, [currentUserId, isEditMode]);

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
          subject_id: subjectId || undefined,
          needs_video_record: needsVideoRecord,
          needs_audio_record: needsAudioRecord,
          needs_transcription: needsTranscription,
          force_end_at_duration: forceEndAtDuration,
          is_permanent: isPermanent,
          allow_anonymous: allowAnonymous,
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
          subject_id: subjectId || undefined,
          needs_video_record: needsVideoRecord,
          needs_audio_record: needsAudioRecord,
          needs_transcription: needsTranscription,
          force_end_at_duration: forceEndAtDuration,
          is_permanent: isPermanent,
          allow_anonymous: allowAnonymous,
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
          <LuLoader className="spinner spinner-lg" />
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
        <div className="alert alert-error">
          <LuInfo />
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

            <div className="form-group">
              <SearchableSelect
                id="subject"
                label={t('meetings.form.subject')}
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
              <DateTimePicker
                id="scheduled-at"
                type="datetime-local"
                value={scheduledAt}
                onChange={(value) => setScheduledAt(value)}
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
              onChange={(e) => {
                const newRecurrence = e.target.value as MeetingRecurrence;
                setRecurrence(newRecurrence);
                // Auto-set isPermanent when selecting permanent recurrence
                setIsPermanent(newRecurrence === 'permanent');
              }}
              className="form-select"
            >
              <option value="none">{t('meetings.recurrence.none')}</option>
              <option value="daily">{t('meetings.recurrence.daily')}</option>
              <option value="weekly">{t('meetings.recurrence.weekly')}</option>
              <option value="monthly">{t('meetings.recurrence.monthly')}</option>
              <option value="permanent">
                {t('meetings.recurrence.permanent')} ({t('meetings.recurrence.alwaysAvailable')})
              </option>
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
            placeholder={t('meetings.form.searchUsers') || ''}
            emptyMessage={t('meetings.form.noUsersAvailable')}
            disabledIds={speakerId ? [speakerId, currentUserId] : [currentUserId]}
            organizerId={currentUserId}
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
            placeholder={t('meetings.form.searchDepartments') || ''}
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
                disabled={needsVideoRecord}
              />
              <span className="checkbox-label">
                <LuMic className="checkbox-icon" /> {t('meetings.form.audioRecording')}
                {needsVideoRecord && <span className="text-muted"> ({t('meetings.form.autoEnabled')})</span>}
              </span>
            </label>
          </div>

          <div className="form-group">
            <label className="checkbox-item">
              <input
                type="checkbox"
                checked={needsTranscription}
                onChange={(e) => setNeedsTranscription(e.target.checked)}
              />
              <span className="checkbox-label">
                <LuMic className="checkbox-icon" /> {t('meetings.form.transcription')}
              </span>
            </label>
            {needsTranscription && (
              <p className="form-help-text">
                {t('meetings.form.transcriptionHelp')}
              </p>
            )}
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

          <div className="form-group">
            <label className="checkbox-item">
              <input
                type="checkbox"
                checked={allowAnonymous}
                onChange={(e) => setAllowAnonymous(e.target.checked)}
              />
              <span className="checkbox-label">
                {t('meetings.form.allowAnonymous')}
              </span>
            </label>
            {allowAnonymous && (
              <p className="form-help-text">
                {t('meetings.form.allowAnonymousHelp')}
              </p>
            )}
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
            className="btn btn-success"
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
