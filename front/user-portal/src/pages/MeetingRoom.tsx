import { useEffect, useRef, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
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
import { getMeetingToken } from '../services/meetings';
import type { MeetingTokenResponse } from '../types/meeting';
import './MeetingRoom.css';

export default function MeetingRoom() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();

  // Set page title
  useEffect(() => {
    document.title = 'Recontext - Meeting Room';
  }, []);

  const [room] = useState(() => new Room({
    adaptiveStream: true,
    dynacast: true,
    videoCaptureDefaults: {
      resolution: VideoPresets.h720.resolution,
    },
  }));

  const [isConnected, setIsConnected] = useState(false);
  const [isJoining, setIsJoining] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [participants, setParticipants] = useState<Map<string, RemoteParticipant>>(new Map());
  const [activeSpeakers, setActiveSpeakers] = useState<Participant[]>([]);
  const [isCameraEnabled, setIsCameraEnabled] = useState(false);
  const [isMicEnabled, setIsMicEnabled] = useState(false);
  const [tokenData, setTokenData] = useState<MeetingTokenResponse | null>(null);

  const videoContainerRef = useRef<HTMLDivElement>(null);
  const localVideoRef = useRef<HTMLDivElement>(null);

  // Set up event listeners
  useEffect(() => {
    if (!isConnected) return;

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
          element.classList.add('meeting-video-element');
        }

        let participantContainer = document.getElementById(`participant-${participant.sid}`);
        if (!participantContainer && videoContainerRef.current) {
          participantContainer = document.createElement('div');
          participantContainer.id = `participant-${participant.sid}`;
          participantContainer.className = 'remote-participant-tile';

          const nameEl = document.createElement('h3');
          nameEl.className = 'remote-participant-name';
          nameEl.textContent = participant.identity || participant.sid;

          participantContainer.appendChild(nameEl);
          videoContainerRef.current.appendChild(participantContainer);
        }

        participantContainer?.appendChild(element);
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
        element.classList.add('meeting-video-element');

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

    room
      .on(RoomEvent.TrackSubscribed, handleTrackSubscribed)
      .on(RoomEvent.TrackUnsubscribed, handleTrackUnsubscribed)
      .on(RoomEvent.LocalTrackPublished, handleLocalTrackPublished)
      .on(RoomEvent.LocalTrackUnpublished, handleLocalTrackUnpublished)
      .on(RoomEvent.ActiveSpeakersChanged, handleActiveSpeakerChange)
      .on(RoomEvent.Disconnected, handleDisconnect)
      .on(RoomEvent.ParticipantConnected, handleParticipantConnected)
      .on(RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);

    return () => {
      room.removeAllListeners();
    };
  }, [room, isConnected]);

  // Auto-join on component mount
  useEffect(() => {
    if (!meetingId) {
      setError('Meeting ID is required');
      setIsJoining(false);
      return;
    }

    const joinRoom = async () => {
      try {
        console.log('Getting token from server for meeting:', meetingId);
        const data = await getMeetingToken(meetingId);
        setTokenData(data);

        console.log('Connecting to LiveKit server...');
        await room.prepareConnection(data.url, data.token);
        await room.connect(data.url, data.token);

        console.log('Connected to room:', room.name);
        setIsConnected(true);
        setError(null);

        room.remoteParticipants.forEach((participant) => {
          setParticipants(prev => new Map(prev).set(participant.sid, participant));
        });
      } catch (err) {
        console.error('Failed to connect:', err);
        setError(err instanceof Error ? err.message : 'Failed to connect to room');
      } finally {
        setIsJoining(false);
      }
    };

    joinRoom();

    return () => {
      room.disconnect();
    };
  }, [meetingId, room]);

  const handleLeaveRoom = () => {
    room.disconnect();
    navigate('/dashboard/meetings');
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

  // If joining, show loading state
  if (isJoining) {
    return (
      <div className="meeting-room-state-card">
        <h1>Joining Meeting...</h1>
        <p>Please wait while we connect you to the meeting room.</p>
      </div>
    );
  }

  // If error occurred before joining
  if (error && !isConnected) {
    return (
      <div className="meeting-room-state-card meeting-room-state-error">
        <h1>
          Unable to Join Meeting
        </h1>
        <div className="state-alert">
          {error}
        </div>
        <button
          onClick={() => navigate('/dashboard/meetings')}
          className="btn btn-primary state-action"
        >
          Back to Meetings
        </button>
      </div>
    );
  }

  // Main room interface
  return (
    <div className="meeting-room-page">
      <div className="meeting-room-status-card">
        <div>
          <h1>
            {tokenData?.roomName ? `Room: ${tokenData.roomName}` : 'Meeting Room'}
          </h1>
          <p className="meeting-room-meta">
            <strong>Status:</strong>{' '}
            <span className={`meeting-status ${isConnected ? 'connected' : 'pending'}`}>
              {isConnected
                ? `Connected as ${tokenData?.participantName || 'Guest'}`
                : 'Connecting...'}
            </span>
          </p>
          <p className="meeting-room-meta">
            <strong>Participants:</strong> {participants.size + 1}
          </p>
          {activeSpeakers.length > 0 && (
            <p className="meeting-room-meta">
              <strong>Speaking:</strong>{' '}
              {activeSpeakers.map(s => s.identity || s.sid).join(', ')}
            </p>
          )}
          {tokenData?.forceEndAt && (
            <p className="meeting-room-meta meeting-warning">
              <strong>Meeting will end at:</strong>{' '}
              {new Date(tokenData.forceEndAt).toLocaleString(undefined, { hour12: false })}
            </p>
          )}
        </div>

        <button
          onClick={handleLeaveRoom}
          className="btn btn-danger meeting-leave-btn"
        >
          Leave Room
        </button>
      </div>

      {error && (
        <div className="alert alert-error meeting-room-inline-alert">
          {error}
        </div>
      )}

      <div className="meeting-room-controls">
        <button
          onClick={enableBoth}
          className="btn btn-success meeting-control-btn"
          disabled={!isConnected}
        >
          Enable Camera &amp; Mic
        </button>

        <button
          onClick={toggleCamera}
          className={`btn meeting-control-btn ${isCameraEnabled ? 'btn-danger' : 'btn-primary'}`}
          disabled={!isConnected}
        >
          {isCameraEnabled ? '📹 Disable Camera' : '📹 Enable Camera'}
        </button>

        <button
          onClick={toggleMicrophone}
          className={`btn meeting-control-btn ${isMicEnabled ? 'btn-danger' : 'btn-primary'}`}
          disabled={!isConnected}
        >
          {isMicEnabled ? '🎤 Disable Mic' : '🎤 Enable Mic'}
        </button>
      </div>

      <div className="meeting-panels">
        <div className="meeting-panel">
          <h2 className="panel-heading">Your Video</h2>
          <div className={`video-surface ${!isCameraEnabled ? 'video-surface-muted' : ''}`}>
            {!isCameraEnabled && (
              <p className="video-placeholder">
                Camera is off
              </p>
            )}
            <div
              ref={localVideoRef}
              className="video-feed"
            />
          </div>
        </div>

        <div className={`meeting-panel ${participants.size === 0 ? 'meeting-panel-full' : ''}`}>
          <h2 className="panel-heading">
            Remote Participants {participants.size > 0 && `(${participants.size})`}
          </h2>
          <div
            ref={videoContainerRef}
            className="remote-participants-grid"
          >
            {participants.size === 0 && (
              <p className="empty-participants-text">Waiting for other participants to join...</p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
