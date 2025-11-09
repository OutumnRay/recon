import { useEffect, useRef, useState } from 'react';
import {
  Room,
  RoomEvent,
  RemoteTrack,
  RemoteTrackPublication,
  RemoteParticipant,
  LocalTrackPublication,
  LocalParticipant,
  Track,
  VideoPresets,
  Participant,
} from 'livekit-client';

// Автоматическое определение backend URL
const BACKEND_URL = import.meta.env.PROD
  ? 'https://meet.recontext.online'  // Production
  : 'http://localhost:3000';          // Development

interface TokenResponse {
  token: string;
  url: string;
  roomName: string;
  participantName: string;
}

export function LiveKitRoom() {
  const [room] = useState(() => new Room({
    adaptiveStream: true,
    dynacast: true,
    videoCaptureDefaults: {
      resolution: VideoPresets.h720.resolution,
    },
  }));

  const [isConnected, setIsConnected] = useState(false);
  const [isJoining, setIsJoining] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [participants, setParticipants] = useState<Map<string, RemoteParticipant>>(new Map());
  const [activeSpeakers, setActiveSpeakers] = useState<Participant[]>([]);
  const [isCameraEnabled, setIsCameraEnabled] = useState(false);
  const [isMicEnabled, setIsMicEnabled] = useState(false);

  // Form state
  const [userName, setUserName] = useState('');
  const [roomName, setRoomName] = useState('my-room');
  const [hasJoined, setHasJoined] = useState(false);

  const videoContainerRef = useRef<HTMLDivElement>(null);
  const localVideoRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!hasJoined) return;

    const handleTrackSubscribed = (
      track: RemoteTrack,
      _publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      console.log('Track subscribed:', track.kind, 'from', participant.identity);

      if (track.kind === Track.Kind.Video || track.kind === Track.Kind.Audio) {
        const element = track.attach();
        element.id = `${participant.sid}-${track.kind}`;

        if (track.kind === Track.Kind.Video) {
          element.style.width = '100%';
          element.style.height = 'auto';
          element.style.borderRadius = '8px';
        }

        const participantContainer = document.getElementById(`participant-${participant.sid}`);
        if (participantContainer) {
          participantContainer.appendChild(element);
        } else if (videoContainerRef.current) {
          const container = document.createElement('div');
          container.id = `participant-${participant.sid}`;
          container.style.marginBottom = '20px';
          container.innerHTML = `<h3 style="color: white;">${participant.identity || participant.sid}</h3>`;
          container.appendChild(element);
          videoContainerRef.current.appendChild(container);
        }
      }
    };

    const handleTrackUnsubscribed = (
      track: RemoteTrack,
      _publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      console.log('Track unsubscribed:', track.kind, 'from', participant.identity);
      track.detach();

      const element = document.getElementById(`${participant.sid}-${track.kind}`);
      if (element) {
        element.remove();
      }
    };

    const handleLocalTrackPublished = (
      publication: LocalTrackPublication,
      _participant: LocalParticipant,
    ) => {
      console.log('Local track published:', publication.kind);

      if (publication.track && publication.kind === Track.Kind.Video && localVideoRef.current) {
        const element = publication.track.attach();
        element.style.width = '100%';
        element.style.height = 'auto';
        element.style.borderRadius = '8px';

        // Clear previous content and add new element
        localVideoRef.current.innerHTML = '';
        localVideoRef.current.appendChild(element);
      }
    };

    const handleLocalTrackUnpublished = (
      publication: LocalTrackPublication,
      _participant: LocalParticipant,
    ) => {
      console.log('Local track unpublished:', publication.kind);
      if (publication.track) {
        publication.track.detach();
      }
      if (publication.kind === Track.Kind.Video && localVideoRef.current) {
        // Clear video elements when camera is turned off
        while (localVideoRef.current.firstChild) {
          localVideoRef.current.removeChild(localVideoRef.current.firstChild);
        }
      }
    };

    const handleActiveSpeakerChange = (speakers: Participant[]) => {
      console.log('Active speakers changed:', speakers.map(s => s.identity));
      setActiveSpeakers(speakers);
    };

    const handleDisconnect = () => {
      console.log('Disconnected from room');
      setIsConnected(false);
      setHasJoined(false);
      setParticipants(new Map());
      setIsCameraEnabled(false);
      setIsMicEnabled(false);
    };

    const handleParticipantConnected = (participant: RemoteParticipant) => {
      console.log('Participant connected:', participant.identity);
      setParticipants(prev => new Map(prev).set(participant.sid, participant));
    };

    const handleParticipantDisconnected = (participant: RemoteParticipant) => {
      console.log('Participant disconnected:', participant.identity);
      setParticipants(prev => {
        const newMap = new Map(prev);
        newMap.delete(participant.sid);
        return newMap;
      });

      const participantContainer = document.getElementById(`participant-${participant.sid}`);
      if (participantContainer) {
        participantContainer.remove();
      }
    };

    // Set up event listeners
    room
      .on(RoomEvent.TrackSubscribed, handleTrackSubscribed)
      .on(RoomEvent.TrackUnsubscribed, handleTrackUnsubscribed)
      .on(RoomEvent.LocalTrackPublished, handleLocalTrackPublished)
      .on(RoomEvent.LocalTrackUnpublished, handleLocalTrackUnpublished)
      .on(RoomEvent.ActiveSpeakersChanged, handleActiveSpeakerChange)
      .on(RoomEvent.Disconnected, handleDisconnect)
      .on(RoomEvent.ParticipantConnected, handleParticipantConnected)
      .on(RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);

    // Cleanup
    return () => {
      room.removeAllListeners();
    };
  }, [room, hasJoined]);

  const getToken = async (room: string, name: string): Promise<TokenResponse> => {
    const response = await fetch(`${BACKEND_URL}/getToken?room=${encodeURIComponent(room)}&name=${encodeURIComponent(name)}`);

    if (!response.ok) {
      throw new Error('Failed to get token from server');
    }

    return response.json();
  };

  const handleJoinRoom = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!userName.trim()) {
      setError('Please enter your name');
      return;
    }

    if (!roomName.trim()) {
      setError('Please enter a room name');
      return;
    }

    setIsJoining(true);
    setError(null);

    try {
      console.log('Getting token from server...');
      const tokenData = await getToken(roomName, userName);

      console.log('Connecting to LiveKit server...');

      // Pre-warm connection
      await room.prepareConnection(tokenData.url, tokenData.token);

      // Connect
      await room.connect(tokenData.url, tokenData.token);

      console.log('Connected to room:', room.name);
      setIsConnected(true);
      setHasJoined(true);
      setError(null);

      // Get existing participants
      room.remoteParticipants.forEach((participant) => {
        setParticipants(prev => new Map(prev).set(participant.sid, participant));
      });

    } catch (err) {
      console.error('Failed to connect:', err);
      setError(err instanceof Error ? err.message : 'Failed to connect to room');
      setIsJoining(false);
    } finally {
      setIsJoining(false);
    }
  };

  const handleLeaveRoom = () => {
    room.disconnect();
    setHasJoined(false);
    setIsConnected(false);
    setParticipants(new Map());
    setIsCameraEnabled(false);
    setIsMicEnabled(false);

    // Clear video containers properly
    if (videoContainerRef.current) {
      while (videoContainerRef.current.firstChild) {
        videoContainerRef.current.removeChild(videoContainerRef.current.firstChild);
      }
    }
    if (localVideoRef.current) {
      while (localVideoRef.current.firstChild) {
        localVideoRef.current.removeChild(localVideoRef.current.firstChild);
      }
    }
  };

  const toggleCamera = async () => {
    try {
      if (isCameraEnabled) {
        await room.localParticipant.setCameraEnabled(false);
        setIsCameraEnabled(false);
      } else {
        await room.localParticipant.setCameraEnabled(true);
        setIsCameraEnabled(true);
      }
    } catch (err) {
      console.error('Failed to toggle camera:', err);
      setError(err instanceof Error ? err.message : 'Failed to toggle camera');
    }
  };

  const toggleMicrophone = async () => {
    try {
      if (isMicEnabled) {
        await room.localParticipant.setMicrophoneEnabled(false);
        setIsMicEnabled(false);
      } else {
        await room.localParticipant.setMicrophoneEnabled(true);
        setIsMicEnabled(true);
      }
    } catch (err) {
      console.error('Failed to toggle microphone:', err);
      setError(err instanceof Error ? err.message : 'Failed to toggle microphone');
    }
  };

  const enableBoth = async () => {
    try {
      await room.localParticipant.enableCameraAndMicrophone();
      setIsCameraEnabled(true);
      setIsMicEnabled(true);
    } catch (err) {
      console.error('Failed to enable camera and microphone:', err);
      setError(err instanceof Error ? err.message : 'Failed to enable camera and microphone');
    }
  };

  // If not joined, show the join form
  if (!hasJoined) {
    return (
      <div style={{
        padding: '20px',
        maxWidth: '500px',
        margin: '100px auto',
        backgroundColor: 'rgba(255, 255, 255, 0.95)',
        borderRadius: '16px',
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.2)',
      }}>
        <h1 style={{ color: '#333', textAlign: 'center', marginBottom: '30px' }}>
          Join LiveKit Room
        </h1>

        {error && (
          <div style={{
            padding: '12px',
            marginBottom: '20px',
            backgroundColor: '#fee',
            color: '#c33',
            borderRadius: '8px',
            border: '1px solid #fcc',
          }}>
            {error}
          </div>
        )}

        <form onSubmit={handleJoinRoom}>
          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'block', marginBottom: '8px', color: '#333', fontWeight: '500' }}>
              Your Name *
            </label>
            <input
              type="text"
              value={userName}
              onChange={(e) => setUserName(e.target.value)}
              placeholder="Enter your name"
              required
              style={{
                width: '100%',
                padding: '12px',
                fontSize: '16px',
                border: '2px solid #ddd',
                borderRadius: '8px',
                boxSizing: 'border-box',
                transition: 'border-color 0.2s',
              }}
              onFocus={(e) => e.target.style.borderColor = '#667eea'}
              onBlur={(e) => e.target.style.borderColor = '#ddd'}
            />
          </div>

          <div style={{ marginBottom: '20px' }}>
            <label style={{ display: 'block', marginBottom: '8px', color: '#333', fontWeight: '500' }}>
              Room Name *
            </label>
            <input
              type="text"
              value={roomName}
              onChange={(e) => setRoomName(e.target.value)}
              placeholder="Enter room name"
              required
              style={{
                width: '100%',
                padding: '12px',
                fontSize: '16px',
                border: '2px solid #ddd',
                borderRadius: '8px',
                boxSizing: 'border-box',
                transition: 'border-color 0.2s',
              }}
              onFocus={(e) => e.target.style.borderColor = '#667eea'}
              onBlur={(e) => e.target.style.borderColor = '#ddd'}
            />
          </div>

          <button
            type="submit"
            disabled={isJoining}
            style={{
              width: '100%',
              padding: '14px',
              fontSize: '18px',
              fontWeight: '600',
              backgroundColor: isJoining ? '#999' : '#667eea',
              color: 'white',
              border: 'none',
              borderRadius: '8px',
              cursor: isJoining ? 'not-allowed' : 'pointer',
              transition: 'background-color 0.2s',
            }}
            onMouseEnter={(e) => {
              if (!isJoining) e.currentTarget.style.backgroundColor = '#5568d3';
            }}
            onMouseLeave={(e) => {
              if (!isJoining) e.currentTarget.style.backgroundColor = '#667eea';
            }}
          >
            {isJoining ? 'Joining...' : 'Join Room'}
          </button>
        </form>

        <div style={{ marginTop: '20px', textAlign: 'center', color: '#666', fontSize: '14px' }}>
          <p>Server: wss://video.recontext.online</p>
        </div>
      </div>
    );
  }

  // If joined, show the room interface
  return (
    <div style={{ padding: '20px', maxWidth: '1400px', margin: '0 auto', color: 'white' }}>
      <div style={{
        display: 'flex',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: '20px',
        backgroundColor: 'rgba(255, 255, 255, 0.1)',
        padding: '20px',
        borderRadius: '12px',
      }}>
        <div>
          <h1 style={{ margin: '0 0 10px 0' }}>Room: {room.name}</h1>
          <p style={{ margin: 0 }}>
            <strong>Status:</strong>{' '}
            {isConnected ? (
              <span style={{ color: '#4ade80' }}>Connected as {userName}</span>
            ) : (
              <span style={{ color: '#fbbf24' }}>Connecting...</span>
            )}
          </p>
          <p style={{ margin: '5px 0 0 0' }}>
            <strong>Participants:</strong> {participants.size + 1}
          </p>
          {activeSpeakers.length > 0 && (
            <p style={{ margin: '5px 0 0 0' }}>
              <strong>Speaking:</strong>{' '}
              {activeSpeakers.map(s => s.identity || s.sid).join(', ')}
            </p>
          )}
        </div>

        <button
          onClick={handleLeaveRoom}
          style={{
            padding: '12px 24px',
            fontSize: '16px',
            backgroundColor: '#ef4444',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            cursor: 'pointer',
            fontWeight: '600',
          }}
        >
          Leave Room
        </button>
      </div>

      {error && (
        <div style={{
          padding: '12px',
          marginBottom: '20px',
          backgroundColor: 'rgba(254, 226, 226, 0.95)',
          color: '#991b1b',
          borderRadius: '8px',
        }}>
          {error}
        </div>
      )}

      <div style={{ marginBottom: '20px' }}>
        <button
          onClick={enableBoth}
          style={{
            padding: '12px 24px',
            marginRight: '10px',
            cursor: 'pointer',
            backgroundColor: '#10b981',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            fontWeight: '600',
            fontSize: '16px',
          }}
          disabled={!isConnected}
        >
          Enable Camera & Mic
        </button>

        <button
          onClick={toggleCamera}
          style={{
            padding: '12px 24px',
            marginRight: '10px',
            cursor: 'pointer',
            backgroundColor: isCameraEnabled ? '#ef4444' : '#3b82f6',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            fontWeight: '600',
            fontSize: '16px',
          }}
          disabled={!isConnected}
        >
          {isCameraEnabled ? '📹 Disable Camera' : '📹 Enable Camera'}
        </button>

        <button
          onClick={toggleMicrophone}
          style={{
            padding: '12px 24px',
            cursor: 'pointer',
            backgroundColor: isMicEnabled ? '#ef4444' : '#3b82f6',
            color: 'white',
            border: 'none',
            borderRadius: '8px',
            fontWeight: '600',
            fontSize: '16px',
          }}
          disabled={!isConnected}
        >
          {isMicEnabled ? '🎤 Disable Mic' : '🎤 Enable Mic'}
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(350px, 1fr))', gap: '20px' }}>
        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.1)',
          padding: '20px',
          borderRadius: '12px',
        }}>
          <h2 style={{ marginTop: 0 }}>Your Video</h2>
          <div
            style={{
              backgroundColor: 'rgba(0, 0, 0, 0.3)',
              minHeight: '250px',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              position: 'relative',
            }}
          >
            {!isCameraEnabled && (
              <p style={{
                position: 'absolute',
                margin: 0,
                color: 'white',
                zIndex: 1
              }}>
                Camera is off
              </p>
            )}
            <div
              ref={localVideoRef}
              style={{
                width: '100%',
                height: '100%',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            />
          </div>
        </div>

        <div style={{
          backgroundColor: 'rgba(255, 255, 255, 0.1)',
          padding: '20px',
          borderRadius: '12px',
          gridColumn: participants.size > 0 ? 'auto' : '1 / -1',
        }}>
          <h2 style={{ marginTop: 0 }}>
            Remote Participants {participants.size > 0 && `(${participants.size})`}
          </h2>
          <div
            ref={videoContainerRef}
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
              gap: '20px',
            }}
          >
            {participants.size === 0 && (
              <p style={{ color: '#9ca3af' }}>Waiting for other participants to join...</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
