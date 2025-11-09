import { useEffect, useState } from 'react';
import { liveKitApi } from '../services/livekit';
import type { Room } from '../services/livekit';
import { LuVideo, LuClock, LuUsers, LuCircleCheck, LuCircleX } from 'react-icons/lu';
import './Rooms.css';

export const Rooms = () => {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [refreshInterval, setRefreshInterval] = useState<number | null>(null);

  const loadRooms = async () => {
    try {
      setError(null);
      const data = await liveKitApi.getRooms(statusFilter || undefined);
      setRooms(data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load rooms');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadRooms();
  }, [statusFilter]);

  // Auto-refresh every 5 seconds for active rooms
  useEffect(() => {
    if (statusFilter === '' || statusFilter === 'active') {
      const interval = window.setInterval(() => {
        loadRooms();
      }, 5000);
      setRefreshInterval(interval);

      return () => {
        if (interval) {
          clearInterval(interval);
        }
      };
    } else {
      if (refreshInterval) {
        clearInterval(refreshInterval);
        setRefreshInterval(null);
      }
    }
  }, [statusFilter]);

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleString();
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

  const handleRoomClick = (roomSid: string) => {
    window.location.href = `/rooms/${roomSid}`;
  };

  if (loading) {
    return (
      <div className="rooms-container">
        <div className="loading">Loading rooms...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="rooms-container">
        <div className="error-message">
          <LuCircleX className="error-icon" />
          <p>{error}</p>
          <button onClick={loadRooms} className="btn-retry">Retry</button>
        </div>
      </div>
    );
  }

  return (
    <div className="rooms-container">
      <div className="rooms-header">
        <div className="header-left">
          <h2 className="rooms-title">
            <LuVideo className="title-icon" />
            Meeting Rooms
          </h2>
          <p className="rooms-subtitle">
            Manage and monitor LiveKit conference rooms
          </p>
        </div>
        <div className="header-right">
          <button onClick={loadRooms} className="btn-refresh">
            Refresh
          </button>
        </div>
      </div>

      <div className="rooms-filters">
        <div className="filter-group">
          <label htmlFor="status-filter">Status:</label>
          <select
            id="status-filter"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="filter-select"
          >
            <option value="">All</option>
            <option value="active">Active</option>
            <option value="finished">Finished</option>
          </select>
        </div>
        {(statusFilter === '' || statusFilter === 'active') && (
          <div className="auto-refresh-indicator">
            <span className="pulse-dot"></span>
            Auto-refreshing every 5s
          </div>
        )}
      </div>

      {rooms.length === 0 ? (
        <div className="empty-state">
          <LuVideo className="empty-icon" />
          <h3>No rooms found</h3>
          <p>No meeting rooms match your current filters</p>
        </div>
      ) : (
        <div className="rooms-grid">
          {rooms.map((room) => (
            <div
              key={room.id}
              className={`room-card ${room.status}`}
              onClick={() => handleRoomClick(room.sid)}
            >
              <div className="room-card-header">
                <div className="room-name">
                  <LuVideo className="room-icon" />
                  <h3>{room.name}</h3>
                </div>
                <div className={`room-status status-${room.status}`}>
                  {room.status === 'active' ? (
                    <>
                      <span className="status-dot"></span>
                      Active
                    </>
                  ) : (
                    <>
                      <LuCircleCheck className="status-icon" />
                      Finished
                    </>
                  )}
                </div>
              </div>

              <div className="room-card-body">
                <div className="room-info-row">
                  <span className="info-label">
                    <LuClock className="info-icon" />
                    Started:
                  </span>
                  <span className="info-value">{formatDate(room.startedAt)}</span>
                </div>

                <div className="room-info-row">
                  <span className="info-label">
                    <LuClock className="info-icon" />
                    Duration:
                  </span>
                  <span className="info-value">{formatDuration(room.startedAt, room.finishedAt)}</span>
                </div>

                <div className="room-info-row">
                  <span className="info-label">Room ID:</span>
                  <span className="info-value room-sid">{room.sid}</span>
                </div>
              </div>

              <div className="room-card-footer">
                <button className="btn-view-details">
                  <LuUsers className="btn-icon" />
                  View Details
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default Rooms;
