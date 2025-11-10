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
  const [volume, setVolume] = useState(100);

  const videoContainerRef = useRef<HTMLDivElement>(null);
  const localPreviewRef = useRef<HTMLDivElement>(null);
  const stageVideoRef = useRef<HTMLDivElement>(null);

  const participantVideoTracks = useRef<Map<string, RemoteTrack>>(new Map());
  const stageTrackRef = useRef<Track | null>(null);
  const stageElementRef = useRef<HTMLMediaElement | null>(null);
  const localPreviewTrackRef = useRef<Track | null>(null);
  const localPreviewElementRef = useRef<HTMLMediaElement | null>(null);
  const volumeRef = useRef<number>(100);

  const getInitials = (value: string) => {
    const parts = value.trim().split(' ');
    const letters = parts.length === 1
      ? parts[0].slice(0, 2)
      : `${parts[0][0]}${parts[1][0]}`;
    return letters.toUpperCase();
  };

  const getParticipantDisplayName = (participant: Participant) => {
    // Use participant.name if available (set by backend), fallback to identity (user ID)
    return participant.name || participant.identity || participant.sid;
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
      const displayName = getParticipantDisplayName(participant);
      avatar.textContent = getInitials(displayName);
      header.appendChild(avatar);

      const name = document.createElement('div');
      name.className = 'remote-participant-name';
      name.textContent = displayName;
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
    if (!isConnected) {
      console.log('[Effect] Not connected yet, skipping event handlers setup');
      return;
    }

    console.log('[Effect] Setting up event handlers for connected room');
    console.log('[Effect] Current remote participants:', room.remoteParticipants.size);

    const handleTrackSubscribed = (
      track: RemoteTrack,
      _publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      const displayName = getParticipantDisplayName(participant);
      console.log(`[Track Subscribed] Participant: ${displayName} (${participant.sid}), Track: ${track.kind}, Source: ${track.source}`);

      if (track.kind === Track.Kind.Video || track.kind === Track.Kind.Audio) {
        const element = track.attach();
        element.id = `${participant.sid}-${track.kind}`;

        if (track.kind === Track.Kind.Video) {
          console.log(`[Video Track] Attaching video for ${displayName}`, element);
          participantVideoTracks.current.set(participant.sid, track);
          element.classList.add('meeting-video-element');
          attachTrackToTile(participant, element);

          if (stageParticipantId === participant.sid) {
            renderStageVideo(participant.sid);
          }
        } else {
          console.log(`[Audio Track] Attaching audio for ${displayName}`, element);
          // Apply current volume to audio element
          if (element instanceof HTMLAudioElement) {
            element.volume = volumeRef.current / 100;
            console.log(`[Audio Track] Set volume to ${volumeRef.current}%`);
          }
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
      console.log(`[Track Unsubscribed] Participant: ${participant.identity}, Track: ${track.kind}`);
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

    const handleTrackPublished = (
      publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      const displayName = getParticipantDisplayName(participant);
      console.log(`[Track Published] Participant: ${displayName}, Track: ${publication.kind}, Source: ${publication.source}`);
      console.log(`  - Is subscribed: ${publication.isSubscribed}`);
      console.log(`  - Is enabled: ${publication.isEnabled}`);
      console.log(`  - Track exists: ${publication.track ? 'yes' : 'no'}`);

      // If track is already subscribed, manually attach it
      if (publication.isSubscribed && publication.track) {
        console.log(`  - Track is already subscribed, manually triggering handleTrackSubscribed`);
        handleTrackSubscribed(publication.track as RemoteTrack, publication, participant);
      }
    };

    const handleTrackUnpublished = (
      publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      const displayName = getParticipantDisplayName(participant);
      console.log(`[Track Unpublished] Participant: ${displayName}, Track: ${publication.kind}`);
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
      const displayName = getParticipantDisplayName(participant);
      console.log(`[Participant Connected] ${displayName} (${participant.sid})`);
      console.log(`[Participant Tracks] Audio publications:`, participant.audioTrackPublications.size);
      console.log(`[Participant Tracks] Video publications:`, participant.videoTrackPublications.size);

      setParticipants(prev => new Map(prev).set(participant.sid, participant));
      ensureParticipantTile(participant);

      // Check if participant already has published tracks that we need to subscribe to
      participant.trackPublications.forEach((publication, trackSid) => {
        console.log(`[Existing Track] Track ${trackSid}: ${publication.kind}, subscribed: ${publication.isSubscribed}, track: ${publication.track ? 'exists' : 'null'}`);

        if (publication.track && !publication.isSubscribed) {
          console.log(`[Manual Subscribe] Attempting to subscribe to ${publication.kind} track`);
        }
      });
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
      .on(RoomEvent.TrackPublished, handleTrackPublished)
      .on(RoomEvent.TrackUnpublished, handleTrackUnpublished)
      .on(RoomEvent.LocalTrackPublished, handleLocalTrackPublished)
      .on(RoomEvent.LocalTrackUnpublished, handleLocalTrackUnpublished)
      .on(RoomEvent.ActiveSpeakersChanged, handleActiveSpeakerChange)
      .on(RoomEvent.Disconnected, handleDisconnect)
      .on(RoomEvent.ParticipantConnected, handleParticipantConnected)
      .on(RoomEvent.ParticipantDisconnected, handleParticipantDisconnected);

    return () => {
      console.log('[Effect] Cleaning up event handlers');
      room.removeAllListeners();
    };
  }, [
    isConnected,
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

        console.log('[Room Connect] Connecting to room:', data.url);
        console.log('[Room Connect] Participant name:', data.participantName);

        await room.prepareConnection(data.url, data.token);
        await room.connect(data.url, data.token);

        console.log('[Room Connect] Connected successfully');
        console.log('[Room Connect] Local participant:', room.localParticipant?.identity);
        console.log('[Room Connect] Remote participants count:', room.remoteParticipants.size);

        setIsConnected(true);
        setError(null);
        setStageParticipantId(current => current ?? room.localParticipant?.sid ?? null);

        // Enable camera and microphone by default after connection
        setTimeout(() => {
          enableMediaByDefault();
        }, 500);

        room.remoteParticipants.forEach((participant) => {
          const displayName = getParticipantDisplayName(participant);
          console.log(`[Existing Participant] ${displayName} (${participant.sid})`);
          console.log(`  - Audio tracks: ${participant.audioTrackPublications.size}`);
          console.log(`  - Video tracks: ${participant.videoTrackPublications.size}`);

          setParticipants(prev => new Map(prev).set(participant.sid, participant));
          ensureParticipantTile(participant);

          // Process existing tracks
          participant.trackPublications.forEach((publication, trackSid) => {
            console.log(`  - Track ${trackSid}: ${publication.kind}, subscribed: ${publication.isSubscribed}, enabled: ${publication.isEnabled}`);

            if (publication.isSubscribed && publication.track) {
              console.log(`  - Track already subscribed, attaching...`);
              const track = publication.track as RemoteTrack;
              const element = track.attach();
              element.id = `${participant.sid}-${track.kind}`;

              if (track.kind === Track.Kind.Video) {
                participantVideoTracks.current.set(participant.sid, track);
                element.classList.add('meeting-video-element');
                attachTrackToTile(participant, element);
              } else if (track.kind === Track.Kind.Audio) {
                // Apply current volume to audio element
                if (element instanceof HTMLAudioElement) {
                  element.volume = volumeRef.current / 100;
                  console.log(`[Existing Audio] Set volume to ${volumeRef.current}%`);
                }
                element.style.display = 'none';
                attachTrackToTile(participant, element);
              }
            }
          });
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

  // Update volume for all audio elements when volume changes
  useEffect(() => {
    volumeRef.current = volume;
    const audioElements = document.querySelectorAll<HTMLAudioElement>('audio');
    console.log(`[Volume] Updating volume to ${volume}% for ${audioElements.length} audio elements`);
    audioElements.forEach(audio => {
      audio.volume = volume / 100;
    });
  }, [volume]);

  const confirmLeave = () => {
    room.disconnect();
    navigate('/dashboard/meetings');
  };

  const enableMediaByDefault = async () => {
    try {
      console.log('[Media] Enabling camera and microphone by default...');
      console.log('[Media] Navigator available:', typeof navigator);
      console.log('[Media] Navigator.mediaDevices available:', typeof navigator?.mediaDevices);
      console.log('[Media] Protocol:', window.location.protocol);
      console.log('[Media] User agent:', navigator.userAgent);

      // Check if mediaDevices is supported
      if (!navigator.mediaDevices || !navigator.mediaDevices.getUserMedia) {
        const protocol = window.location.protocol;
        const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';

        let errorMsg = 'Camera and microphone access is not available.\n\n';

        if (protocol === 'http:' && !isLocalhost) {
          errorMsg += '⚠️ The page is loaded via HTTP (not HTTPS).\n';
          errorMsg += 'WebRTC requires HTTPS for security.\n\n';
          errorMsg += 'Please access this page via HTTPS:\n';
          errorMsg += `https://${window.location.host}${window.location.pathname}`;
        } else {
          errorMsg += 'Your browser does not support WebRTC or camera/microphone access is blocked.\n\n';
          errorMsg += 'Please:\n';
          errorMsg += '1. Use a modern browser (Chrome, Firefox, Safari, Edge)\n';
          errorMsg += '2. Allow camera and microphone permissions when prompted\n';
          errorMsg += '3. Check that no other app is using the camera';
        }

        console.error('[Media] WebRTC not supported:', errorMsg);
        setError(errorMsg);
        return;
      }

      // For iOS Safari, we need to request permissions explicitly first
      const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
      console.log('[Media] iOS device:', isIOS);

      if (isIOS) {
        console.log('[Media] iOS detected, requesting permissions explicitly...');
        try {
          // Request permissions first on iOS
          const stream = await navigator.mediaDevices.getUserMedia({
            audio: true,
            video: { facingMode: 'user' }
          });
          console.log('[Media] iOS permissions granted, stopping test stream');
          stream.getTracks().forEach(track => track.stop());
        } catch (permErr) {
          console.error('[Media] Failed to get iOS permissions:', permErr);
          setError(`Please allow camera and microphone access: ${permErr instanceof Error ? permErr.message : 'Unknown error'}`);
          return;
        }
      }

      // Enable microphone first (more likely to succeed)
      try {
        await room.localParticipant.setMicrophoneEnabled(true);
        setIsMicEnabled(true);
        console.log('[Media] Microphone enabled successfully');
      } catch (micErr) {
        console.error('[Media] Failed to enable microphone:', micErr);
        const errorMsg = micErr instanceof Error ? micErr.message : 'Unknown error';
        console.error('[Media] Microphone error details:', errorMsg);
        // Don't set error state for mic, continue trying camera
      }

      // Then enable camera
      try {
        await room.localParticipant.setCameraEnabled(true);
        setIsCameraEnabled(true);
        renderLocalPreview();
        console.log('[Media] Camera enabled successfully');
      } catch (camErr) {
        console.error('[Media] Failed to enable camera:', camErr);
        const errorMsg = camErr instanceof Error ? camErr.message : 'Unknown error';
        console.error('[Media] Camera error details:', errorMsg);
        // Don't set error state, user can enable manually
      }

      if (stageParticipantId === room.localParticipant?.sid) {
        renderStageVideo(room.localParticipant.sid);
      }
    } catch (err) {
      console.error('[Media] Failed to enable media:', err);
      setError(err instanceof Error ? err.message : 'Failed to enable camera/microphone');
    }
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
  const speakingNames = activeSpeakers.map(s => getParticipantDisplayName(s)).join(', ');
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
              <div className="volume-control">
                <label htmlFor="volume-slider" style={{ fontSize: '14px', marginRight: '8px' }}>
                  🔊 {volume}%
                </label>
                <input
                  id="volume-slider"
                  type="range"
                  min="0"
                  max="100"
                  value={volume}
                  onChange={(e) => setVolume(Number(e.target.value))}
                  style={{ width: '100px' }}
                  title={t('meetingRoom.volumeControl') || 'Volume'}
                />
              </div>
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
