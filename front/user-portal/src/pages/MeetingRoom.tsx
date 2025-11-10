import { useEffect, useRef, useState, useCallback } from 'react';
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
import { useTranslation } from 'react-i18next';
import { getMeeting, getMeetingToken } from '../services/meetings';
import type { MeetingTokenResponse } from '../types/meeting';
import {
  LuMic,
  LuMicOff,
  LuVideo,
  LuVideoOff,
  LuMaximize2,
  LuMinimize2,
  LuMenu,
  LuLogOut,
} from 'react-icons/lu';
import './MeetingRoom.css';

export default function MeetingRoom() {
  const { meetingId } = useParams<{ meetingId: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();

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
  const [meetingTitle, setMeetingTitle] = useState('');
  const [stageParticipantId, setStageParticipantId] = useState<string | null>(null);
  const [isFullscreen, setIsFullscreen] = useState(false);
  const [isParticipantsCollapsed, setIsParticipantsCollapsed] = useState(false);
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);

  const videoContainerRef = useRef<HTMLDivElement>(null);
  const localPreviewRef = useRef<HTMLDivElement>(null);
  const stageVideoRef = useRef<HTMLDivElement>(null);

  const participantVideoTracks = useRef<Map<string, RemoteTrack>>(new Map());
  const stageTrackRef = useRef<Track | null>(null);
  const stageElementRef = useRef<HTMLMediaElement | null>(null);
  const localPreviewTrackRef = useRef<Track | null>(null);
  const localPreviewElementRef = useRef<HTMLMediaElement | null>(null);

  const getInitials = (value: string) => {
    const parts = value.trim().split(' ');
    const letters = parts.length === 1
      ? parts[0].slice(0, 2)
      : `${parts[0][0]}${parts[1][0]}`;
    return letters.toUpperCase();
  };

  const detachStageVideo = useCallback(() => {
    if (stageTrackRef.current && stageElementRef.current) {
      stageTrackRef.current.detach(stageElementRef.current);
      stageElementRef.current.remove();
    }
    stageTrackRef.current = null;
    stageElementRef.current = null;
  }, []);

  const detachLocalPreview = useCallback(() => {
    if (localPreviewTrackRef.current && localPreviewElementRef.current) {
      localPreviewTrackRef.current.detach(localPreviewElementRef.current);
      localPreviewElementRef.current.remove();
    }
    localPreviewTrackRef.current = null;
    localPreviewElementRef.current = null;
  }, []);

  const renderLocalPreview = useCallback(() => {
    const container = localPreviewRef.current;
    if (!container) return;

    detachLocalPreview();
    container.innerHTML = '';

    const publication = room.localParticipant.getTrackPublication(Track.Source.Camera);
    if (publication?.track) {
      const element = publication.track.attach();
      element.classList.add('meeting-video-element');
      container.appendChild(element);
      localPreviewTrackRef.current = publication.track;
      localPreviewElementRef.current = element;
    }
  }, [detachLocalPreview, room]);

  const renderStageVideo = useCallback((preferredId?: string) => {
    const container = stageVideoRef.current;
    if (!container) return;

    detachStageVideo();
    container.innerHTML = '';

    const targetId = preferredId || stageParticipantId || room.localParticipant?.sid || null;

    let track: Track | null = null;
    let element: HTMLMediaElement | null = null;

    if (targetId && targetId === room.localParticipant?.sid) {
      const publication = room.localParticipant.getTrackPublication(Track.Source.Camera);
      if (publication?.track) {
        track = publication.track;
        element = publication.track.attach();
      }
    } else if (targetId) {
      const remoteTrack = participantVideoTracks.current.get(targetId);
      if (remoteTrack) {
        track = remoteTrack;
        element = remoteTrack.attach();
      }
    }

    if (element) {
      element.classList.add('meeting-video-element');
      container.appendChild(element);
      stageTrackRef.current = track;
      stageElementRef.current = element;
    } else {
      const placeholder = document.createElement('div');
      placeholder.className = 'stage-placeholder';
      placeholder.textContent = t('meetingRoom.waitingForParticipants');
      container.appendChild(placeholder);
    }
  }, [detachStageVideo, room, stageParticipantId, t]);

  const updateSidebarHighlight = useCallback((selectedId: string | null) => {
    if (!videoContainerRef.current) return;
    const tiles = videoContainerRef.current.querySelectorAll<HTMLElement>('.remote-participant-tile');
    tiles.forEach((tile) => {
      tile.classList.toggle('active', !!selectedId && tile.dataset.participant === selectedId);
    });
  }, []);

  const ensureParticipantTile = (participant: RemoteParticipant) => {
    let container = document.getElementById(`participant-${participant.sid}`) as HTMLElement | null;
    if (!container) {
      container = document.createElement('div');
      container.id = `participant-${participant.sid}`;
      container.className = 'remote-participant-tile';
      container.dataset.participant = participant.sid;
      container.addEventListener('click', () => {
        setStageParticipantId(participant.sid);
      });

      const header = document.createElement('div');
      header.className = 'remote-participant-header';

      const avatar = document.createElement('div');
      avatar.className = 'participant-avatar';
      avatar.textContent = getInitials(participant.identity || participant.sid);
      header.appendChild(avatar);

      const name = document.createElement('div');
      name.className = 'remote-participant-name';
      name.textContent = participant.identity || participant.sid;
      header.appendChild(name);

      container.appendChild(header);

      const videoSlot = document.createElement('div');
      videoSlot.className = 'remote-participant-video';
      videoSlot.dataset.slot = participant.sid;
      container.appendChild(videoSlot);

      videoContainerRef.current?.appendChild(container);
    }
    return container;
  };

  const attachTrackToTile = (participant: RemoteParticipant, element: HTMLElement) => {
    const container = ensureParticipantTile(participant);
    const videoSlot = container.querySelector<HTMLDivElement>('.remote-participant-video');

    if (videoSlot) {
      videoSlot.innerHTML = '';
      videoSlot.appendChild(element);
    } else {
      container.appendChild(element);
    }
  };

  useEffect(() => {
    if (!meetingId) {
      setError(t('meetingRoom.errors.missingMeetingId'));
      setIsJoining(false);
      return;
    }

    const loadMeetingTitle = async () => {
      try {
        const meeting = await getMeeting(meetingId);
        setMeetingTitle(meeting.title);
      } catch (err) {
        console.error('Failed to load meeting details:', err);
      }
    };

    loadMeetingTitle();
  }, [meetingId, t]);

  useEffect(() => {
    const pageTitle = meetingTitle || tokenData?.roomName || t('meetingRoom.pageTitle');
    document.title = `Recontext - ${pageTitle}`;
  }, [meetingTitle, tokenData, t]);

  useEffect(() => {
    if (!isConnected) return;

    const handleTrackSubscribed = (
      track: RemoteTrack,
      _publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      if (track.kind === Track.Kind.Video || track.kind === Track.Kind.Audio) {
        const element = track.attach();
        element.id = `${participant.sid}-${track.kind}`;

        if (track.kind === Track.Kind.Video) {
          participantVideoTracks.current.set(participant.sid, track);
          element.classList.add('meeting-video-element');
          attachTrackToTile(participant, element);

          if (stageParticipantId === participant.sid) {
            renderStageVideo(participant.sid);
          }
        } else {
          element.style.display = 'none';
          attachTrackToTile(participant, element);
        }

        updateSidebarHighlight(stageParticipantId);
      }
    };

    const handleTrackUnsubscribed = (
      track: RemoteTrack,
      _publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      track.detach();
      participantVideoTracks.current.delete(participant.sid);

      const container = document.getElementById(`participant-${participant.sid}`);
      const videoSlot = container?.querySelector<HTMLDivElement>('.remote-participant-video');
      if (videoSlot) {
        videoSlot.innerHTML = '';
      }

      if (stageParticipantId === participant.sid) {
        renderStageVideo();
      }
    };

    const handleLocalTrackPublished = (
      publication: LocalTrackPublication,
      _participant: LocalParticipant,
    ) => {
      if (publication.track && publication.kind === Track.Kind.Video) {
        renderLocalPreview();
        if (stageParticipantId === room.localParticipant?.sid) {
          renderStageVideo(room.localParticipant.sid);
        }
      }
    };

    const handleLocalTrackUnpublished = (
      publication: LocalTrackPublication,
      _participant: LocalParticipant,
    ) => {
      if (publication.track && publication.kind === Track.Kind.Video) {
        detachLocalPreview();
        if (stageParticipantId === room.localParticipant?.sid) {
          renderStageVideo();
        }
      }
    };

    const handleActiveSpeakerChange = (speakers: Participant[]) => {
      setActiveSpeakers(speakers);
    };

    const handleDisconnect = () => {
      setIsConnected(false);
      setParticipants(new Map());
      setIsCameraEnabled(false);
      setIsMicEnabled(false);
      participantVideoTracks.current.clear();
      renderStageVideo();
      detachLocalPreview();
    };

    const handleParticipantConnected = (participant: RemoteParticipant) => {
      setParticipants(prev => new Map(prev).set(participant.sid, participant));
      ensureParticipantTile(participant);
    };

    const handleParticipantDisconnected = (participant: RemoteParticipant) => {
      setParticipants(prev => {
        const next = new Map(prev);
        next.delete(participant.sid);
        return next;
      });

      participantVideoTracks.current.delete(participant.sid);
      const container = document.getElementById(`participant-${participant.sid}`);
      if (container) {
        container.remove();
      }

      if (stageParticipantId === participant.sid) {
        renderStageVideo();
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
  }, [
    detachLocalPreview,
    renderLocalPreview,
    renderStageVideo,
    room,
    stageParticipantId,
    updateSidebarHighlight,
  ]);

  useEffect(() => {
    if (!meetingId) {
      setError(t('meetingRoom.errors.missingMeetingId'));
      setIsJoining(false);
      return;
    }

    const joinRoom = async () => {
      try {
        const data = await getMeetingToken(meetingId);
        setTokenData(data);

        await room.prepareConnection(data.url, data.token);
        await room.connect(data.url, data.token);

        setIsConnected(true);
        setError(null);
        setStageParticipantId(current => current ?? room.localParticipant?.sid ?? null);

        room.remoteParticipants.forEach((participant) => {
          setParticipants(prev => new Map(prev).set(participant.sid, participant));
          ensureParticipantTile(participant);
        });
      } catch (err) {
        console.error('Failed to connect:', err);
        setError(err instanceof Error ? err.message : t('meetingRoom.errors.connect'));
      } finally {
        setIsJoining(false);
      }
    };

    joinRoom();

    return () => {
      room.disconnect();
    };
  }, [meetingId, room, t]);

  useEffect(() => {
    if (stageParticipantId) return;
    const fallback = participants.keys().next().value || room.localParticipant?.sid || null;
    if (fallback) {
      setStageParticipantId(fallback);
    }
  }, [participants, room, stageParticipantId]);

  useEffect(() => {
    if (!stageParticipantId) return;
    const isLocal = stageParticipantId === room.localParticipant?.sid;
    if (!isLocal && !participants.has(stageParticipantId)) {
      const fallback = participants.keys().next().value || room.localParticipant?.sid || null;
      if (fallback && fallback !== stageParticipantId) {
        setStageParticipantId(fallback);
      }
    }
  }, [participants, room, stageParticipantId]);

  useEffect(() => {
    renderStageVideo();
  }, [renderStageVideo, stageParticipantId, isCameraEnabled]);

  useEffect(() => {
    updateSidebarHighlight(stageParticipantId);
  }, [stageParticipantId, updateSidebarHighlight]);

  const confirmLeave = () => {
    room.disconnect();
    navigate('/dashboard/meetings');
  };

  const toggleCamera = async () => {
    try {
      if (isCameraEnabled) {
        await room.localParticipant.setCameraEnabled(false);
        setIsCameraEnabled(false);
        detachLocalPreview();
      } else {
        await room.localParticipant.setCameraEnabled(true);
        setIsCameraEnabled(true);
        renderLocalPreview();
      }

      if (stageParticipantId === room.localParticipant?.sid) {
        renderStageVideo(room.localParticipant.sid);
      }
    } catch (err) {
      console.error('Failed to toggle camera:', err);
      setError(err instanceof Error ? err.message : t('meetingRoom.errors.toggleCamera'));
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
      setError(err instanceof Error ? err.message : t('meetingRoom.errors.toggleMicrophone'));
    }
  };

  if (isJoining) {
    return (
      <div className="meeting-room-state-card">
        <h1>{t('meetingRoom.connectingTitle')}</h1>
        <p>{t('meetingRoom.connectingDescription')}</p>
      </div>
    );
  }

  if (error && !isConnected) {
    return (
      <div className="meeting-room-state-card meeting-room-state-error">
        <h1>{t('meetingRoom.unableToJoin')}</h1>
        <div className="state-alert">
          {error}
        </div>
        <button
          onClick={() => navigate('/dashboard/meetings')}
          className="btn btn-primary state-action"
        >
          {t('meetings.backToList')}
        </button>
      </div>
    );
  }

  const layoutClass = `meeting-layout ${isParticipantsCollapsed ? 'sidebar-collapsed' : ''}`;
  const speakingNames = activeSpeakers.map(s => s.identity || s.sid).join(', ');
  const statusText = isConnected
    ? t('meetingRoom.connectedAs', { name: tokenData?.participantName || t('meetingRoom.guest') })
    : t('meetingRoom.connecting');

  return (
    <div className={`meeting-room-page ${isFullscreen ? 'fullscreen' : ''}`}>
      {!isFullscreen && (
        <div className="meeting-room-header">
          <div className="meeting-room-header-info">
            <h1>
              {meetingTitle || tokenData?.roomName || t('meetingRoom.pageTitle')}
              <span className={`meeting-status ${isConnected ? 'connected' : 'pending'}`}>
                {' '}
                ({statusText})
              </span>
            </h1>
            <div className="meeting-room-header-meta">
              {tokenData?.forceEndAt && (
                <p className="meeting-room-meta meeting-warning">
                  {t('meetingRoom.endsAt')}: {new Date(tokenData.forceEndAt).toLocaleString(undefined, { hour12: false })}
                </p>
              )}
            </div>
          </div>
          <div className="meeting-room-header-actions">
            <button
              onClick={() => setIsFullscreen(prev => !prev)}
              className="icon-circle-button"
              aria-label={isFullscreen ? t('meetingRoom.exitFullscreen') : t('meetingRoom.enterFullscreen')}
              title={isFullscreen ? t('meetingRoom.exitFullscreen') : t('meetingRoom.enterFullscreen')}
            >
              {isFullscreen ? <LuMinimize2 /> : <LuMaximize2 />}
            </button>
            <button
              onClick={() => setShowLeaveConfirm(true)}
              className="icon-circle-button danger"
              aria-label={t('meetingRoom.leaveRoom')}
              title={t('meetingRoom.leaveRoom')}
            >
              <LuLogOut />
            </button>
          </div>
        </div>
      )}

      {error && isConnected && (
        <div className="alert alert-error meeting-room-inline-alert">
          {error}
        </div>
      )}

      <div className={layoutClass}>
        <div className="stage-section">
          <div className="stage-video-wrapper">
            {speakingNames && (
              <div className="stage-speaking-indicator">
                <span>{t('meetingRoom.speakingNow')}</span>
                <strong>{speakingNames}</strong>
              </div>
            )}
            <div className="stage-video" ref={stageVideoRef} />
            <div className={`local-preview ${isCameraEnabled ? 'visible' : ''}`}>
              <div className="local-preview-video" ref={localPreviewRef} />
            </div>
            <div className="stage-controls">
              <button
                onClick={toggleCamera}
                className="icon-circle-button"
                aria-label={isCameraEnabled ? t('meetingRoom.disableCamera') : t('meetingRoom.enableCamera')}
                title={isCameraEnabled ? t('meetingRoom.disableCamera') : t('meetingRoom.enableCamera')}
                disabled={!isConnected}
              >
                {isCameraEnabled ? <LuVideo /> : <LuVideoOff />}
              </button>
              <button
                onClick={toggleMicrophone}
                className="icon-circle-button"
                aria-label={isMicEnabled ? t('meetingRoom.disableMic') : t('meetingRoom.enableMic')}
                title={isMicEnabled ? t('meetingRoom.disableMic') : t('meetingRoom.enableMic')}
                disabled={!isConnected}
              >
                {isMicEnabled ? <LuMic /> : <LuMicOff />}
              </button>
              <button
                onClick={() => setIsFullscreen(prev => !prev)}
                className="icon-circle-button"
                aria-label={isFullscreen ? t('meetingRoom.exitFullscreen') : t('meetingRoom.enterFullscreen')}
                title={isFullscreen ? t('meetingRoom.exitFullscreen') : t('meetingRoom.enterFullscreen')}
              >
                {isFullscreen ? <LuMinimize2 /> : <LuMaximize2 />}
              </button>
            </div>
          </div>
        </div>

        <aside className={`participant-sidebar ${isParticipantsCollapsed ? 'collapsed' : ''}`}>
          <div className="participant-sidebar-header">
            {!isParticipantsCollapsed && (
              <div className="participant-sidebar-title">
                <span>{t('meetingRoom.remoteParticipants')}</span>
                <span className="participant-count-pill">
                  {participants.size + 1}
                </span>
              </div>
            )}
            <button
              className="icon-circle-button"
              onClick={() => setIsParticipantsCollapsed(prev => !prev)}
              aria-label={t('meetingRoom.toggleParticipants')}
              title={t('meetingRoom.toggleParticipants')}
            >
              <LuMenu />
            </button>
          </div>
          <div className="participant-sidebar-list" ref={videoContainerRef} />
          {participants.size === 0 && (
            <p className="empty-participants-text">{t('meetingRoom.waitingForParticipants')}</p>
          )}
        </aside>
      </div>

      {showLeaveConfirm && (
        <div className="meeting-room-modal-backdrop">
          <div className="meeting-room-modal">
            <h3>{t('meetingRoom.leaveConfirmTitle')}</h3>
            <p>{t('meetingRoom.leaveConfirmDescription')}</p>
            <div className="meeting-room-modal-actions">
              <button
                className="btn btn-ghost"
                onClick={() => setShowLeaveConfirm(false)}
              >
                {t('meetingRoom.cancelLeave')}
              </button>
              <button
                className="btn btn-danger"
                onClick={confirmLeave}
              >
                {t('meetingRoom.confirmLeave')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
