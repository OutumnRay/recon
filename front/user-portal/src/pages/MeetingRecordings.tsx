import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { getMeeting, getMeetingRecordings, getRoomTranscripts } from '../services/meetings';
import type { Meeting, RoomRecording, RoomTranscripts } from '../types/meeting';
import HLSPlayer from '../components/HLSPlayer';
import { LuFilm, LuMic, LuClock3 } from 'react-icons/lu';
import './MeetingRecordings.css';

type SelectedRecording = {
  playlist_url: string;
  started_at: string;
  ended_at?: string;
  status: string;
  audio_only: boolean;
  room_sid: string; // For finding tracks
};

type AccordionSection = 'video' | 'tracks' | 'transcript' | 'memo' | null;

export default function MeetingRecordings() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();
  const { t, i18n } = useTranslation();
  const locale = i18n.language?.startsWith('ru') ? 'ru-RU' : 'en-US';
  const formatDateTime = (value: string) => new Date(value).toLocaleString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
  const formatTime = (value: string) => new Date(value).toLocaleTimeString(locale, {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  });
  const formatMinutes = (minutes: number) => {
    if (minutes <= 0) return '—';
    if (minutes < 60) return `${minutes}m`;
    const hours = Math.floor(minutes / 60);
    const rest = minutes % 60;
    if (rest === 0) return `${hours}h`;
    return `${hours}h ${rest}m`;
  };

  const formatParticipantDisplayName = (participant?: any) => {
    if (!participant) return t('meetingRecordings.unknownParticipant');
    const first = participant.first_name?.trim();
    const last = participant.last_name?.trim();
    if (first && last) {
      return `${first} ${last}`;
    }
    return (
      participant.username ||
      first ||
      participant.email ||
      t('meetingRecordings.unknownParticipant')
    );
  };

  const getParticipantName = (track: any) => formatParticipantDisplayName(track.participant);

  const getInitials = (value: string) => {
    const cleaned = value.trim();
    if (!cleaned) return '??';
    const parts = cleaned.split(' ');
    if (parts.length === 1) {
      return parts[0].slice(0, 2).toUpperCase();
    }
    return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
  };

  const speakerAccentMap = useRef<Map<string, string>>(new Map());
  const speakerColorPalette = ['accent-blue', 'accent-purple', 'accent-orange', 'accent-green', 'accent-pink'];
  const getSpeakerAccent = (speaker: string) => {
    if (!speakerAccentMap.current.has(speaker)) {
      const nextColor = speakerColorPalette[speakerAccentMap.current.size % speakerColorPalette.length];
      speakerAccentMap.current.set(speaker, nextColor);
    }
    return speakerAccentMap.current.get(speaker) as string;
  };

  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [roomRecordings, setRoomRecordings] = useState<RoomRecording[]>([]);
  const [selectedRecording, setSelectedRecording] = useState<SelectedRecording | null>(null);
  const [roomTranscripts, setRoomTranscripts] = useState<RoomTranscripts | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadingTranscripts, setLoadingTranscripts] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [openSection, setOpenSection] = useState<AccordionSection>('video');
  const [transcribingTracks, setTranscribingTracks] = useState<Set<string>>(new Set());

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
              room_sid: firstRoom.room_sid,
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
      room_sid: room.room_sid,
    });
  };

  // Fetch transcripts when selected room changes
  useEffect(() => {
    if (!selectedRecording?.room_sid) return;

    const fetchTranscripts = async () => {
      try {
        setLoadingTranscripts(true);
        console.log('📝 [TRANSCRIPTS] Fetching for room:', selectedRecording.room_sid);
        const transcripts = await getRoomTranscripts(selectedRecording.room_sid);
        setRoomTranscripts(transcripts);
        console.log('📝 [TRANSCRIPTS] Loaded:', transcripts);
      } catch (err) {
        console.error('📝 [TRANSCRIPTS] Failed to load:', err);
        setRoomTranscripts(null); // Clear transcripts on error
      } finally {
        setLoadingTranscripts(false);
      }
    };

    fetchTranscripts();
  }, [selectedRecording?.room_sid]);

  const forceTranscription = async (trackId: string) => {
    if (transcribingTracks.has(trackId)) return;

    try {
      setTranscribingTracks(prev => new Set(prev).add(trackId));

      const token = localStorage.getItem('token') || sessionStorage.getItem('token');
      const response = await fetch(`/api/tracks/${trackId}/transcribe`, {
        method: 'POST',
        headers: {
          'Authorization': token ? `Bearer ${token}` : '',
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to start transcription: ${response.statusText}`);
      }

      console.log(`[Transcription] Started transcription for track ${trackId}`);
      // TODO: Add toast notification or update UI to show transcription in progress
    } catch (err) {
      console.error('[Transcription] Error starting transcription:', err);
      setTranscribingTracks(prev => {
        const next = new Set(prev);
        next.delete(trackId);
        return next;
      });
      // TODO: Add error toast notification
    }
  };

  const toggleSection = (section: AccordionSection) => {
    setOpenSection(openSection === section ? null : section);
  };

  const meetingTitle = meeting?.title || t('meetingRecordings.defaultTitle');
  const selectedRecordingTitle = selectedRecording
    ? t('meetingRecordings.roomRecordingTitle', { date: formatDateTime(selectedRecording.started_at) })
    : '';
  const totalTracks = roomRecordings.reduce((sum, room) => sum + (room.tracks?.length || 0), 0);
  const totalDurationMinutes = roomRecordings.reduce((sum, room) => {
    if (room.ended_at) {
      const diff = (new Date(room.ended_at).getTime() - new Date(room.started_at).getTime()) / 60000;
      return sum + Math.max(0, Math.round(diff));
    }
    return sum;
  }, 0);

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
        <>
          <div className="meeting-highlights">
            <div className="highlight-card">
              <div className="highlight-icon accent-blue">
                <LuFilm />
              </div>
              <div>
                <p>{t('meetingRecordings.stats.sessionsLabel')}</p>
                <strong>{roomRecordings.length}</strong>
                <span>{t('meetingRecordings.stats.completed')}</span>
              </div>
            </div>
            <div className="highlight-card">
              <div className="highlight-icon accent-purple">
                <LuMic />
              </div>
              <div>
                <p>{t('meetingRecordings.stats.tracksLabel')}</p>
                <strong>{totalTracks}</strong>
                <span>{t('meetingRecordings.stats.tracksSubtitle')}</span>
              </div>
            </div>
            <div className="highlight-card">
              <div className="highlight-icon accent-orange">
                <LuClock3 />
              </div>
              <div>
                <p>{t('meetingRecordings.stats.duration')}</p>
                <strong>{formatMinutes(totalDurationMinutes)}</strong>
                <span>{t('meetingRecordings.stats.durationSubtitle')}</span>
              </div>
            </div>
          </div>

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

                {/* Accordion Sections */}
                <div className="accordion-sections">
                  {/* Video Section */}
                  <div className="accordion-item">
                    <button
                      className={`accordion-header ${openSection === 'video' ? 'active' : ''}`}
                      onClick={() => toggleSection('video')}
                    >
                      <span>{t('meetingRecordings.sections.video')}</span>
                      <span className="accordion-icon">{openSection === 'video' ? '▼' : '▶'}</span>
                    </button>
                    {openSection === 'video' && (
                      <div className="accordion-content">
                        {meeting && meeting.needs_video_record && !selectedRecording.audio_only ? (
                          <div className="player-surface">
                            <HLSPlayer
                              src={selectedRecording.playlist_url}
                              autoplay={false}
                            />
                          </div>
                        ) : (
                          meeting && meeting.needs_audio_record && (
                            <div className="audio-player-container">
                              <audio controls style={{ width: '100%' }}>
                                <source src={selectedRecording.playlist_url} type="application/x-mpegURL" />
                                Your browser does not support the audio element.
                              </audio>
                            </div>
                          )
                        )}
                      </div>
                    )}
                  </div>

                  {/* Tracks Section */}
                  {(() => {
                    const currentRoom = roomRecordings.find(r => r.room_sid === selectedRecording.room_sid);
                    if (currentRoom && currentRoom.tracks && currentRoom.tracks.length > 0) {
                      return (
                        <div className="accordion-item">
                          <button
                            className={`accordion-header ${openSection === 'tracks' ? 'active' : ''}`}
                            onClick={() => toggleSection('tracks')}
                          >
                            <span>{t('meetingRecordings.sections.tracks')} ({currentRoom.tracks.length})</span>
                            <span className="accordion-icon">{openSection === 'tracks' ? '▼' : '▶'}</span>
                          </button>
                          {openSection === 'tracks' && (
                            <div className="accordion-content">
                              <div className="tracks-list">
                                {currentRoom.tracks.map((track) => {
                                  const participantName = getParticipantName(track);
                                  const trackDuration = track.ended_at
                                    ? Math.max(
                                      1,
                                      Math.round(
                                        (new Date(track.ended_at).getTime() - new Date(track.started_at).getTime()) / 60000,
                                      ),
                                    )
                                    : null;
                                  return (
                                    <div key={track.id} className="track-item">
                                      <div className="track-avatar">
                                        {getInitials(participantName)}
                                      </div>
                                      <div className="track-body">
                                        <div className="track-info">
                                          <span className="track-participant">
                                            {participantName}
                                          </span>
                                          <div className="track-meta">
                                            <span className={`track-type-pill ${(track.type || 'audio').toLowerCase()}`}>
                                              {(track.type || 'audio').toUpperCase()}
                                            </span>
                                            <span className="track-time">
                                              {formatTime(track.started_at)}
                                              {track.ended_at && ` – ${formatTime(track.ended_at)}`}
                                            </span>
                                            {trackDuration && (
                                              <span className="track-duration">
                                                {formatMinutes(trackDuration)}
                                              </span>
                                            )}
                                          </div>
                                        </div>
                                        <div className="track-player-wrapper">
                                          <HLSPlayer
                                            src={track.playlist_url}
                                            audioOnly={true}
                                            className="track-player"
                                          />
                                          {track.transcription_phrases && track.transcription_phrases.length > 0 ? (
                                            <div className="transcribed-badge">
                                              ✓ {t('meetingRecordings.transcribed')}
                                            </div>
                                          ) : track.transcription_status === 'processing' || transcribingTracks.has(track.id) ? (
                                            <button
                                              className="transcribe-button"
                                              disabled
                                            >
                                              {t('meetingRecordings.transcribing')}
                                            </button>
                                          ) : (
                                            <button
                                              className="transcribe-button"
                                              onClick={() => forceTranscription(track.id)}
                                            >
                                              {t('meetingRecordings.forceTranscribe')}
                                            </button>
                                          )}
                                        </div>
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            </div>
                          )}
                        </div>
                      );
                    }
                    return null;
                  })()}

                  {/* Transcript Section */}
                  {(() => {
                    // Collect all transcription phrases from the selected room's transcripts
                    const allPhrases: Array<{
                      absoluteTimestamp: number;
                      start: number;
                      end: number;
                      text: string;
                      speaker: string;
                      trackId: string;
                      trackStartedAt: string;
                    }> = [];

                    // Use the fetched room transcripts
                    if (roomTranscripts && roomTranscripts.tracks && roomTranscripts.tracks.length > 0) {
                      roomTranscripts.tracks.forEach((trackTranscript) => {
                        if (trackTranscript.transcription_phrases && trackTranscript.transcription_phrases.length > 0) {
                          // Get track start time in milliseconds
                          const trackStartTime = new Date(trackTranscript.started_at).getTime();
                          const participantName = formatParticipantDisplayName(trackTranscript.participant);

                          trackTranscript.transcription_phrases.forEach((phrase) => {
                            // Calculate absolute timestamp by adding track start time to phrase start time
                            const absoluteTimestamp = trackStartTime + (phrase.start * 1000);

                            allPhrases.push({
                              absoluteTimestamp,
                              start: phrase.start,
                              end: phrase.end,
                              text: phrase.text,
                              speaker: participantName,
                              trackId: trackTranscript.track_id,
                              trackStartedAt: trackTranscript.started_at,
                            });
                          });
                        }
                      });
                    }

                    // Sort by absolute timestamp
                    allPhrases.sort((a, b) => a.absoluteTimestamp - b.absoluteTimestamp);

                    const hasTranscriptions = allPhrases.length > 0;

                    return (
                      <div className="accordion-item">
                        <button
                          className={`accordion-header ${openSection === 'transcript' ? 'active' : ''}`}
                          onClick={() => toggleSection('transcript')}
                        >
                          <span>{t('meetingRecordings.sections.transcript')}</span>
                          <span className="accordion-icon">{openSection === 'transcript' ? '▼' : '▶'}</span>
                        </button>
                        {openSection === 'transcript' && (
                          <div className="accordion-content">
                            {loadingTranscripts ? (
                              <div className="loading-state">
                                <div className="loading-spinner"></div>
                                <p>{t('common.loading')}</p>
                              </div>
                            ) : hasTranscriptions ? (
                              <div className="transcript-timeline">
                                {allPhrases.map((phrase, idx) => {
                                  const accent = getSpeakerAccent(phrase.speaker);
                                  const isLast = idx === allPhrases.length - 1;
                                  return (
                                    <div key={`${phrase.trackId}-${idx}`} className="transcript-entry">
                                      <div className="transcript-marker">
                                        <span className={`marker-dot ${accent}`} />
                                        {!isLast && <span className="marker-line" />}
                                      </div>
                                      <div className={`transcript-bubble ${accent}`}>
                                        <div className="bubble-header">
                                          <span className="bubble-speaker">{phrase.speaker}</span>
                                        </div>
                                        <p className="bubble-text">{phrase.text}</p>
                                      </div>
                                      <div className="transcript-absolute">
                                        {new Date(phrase.absoluteTimestamp).toLocaleTimeString(locale, {
                                          hour: '2-digit',
                                          minute: '2-digit',
                                          second: '2-digit'
                                        })}
                                      </div>
                                    </div>
                                  );
                                })}
                              </div>
                            ) : (
                              <p className="no-transcription-message">{t('meetingRecordings.noTranscription')}</p>
                            )}
                          </div>
                        )}
                      </div>
                    );
                  })()}

                  {/* Memo Section (Coming Soon) */}
                  <div className="accordion-item">
                    <button
                      className={`accordion-header ${openSection === 'memo' ? 'active' : ''}`}
                      onClick={() => toggleSection('memo')}
                    >
                      <span>{t('meetingRecordings.sections.memo')}</span>
                      <span className="accordion-badge">{t('meetingRecordings.comingSoon')}</span>
                      <span className="accordion-icon">{openSection === 'memo' ? '▼' : '▶'}</span>
                    </button>
                    {openSection === 'memo' && (
                      <div className="accordion-content">
                        <p className="coming-soon-message">{t('meetingRecordings.memoComingSoon')}</p>
                      </div>
                    )}
                  </div>
                </div>
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
                const showLine = idx < roomRecordings.length - 1;

                return (
                  <div key={room.id} className={`session-card ${roomSelected ? 'active' : ''}`}>
                    <div className="session-timeline">
                      <span className="session-dot" />
                      {showLine && <span className="session-line" />}
                    </div>
                    <div className="session-card-content">
                      <div className="session-card-header">
                        <div>
                          <p className="session-label">{t('meetingRecordings.sessionLabel', { number: idx + 1 })}</p>
                          <h3>{formatDateTime(room.started_at)}</h3>
                        </div>
                        <div className="session-meta">
                          {roomDurationMinutes && (
                            <span>{t('meetingRecordings.durationMinutes', { minutes: roomDurationMinutes })}</span>
                          )}
                          <span className={`session-status ${room.status}`}>{room.status}</span>
                        </div>
                      </div>

                      {room.playlist_url && (
                        <button
                          className={`session-recording ${roomSelected ? 'active' : ''}`}
                          onClick={() => selectRoomRecording(room)}
                        >
                          <div className="session-recording-info">
                            <span className="session-recording-label">{t('meetingRecordings.roomRecordingLabel')}</span>
                            <span className="session-recording-time">{formatTime(room.started_at)}</span>
                          </div>
                          <span className="session-recording-cta">
                            {roomSelected ? t('meetingRecordings.nowPlaying') : t('meetingRecordings.playRecording')}
                          </span>
                        </button>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
        </>
      )}
    </div>
  );
}
