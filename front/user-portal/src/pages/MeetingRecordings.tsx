import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { getMeeting, getMeetingRecordings } from '../services/meetings';
import type { Meeting, RoomRecording } from '../types/meeting';
import HLSPlayer from '../components/HLSPlayer';
import './MeetingRecordings.css';

type SelectedRecording = {
  playlist_url: string;
  started_at: string;
  ended_at?: string;
  status: string;
  audio_only: boolean;
};

export default function MeetingRecordings() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const locale = i18n.language?.startsWith('ru') ? 'ru-RU' : 'en-US';
  const formatDateTime = (value: string) => new Date(value).toLocaleString(locale);
  const formatTime = (value: string) => new Date(value).toLocaleTimeString(locale);

  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [roomRecordings, setRoomRecordings] = useState<RoomRecording[]>([]);
  const [selectedRecording, setSelectedRecording] = useState<SelectedRecording | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!meetingId) return;

    const fetchData = async () => {
      try {
        setLoading(true);

        // Fetch meeting details
        const meetingData = await getMeeting(meetingId);
        setMeeting(meetingData);

        // Check if authentication is required for non-anonymous meetings
        if (!meetingData.allow_anonymous) {
          const token = localStorage.getItem('token') || sessionStorage.getItem('token');
          if (!token) {
            // Redirect to login page for non-anonymous meetings
            console.log('[Auth] Meeting recordings require authentication, redirecting to login');
            navigate('/login', { state: { from: `/meeting/${meetingId}/recordings` } });
            return;
          }
        }

        // Fetch room recordings with tracks
        const recordingsData = await getMeetingRecordings(meetingId);
        setRoomRecordings(recordingsData);

        // Auto-select first room composite recording
        if (recordingsData.length > 0) {
          const firstRoom = recordingsData[0];
          if (firstRoom.playlist_url) {
            setSelectedRecording({
              playlist_url: firstRoom.playlist_url,
              started_at: firstRoom.started_at,
              ended_at: firstRoom.ended_at,
              status: firstRoom.status,
              audio_only: firstRoom.audio_only || false,
            });
          }
        }
      } catch (err: any) {
        console.error('Failed to fetch recordings:', err);
        setError(err instanceof Error ? err.message : t('meetingRecordings.errors.loadFailed'));
        // If meeting fetch fails with 401/403, it might be auth issue
        if (err instanceof Error && (err.message.includes('401') || err.message.includes('403'))) {
          navigate('/login', { state: { from: `/meeting/${meetingId}/recordings` } });
        }
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [meetingId, navigate, t]);

  const selectRoomRecording = (room: RoomRecording) => {
    if (!room.playlist_url) return;

    setSelectedRecording({
      playlist_url: room.playlist_url,
      started_at: room.started_at,
      ended_at: room.ended_at,
      status: room.status,
      audio_only: room.audio_only || false,
    });
  };

  const meetingTitle = meeting?.title || t('meetingRecordings.defaultTitle');
  const selectedRecordingTitle = selectedRecording
    ? t('meetingRecordings.roomRecordingTitle', { date: formatDateTime(selectedRecording.started_at) })
    : '';

  if (loading) {
    return (
      <div className="meeting-recordings-page">
        <div className="loading">{t('meetingRecordings.loading')}</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="meeting-recordings-page">
        <div className="error">
          <h2>{t('meetingRecordings.errorTitle')}</h2>
          <p>{error}</p>
          <button onClick={() => navigate('/dashboard/meetings')}>
            {t('meetingRecordings.backToMeetings')}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="meeting-recordings-page">
      <div className="recordings-header">
        <button className="back-button" onClick={() => navigate('/dashboard/meetings')}>
          ← {t('meetingRecordings.backToMeetings')}
        </button>
        <h1>{t('meetingRecordings.heading', { title: meetingTitle })}</h1>
        {meeting?.scheduled_at && (
          <p className="meeting-date">
            {formatDateTime(meeting.scheduled_at)}
          </p>
        )}
      </div>

      {roomRecordings.length === 0 ? (
        <div className="no-recordings">
          <p>{t('meetingRecordings.noRecordings')}</p>
        </div>
      ) : (
        <div className="recordings-layout">
          <div className="player-panel">
            {selectedRecording ? (
              <>
                <div className="player-header">
                  <div>
                    <p className="player-subtitle">{meetingTitle}</p>
                    <h2>{selectedRecordingTitle}</h2>
                  </div>
                  <div className="recording-meta">
                    <span>
                      {t('meetingRecordings.meta.start')}: {formatDateTime(selectedRecording.started_at)}
                    </span>
                    {selectedRecording.ended_at && (
                      <span>
                        {t('meetingRecordings.meta.end')}: {formatDateTime(selectedRecording.ended_at)}
                      </span>
                    )}
                    <span className={`status ${selectedRecording.status}`}>
                      {selectedRecording.status}
                    </span>
                  </div>
                </div>
                {/* Show video player only if video recording is enabled */}
                {meeting && meeting.needs_video_record && !selectedRecording.audio_only ? (
                  <div className="player-surface">
                    <HLSPlayer
                      src={selectedRecording.playlist_url}
                      autoplay={false}
                    />
                  </div>
                ) : (
                  /* Show audio player for audio-only recordings */
                  meeting && meeting.needs_audio_record && (
                    <div className="audio-player-container">
                      <audio controls style={{ width: '100%' }}>
                        <source src={selectedRecording.playlist_url} type="application/x-mpegURL" />
                        Your browser does not support the audio element.
                      </audio>
                    </div>
                  )
                )}
              </>
            ) : (
              <div className="no-selection">
                <p>{t('meetingRecordings.selectPrompt')}</p>
              </div>
            )}
          </div>

          <div className="sessions-panel">
            <div className="sessions-controls">
              <div className="sessions-stats">
                <span className="stat-pill">
                  {t('meetingRecordings.stats.sessions', { count: roomRecordings.length })}
                </span>
              </div>
            </div>

            <div className="sessions-list">
              {roomRecordings.map((room, idx) => {
                const roomSelected = selectedRecording?.playlist_url === room.playlist_url;
                const roomDurationMinutes = room.ended_at
                  ? Math.max(1, Math.floor((new Date(room.ended_at).getTime() - new Date(room.started_at).getTime()) / 60000))
                  : null;

                return (
                  <div key={room.id} className="session-card">
                    <div className="session-card-header">
                      <div>
                        <p className="session-label">{t('meetingRecordings.sessionLabel', { number: idx + 1 })}</p>
                        <h3>{formatDateTime(room.started_at)}</h3>
                      </div>
                      <div className="session-meta">
                        {roomDurationMinutes && (
                          <span>{t('meetingRecordings.durationMinutes', { minutes: roomDurationMinutes })}</span>
                        )}
                      </div>
                    </div>

                    {room.playlist_url && (
                      <button
                        className={`session-recording ${roomSelected ? 'active' : ''}`}
                        onClick={() => selectRoomRecording(room)}
                      >
                        <div className="session-recording-icon">{room.audio_only ? '🎙️' : '🎥'}</div>
                        <div className="session-recording-info">
                          <span className="session-recording-label">{t('meetingRecordings.roomRecordingLabel')}</span>
                          <span className="session-recording-time">{formatTime(room.started_at)}</span>
                        </div>
                        <div className={`session-status ${room.status}`}>
                          {room.status}
                        </div>
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
