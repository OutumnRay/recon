import { useState, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { useParams, useNavigate } from 'react-router-dom';
import { getMeeting, getMeetingRecordings } from '../services/meetings';
import type { Meeting, RoomRecording, TrackRecording } from '../types/meeting';
import HLSPlayer from '../components/HLSPlayer';
import './MeetingRecordings.css';

type SelectedRecording = {
  type: 'room' | 'track';
  playlist_url: string;
  started_at: string;
  ended_at?: string;
  status: string;
  participantName?: string;
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
  const [expandedRooms, setExpandedRooms] = useState<Set<string>>(new Set());
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

        // Fetch room recordings with tracks
        const recordingsData = await getMeetingRecordings(meetingId);
        setRoomRecordings(recordingsData);

        // Expand all rooms by default and auto-select first available recording
        const allRoomIds = new Set(recordingsData.map(r => r.id));
        setExpandedRooms(allRoomIds);

        // Auto-select first room composite or first track
        if (recordingsData.length > 0) {
          const firstRoom = recordingsData[0];
          if (firstRoom.playlist_url) {
            // Select room composite
            setSelectedRecording({
              type: 'room',
              playlist_url: firstRoom.playlist_url,
              started_at: firstRoom.started_at,
              ended_at: firstRoom.ended_at,
              status: firstRoom.status,
            });
          } else if (firstRoom.tracks.length > 0) {
            // Select first track
            const firstTrack = firstRoom.tracks[0];
            setSelectedRecording({
              type: 'track',
              playlist_url: firstTrack.playlist_url,
              started_at: firstTrack.started_at,
              ended_at: firstTrack.ended_at,
              status: firstTrack.status,
              participantName: getParticipantLabel(firstTrack),
            });
          }
        }
      } catch (err: any) {
        console.error('Failed to fetch recordings:', err);
        setError(err instanceof Error ? err.message : t('meetingRecordings.errors.loadFailed'));
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [meetingId]);

  const toggleRoom = (roomId: string) => {
    setExpandedRooms(prev => {
      const newSet = new Set(prev);
      if (newSet.has(roomId)) {
        newSet.delete(roomId);
      } else {
        newSet.add(roomId);
      }
      return newSet;
    });
  };

  const selectRoomRecording = (room: RoomRecording) => {
    if (!room.playlist_url) return;

    setSelectedRecording({
      type: 'room',
      playlist_url: room.playlist_url,
      title: `Общая запись (${new Date(room.started_at).toLocaleString('ru-RU')})`,
      started_at: room.started_at,
      ended_at: room.ended_at,
      status: room.status,
    });
  };

  const selectTrackRecording = (track: TrackRecording) => {
    setSelectedRecording({
      type: 'track',
      playlist_url: track.playlist_url,
      title: `Трек участника`,
      started_at: track.started_at,
      ended_at: track.ended_at,
      status: track.status,
    });
  };

  const getTotalTracks = () => {
    return roomRecordings.reduce((sum, room) => sum + room.tracks.length, 0);
  };

  if (loading) {
    return (
      <div className="meeting-recordings-page">
        <div className="loading">Загрузка записей...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="meeting-recordings-page">
        <div className="error">
          <h2>Ошибка</h2>
          <p>{error}</p>
          <button onClick={() => navigate('/dashboard/meetings')}>Вернуться к встречам</button>
        </div>
      </div>
    );
  }

  return (
    <div className="meeting-recordings-page">
      <div className="recordings-header">
        <button className="back-button" onClick={() => navigate('/dashboard/meetings')}>
          ← Назад к встречам
        </button>
        <h1>Записи встречи: {meeting?.title || 'Встреча'}</h1>
        {meeting?.scheduled_at && (
          <p className="meeting-date">
            {new Date(meeting.scheduled_at).toLocaleString('ru-RU')}
          </p>
        )}
      </div>

      {roomRecordings.length === 0 ? (
        <div className="no-recordings">
          <p>Нет доступных записей для этой встречи</p>
        </div>
      ) : (
        <div className="recordings-container">
          <div className="recordings-sidebar">
            <div className="recording-stats">
              <span>Сессий: {roomRecordings.length}</span>
              <span>Треков: {getTotalTracks()}</span>
            </div>

            <div className="recordings-list">
              {roomRecordings.map((room, idx) => (
                <div key={room.id} className="room-recording-group">
                  <div className="room-header" onClick={() => toggleRoom(room.id)}>
                    <span className="room-toggle">
                      {expandedRooms.has(room.id) ? '▼' : '▶'}
                    </span>
                    <div className="room-info">
                      <div className="room-title">
                        Сессия {idx + 1}
                      </div>
                      <div className="room-time">
                        {new Date(room.started_at).toLocaleString('ru-RU')}
                      </div>
                    </div>
                    <div className="room-badge">
                      {room.tracks.length} трек{room.tracks.length !== 1 ? 'ов' : ''}
                    </div>
                  </div>

                  {expandedRooms.has(room.id) && (
                    <div className="room-recordings">
                      {room.playlist_url && (
                        <div
                          className={`recording-card ${selectedRecording?.type === 'room' && selectedRecording.playlist_url === room.playlist_url ? 'selected' : ''}`}
                          onClick={() => selectRoomRecording(room)}
                        >
                          <div className="recording-icon">🎥</div>
                          <div className="recording-info">
                            <div className="recording-title">Общая запись комнаты</div>
                            <div className="recording-details">
                              <span>{new Date(room.started_at).toLocaleTimeString('ru-RU')}</span>
                              {room.ended_at && (
                                <span>
                                  {Math.floor((new Date(room.ended_at).getTime() - new Date(room.started_at).getTime()) / 60000)} мин
                                </span>
                              )}
                            </div>
                          </div>
                          <div className={`recording-status ${room.status}`}>✓</div>
                        </div>
                      )}

                      {room.tracks.map(track => (
                        <div
                          key={track.id}
                          className={`recording-card track ${selectedRecording?.type === 'track' && selectedRecording.playlist_url === track.playlist_url ? 'selected' : ''}`}
                          onClick={() => selectTrackRecording(track)}
                        >
                          <div className="recording-icon">🎤</div>
                          <div className="recording-info">
                            <div className="recording-title">
                              {track.participant ?
                                (track.participant.first_name && track.participant.last_name ?
                                  `${track.participant.first_name} ${track.participant.last_name}` :
                                  track.participant.username) :
                                `Участник ${track.participant_id.slice(0, 8)}`}
                            </div>
                            <div className="recording-details">
                              <span>{new Date(track.started_at).toLocaleTimeString('ru-RU')}</span>
                              {track.ended_at && (
                                <span>
                                  {Math.floor((new Date(track.ended_at).getTime() - new Date(track.started_at).getTime()) / 60000)} мин
                                </span>
                              )}
                            </div>
                          </div>
                          <div className={`recording-status ${track.status}`}>✓</div>
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div className="player-container">
            {selectedRecording ? (
              <>
                <div className="player-header">
                  <h2>{selectedRecording.title}</h2>
                  <div className="recording-meta">
                    <span>Начало: {new Date(selectedRecording.started_at).toLocaleString('ru-RU')}</span>
                    {selectedRecording.ended_at && (
                      <span>Окончание: {new Date(selectedRecording.ended_at).toLocaleString('ru-RU')}</span>
                    )}
                    <span className={`status ${selectedRecording.status}`}>
                      {selectedRecording.status}
                    </span>
                  </div>
                </div>
                <HLSPlayer
                  src={selectedRecording.playlist_url}
                  autoplay={false}
                />
              </>
            ) : (
              <div className="no-selection">
                <p>Выберите запись для воспроизведения</p>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
