import { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getMeeting, getMeetingRecordings } from '../services/meetings';
import type { Meeting, Recording } from '../types/meeting';
import HLSPlayer from '../components/HLSPlayer';
import './MeetingRecordings.css';

export default function MeetingRecordings() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();

  const [meeting, setMeeting] = useState<Meeting | null>(null);
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [selectedRecording, setSelectedRecording] = useState<Recording | null>(null);
  const [showTracksOnly, setShowTracksOnly] = useState(false);
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

        // Fetch recordings
        const recordingsData = await getMeetingRecordings(meetingId);
        setRecordings(recordingsData);

        // Auto-select first recording if available
        if (recordingsData.length > 0) {
          setSelectedRecording(recordingsData[0]);
        }
      } catch (err: any) {
        console.error('Failed to fetch recordings:', err);
        setError(err.message || 'Failed to load recordings');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [meetingId]);

  const filteredRecordings = recordings.filter(r =>
    !showTracksOnly || r.type === 'track'
  );

  const roomRecordings = recordings.filter(r => r.type === 'room');
  const trackRecordings = recordings.filter(r => r.type === 'track');

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
          <button onClick={() => navigate('/meetings')}>Вернуться к встречам</button>
        </div>
      </div>
    );
  }

  return (
    <div className="meeting-recordings-page">
      <div className="recordings-header">
        <button className="back-button" onClick={() => navigate('/meetings')}>
          ← Назад к встречам
        </button>
        <h1>Записи встречи: {meeting?.title || 'Встреча'}</h1>
        {meeting?.scheduled_at && (
          <p className="meeting-date">
            {new Date(meeting.scheduled_at).toLocaleString('ru-RU')}
          </p>
        )}
      </div>

      {recordings.length === 0 ? (
        <div className="no-recordings">
          <p>Нет доступных записей для этой встречи</p>
        </div>
      ) : (
        <div className="recordings-container">
          <div className="recordings-sidebar">
            <div className="recording-type-selector">
              <button
                className={!showTracksOnly ? 'active' : ''}
                onClick={() => setShowTracksOnly(false)}
              >
                Все записи ({recordings.length})
              </button>
              <button
                className={showTracksOnly ? 'active' : ''}
                onClick={() => setShowTracksOnly(true)}
              >
                По участникам ({trackRecordings.length})
              </button>
            </div>

            <div className="recordings-list">
              {!showTracksOnly && roomRecordings.length > 0 && (
                <div className="recording-group">
                  <h3>Общая запись комнаты</h3>
                  {roomRecordings.map(recording => (
                    <RecordingCard
                      key={recording.id}
                      recording={recording}
                      isSelected={selectedRecording?.id === recording.id}
                      onSelect={() => setSelectedRecording(recording)}
                    />
                  ))}
                </div>
              )}

              {filteredRecordings.filter(r => r.type === 'track').length > 0 && (
                <div className="recording-group">
                  <h3>Треки участников</h3>
                  {filteredRecordings
                    .filter(r => r.type === 'track')
                    .map(recording => (
                      <RecordingCard
                        key={recording.id}
                        recording={recording}
                        isSelected={selectedRecording?.id === recording.id}
                        onSelect={() => setSelectedRecording(recording)}
                      />
                    ))}
                </div>
              )}
            </div>
          </div>

          <div className="player-container">
            {selectedRecording ? (
              <>
                <div className="player-header">
                  <h2>
                    {selectedRecording.type === 'room'
                      ? 'Общая запись комнаты'
                      : `Трек: ${selectedRecording.participant?.full_name || selectedRecording.participant?.username || 'Участник'}`}
                  </h2>
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

interface RecordingCardProps {
  recording: Recording;
  isSelected: boolean;
  onSelect: () => void;
}

function RecordingCard({ recording, isSelected, onSelect }: RecordingCardProps) {
  const duration = recording.ended_at
    ? Math.floor((new Date(recording.ended_at).getTime() - new Date(recording.started_at).getTime()) / 1000 / 60)
    : null;

  return (
    <div
      className={`recording-card ${isSelected ? 'selected' : ''}`}
      onClick={onSelect}
    >
      <div className="recording-icon">
        {recording.type === 'room' ? '🎥' : '🎤'}
      </div>
      <div className="recording-info">
        <div className="recording-title">
          {recording.type === 'room'
            ? 'Общая запись'
            : recording.participant?.full_name || recording.participant?.username || 'Участник'}
        </div>
        <div className="recording-details">
          <span>{new Date(recording.started_at).toLocaleTimeString('ru-RU')}</span>
          {duration && <span>{duration} мин</span>}
        </div>
      </div>
      <div className={`recording-status ${recording.status}`}>
        {recording.status === 'completed' ? '✓' : '...'}
      </div>
    </div>
  );
}
