import { useEffect, useState } from 'react';
import { liveKitApi } from '../services/livekit';
import type { Room, Participant, Track } from '../services/livekit';
import {
  LuVideo,
  LuArrowLeft,
  LuUsers,
  LuClock,
  LuCircleCheck,
  LuMic,
  LuMicOff,
  LuMonitor,
  LuUser,
  LuActivity,
} from 'react-icons/lu';
import './RoomDetails.css';

interface RoomDetailsProps {
  roomSid: string;
}

export const RoomDetails = ({ roomSid }: RoomDetailsProps) => {
  const [room, setRoom] = useState<Room | null>(null);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [tracks, setTracks] = useState<Track[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'participants' | 'tracks'>('participants');

  useEffect(() => {
    loadRoomData();
  }, [roomSid]);

  const loadRoomData = async () => {
    try {
      setLoading(true);
      setError(null);

      const [roomData, participantsData, tracksData] = await Promise.all([
        liveKitApi.getRoom(roomSid),
        liveKitApi.getParticipants(roomSid),
        liveKitApi.getTracks(roomSid),
      ]);

      setRoom(roomData);
      setParticipants(participantsData || []);
      setTracks(tracksData || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load room data');
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString(undefined, { hour12: false });
  };

  const formatDuration = (startedAt: string, finishedAt?: string) => {
    const start = new Date(startedAt);
    const end = finishedAt ? new Date(finishedAt) : new Date();
    const duration = Math.floor((end.getTime() - start.getTime()) / 1000);

    const hours = Math.floor(duration / 3600);
    const minutes = Math.floor((duration % 3600) / 60);
    const seconds = duration % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${seconds}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${seconds}s`;
    } else {
      return `${seconds}s`;
    }
  };

  const getTrackIcon = (track: Track) => {
    if (track.source === 'MICROPHONE') {
      return track.status === 'published' ? <LuMic /> : <LuMicOff />;
    } else if (track.source === 'CAMERA') {
      return <LuMonitor />;
    } else if (track.source === 'SCREEN_SHARE') {
      return <LuMonitor />;
    }
    return <LuActivity />;
  };

  const handleBack = () => {
    window.location.href = '/rooms';
  };

  if (loading) {
    return (
      <div className="room-details-container">
        <div className="loading">Loading room details...</div>
      </div>
    );
  }

  if (error || !room) {
    return (
      <div className="room-details-container">
        <div className="error-message">
          <p>{error || 'Room not found'}</p>
          <button onClick={handleBack} className="btn-back">Back to Rooms</button>
        </div>
      </div>
    );
  }

  return (
    <div className="room-details-container">
      <div className="room-details-header">
        <button onClick={handleBack} className="btn-back-icon">
          <LuArrowLeft />
          Back
        </button>

        <div className="room-header-info">
          <div className="room-header-title">
            <LuVideo className="room-header-icon" />
            <h1>{room.name}</h1>
            <div className={`room-badge status-${room.status}`}>
              {room.status === 'active' ? (
                <>
                  <span className="status-dot"></span>
                  Active
                </>
              ) : (
                <>
                  <LuCircleCheck />
                  Finished
                </>
              )}
            </div>
          </div>
          <p className="room-sid">Room ID: {room.sid}</p>
        </div>

        <button onClick={loadRoomData} className="btn-refresh-details">
          Refresh
        </button>
      </div>

      <div className="room-stats-grid">
        <div className="stat-card">
          <div className="stat-icon">
            <LuUsers />
          </div>
          <div className="stat-content">
            <div className="stat-label">Participants</div>
            <div className="stat-value">{participants.length}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">
            <LuActivity />
          </div>
          <div className="stat-content">
            <div className="stat-label">Tracks</div>
            <div className="stat-value">{tracks.length}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">
            <LuClock />
          </div>
          <div className="stat-content">
            <div className="stat-label">Started</div>
            <div className="stat-value-small">{formatDate(room.startedAt)}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">
            <LuClock />
          </div>
          <div className="stat-content">
            <div className="stat-label">Duration</div>
            <div className="stat-value-small">{formatDuration(room.startedAt, room.finishedAt)}</div>
          </div>
        </div>
      </div>

      <div className="room-tabs">
        <button
          className={`tab-button ${activeTab === 'participants' ? 'active' : ''}`}
          onClick={() => setActiveTab('participants')}
        >
          <LuUsers />
          Participants ({participants.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'tracks' ? 'active' : ''}`}
          onClick={() => setActiveTab('tracks')}
        >
          <LuActivity />
          Tracks ({tracks.length})
        </button>
      </div>

      <div className="room-content">
        {activeTab === 'participants' && (
          <div className="participants-list">
            {participants.length === 0 ? (
              <div className="empty-state-small">
                <LuUsers className="empty-icon" />
                <p>No participants</p>
              </div>
            ) : (
              participants.map((participant) => (
                <div key={participant.id} className="participant-card">
                  <div className="participant-header">
                    <div className="participant-avatar">
                      <LuUser />
                    </div>
                    <div className="participant-info">
                      <h3>{participant.name}</h3>
                      <p className="participant-identity">{participant.identity}</p>
                    </div>
                    <div className={`participant-state state-${participant.state.toLowerCase()}`}>
                      {participant.state}
                    </div>
                  </div>

                  <div className="participant-details">
                    <div className="detail-row">
                      <span className="detail-label">Joined:</span>
                      <span className="detail-value">{formatDate(participant.joinedAt)}</span>
                    </div>
                    {participant.leftAt && (
                      <div className="detail-row">
                        <span className="detail-label">Left:</span>
                        <span className="detail-value">{formatDate(participant.leftAt)}</span>
                      </div>
                    )}
                    {participant.disconnectReason && (
                      <div className="detail-row">
                        <span className="detail-label">Reason:</span>
                        <span className="detail-value">{participant.disconnectReason}</span>
                      </div>
                    )}
                    <div className="detail-row">
                      <span className="detail-label">Publisher:</span>
                      <span className="detail-value">{participant.isPublisher ? 'Yes' : 'No'}</span>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {activeTab === 'tracks' && (
          <div className="tracks-list">
            {tracks.length === 0 ? (
              <div className="empty-state-small">
                <LuActivity className="empty-icon" />
                <p>No tracks</p>
              </div>
            ) : (
              <div className="tracks-grid">
                {tracks.map((track) => (
                  <div key={track.id} className={`track-card track-${track.status}`}>
                    <div className="track-header">
                      <div className="track-icon-wrapper">
                        {getTrackIcon(track)}
                      </div>
                      <div className="track-title">
                        <h4>{track.source}</h4>
                        <p className="track-type">{track.type || 'AUDIO'}</p>
                      </div>
                      <div className={`track-status status-${track.status}`}>
                        {track.status}
                      </div>
                    </div>

                    <div className="track-details">
                      <div className="track-detail-row">
                        <span className="track-label">Codec:</span>
                        <span className="track-value">{track.mimeType}</span>
                      </div>
                      {track.type === 'VIDEO' && track.width && track.height && (
                        <div className="track-detail-row">
                          <span className="track-label">Resolution:</span>
                          <span className="track-value">{track.width} × {track.height}</span>
                        </div>
                      )}
                      {track.simulcast && (
                        <div className="track-detail-row">
                          <span className="track-label">Simulcast:</span>
                          <span className="track-value track-badge">Enabled</span>
                        </div>
                      )}
                      <div className="track-detail-row">
                        <span className="track-label">Published:</span>
                        <span className="track-value">{formatDate(track.publishedAt)}</span>
                      </div>
                      {track.unpublishedAt && (
                        <div className="track-detail-row">
                          <span className="track-label">Unpublished:</span>
                          <span className="track-value">{formatDate(track.unpublishedAt)}</span>
                        </div>
                      )}
                    </div>

                    <div className="track-footer">
                      <span className="track-sid">{track.sid}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default RoomDetails;
