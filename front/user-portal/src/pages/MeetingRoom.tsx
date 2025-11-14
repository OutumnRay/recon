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
  TrackPublication,
  VideoPresets,
  Participant,
} from 'livekit-client';
import { useTranslation } from 'react-i18next';
import {
  getMeeting,
  getMeetingToken,
  startRecording,
  stopRecording,
  startTranscription,
  stopTranscription,
} from '../services/meetings';
import type { MeetingTokenResponse } from '../types/meeting';
import {
  LuMic,
  LuMicOff,
  LuVideo,
  LuVideoOff,
  LuMenu,
  LuLogOut,
  LuRefreshCw,
  LuCircle,
  LuFileText,
  LuUsers,
  LuSettings,
  LuMonitor,
  LuMonitorOff,
} from 'react-icons/lu';
import './MeetingRoom.css';
import MediaSettingsModal from '../components/MediaSettingsModal';
import { useWebSocket } from '../hooks/useWebSocket';

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
  const [showVolumeSlider, setShowVolumeSlider] = useState(false);
  const [micVolume, setMicVolume] = useState(100);
  const [tokenData, setTokenData] = useState<MeetingTokenResponse | null>(null);
  const [meetingTitle, setMeetingTitle] = useState('');
  const [stageParticipantId, setStageParticipantId] = useState<string | null>(null);
  const [isParticipantsCollapsed, setIsParticipantsCollapsed] = useState(false);
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);
  const [isMobile, setIsMobile] = useState(false);
  const [touchStart, setTouchStart] = useState<number | null>(null);
  const [touchEnd, setTouchEnd] = useState<number | null>(null);
  const [isRecording, setIsRecording] = useState(false);
  const [isTranscribing, setIsTranscribing] = useState(false);
  const [recordingError, setRecordingError] = useState<string | null>(null);
  const [showControls, setShowControls] = useState(true);
  const [showMediaSettings, setShowMediaSettings] = useState(false);
  const [selectedVideoDeviceId, setSelectedVideoDeviceId] = useState<string>('');
  const [selectedAudioDeviceId, setSelectedAudioDeviceId] = useState<string>('');
  const [screenShareQuality, setScreenShareQuality] = useState<'low' | 'medium' | 'high'>('medium');
  const [isScreenSharing, setIsScreenSharing] = useState(false);
  const [currentScreenSharer, setCurrentScreenSharer] = useState<string | null>(null);

  const videoContainerRef = useRef<HTMLDivElement>(null);
  const stageVideoRef = useRef<HTMLDivElement>(null);

  const participantVideoTracks = useRef<Map<string, RemoteTrack>>(new Map());
  const stageTrackRef = useRef<Track | null>(null);
  const stageElementRef = useRef<HTMLMediaElement | null>(null);
  const localPreviewTrackRef = useRef<Track | null>(null);
  const localPreviewElementRef = useRef<HTMLMediaElement | null>(null);
  const volumeRef = useRef<number>(100);

  // WebSocket connection for real-time communication
  const { sendMessage: sendWSMessage } = useWebSocket({
    meetingId: meetingId || '',
    enabled: !!meetingId && isConnected,
    onMessage: (message) => {
      console.log('[WebSocket] Received message:', message);

      switch (message.type) {
        case 'screen_share_started':
          if (message.data?.user_id && message.data.user_id !== room.localParticipant?.sid) {
            setCurrentScreenSharer(message.data.user_id);
            console.log(`[Screen Share] User ${message.data.user_id} started sharing`);
          }
          break;

        case 'screen_share_stopped':
          if (message.data?.user_id === currentScreenSharer) {
            setCurrentScreenSharer(null);
            console.log(`[Screen Share] User ${message.data.user_id} stopped sharing`);
          }
          break;
      }
    },
    onConnect: () => {
      console.log('[WebSocket] Connected to meeting room');
    },
    onDisconnect: () => {
      console.log('[WebSocket] Disconnected from meeting room');
    },
  });

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

  const ensureLocalParticipantTile = useCallback(() => {
    if (!room.localParticipant) return null;

    const sid = room.localParticipant.sid;
    let container = document.getElementById(`participant-${sid}`) as HTMLElement | null;

    if (!container) {
      container = document.createElement('div');
      container.id = `participant-${sid}`;
      container.className = 'remote-participant-tile local-participant-tile';
      container.dataset.participant = sid;
      container.addEventListener('click', () => {
        setStageParticipantId(sid);
      });

      const header = document.createElement('div');
      header.className = 'remote-participant-header';

      const avatar = document.createElement('div');
      avatar.className = 'participant-avatar';
      const displayName = tokenData?.participantName || 'You';
      avatar.textContent = getInitials(displayName);
      header.appendChild(avatar);

      const nameContainer = document.createElement('div');
      nameContainer.className = 'participant-name-container';

      const name = document.createElement('div');
      name.className = 'remote-participant-name';
      name.textContent = `${displayName} (You)`;
      nameContainer.appendChild(name);

      const micIndicator = document.createElement('div');
      micIndicator.className = 'mic-indicator muted';
      micIndicator.innerHTML = '<svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round" height="1em" width="1em"><line x1="1" y1="1" x2="23" y2="23"></line><path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"></path><path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"></path><line x1="12" y1="19" x2="12" y2="23"></line><line x1="8" y1="23" x2="16" y2="23"></line></svg>';
      micIndicator.style.display = 'none';
      nameContainer.appendChild(micIndicator);

      header.appendChild(nameContainer);
      container.appendChild(header);

      const videoSlot = document.createElement('div');
      videoSlot.className = 'remote-participant-video';
      videoSlot.dataset.slot = sid;
      container.appendChild(videoSlot);

      // Insert at the beginning of the list
      videoContainerRef.current?.insertBefore(container, videoContainerRef.current.firstChild);
    }
    return container;
  }, [room, tokenData]);

  const renderLocalPreview = useCallback(() => {
    const publication = room.localParticipant.getTrackPublication(Track.Source.Camera);

    const container = ensureLocalParticipantTile();
    if (!container) return;

    const videoSlot = container.querySelector<HTMLDivElement>('.remote-participant-video');
    if (!videoSlot) return;

    // Clear previous video
    videoSlot.innerHTML = '';

    // Only show video if track exists
    if (publication?.track) {
      const element = publication.track.attach();
      element.classList.add('meeting-video-element');
      videoSlot.appendChild(element);
      localPreviewTrackRef.current = publication.track;
      localPreviewElementRef.current = element;
      showParticipantVideo(room.localParticipant.sid);
    } else {
      hideParticipantVideo(room.localParticipant.sid);
    }
  }, [ensureLocalParticipantTile, room]);

  const renderStageVideo = useCallback((preferredId?: string) => {
    const container = stageVideoRef.current;
    if (!container) return;

    detachStageVideo();
    container.innerHTML = '';

    const targetId = preferredId || stageParticipantId || room.localParticipant?.sid || null;

    let track: Track | null = null;
    let element: HTMLMediaElement | null = null;

    if (targetId && targetId === room.localParticipant?.sid) {
      // For local participant, prioritize screen share over camera
      const screenPublication = room.localParticipant.getTrackPublication(Track.Source.ScreenShare);
      const cameraPublication = room.localParticipant.getTrackPublication(Track.Source.Camera);

      if (screenPublication?.track) {
        track = screenPublication.track;
        element = screenPublication.track.attach();
        console.log('[Stage] Rendering local screen share on stage');
      } else if (cameraPublication?.track) {
        track = cameraPublication.track;
        element = cameraPublication.track.attach();
        console.log('[Stage] Rendering local camera on stage');
      }
    } else if (targetId) {
      // For remote participant, check for screen share first
      const participant = room.remoteParticipants.get(targetId);
      if (participant) {
        const screenPub = Array.from(participant.videoTrackPublications.values())
          .find(pub => pub.source === Track.Source.ScreenShare);
        const cameraPub = Array.from(participant.videoTrackPublications.values())
          .find(pub => pub.source === Track.Source.Camera);

        if (screenPub?.track) {
          track = screenPub.track as RemoteTrack;
          element = track.attach();
          console.log('[Stage] Rendering remote screen share on stage');
        } else if (cameraPub?.track) {
          track = cameraPub.track as RemoteTrack;
          element = track.attach();
          console.log('[Stage] Rendering remote camera on stage');
        } else {
          // Fallback to old method
          const remoteTrack = participantVideoTracks.current.get(targetId);
          if (remoteTrack) {
            track = remoteTrack;
            element = remoteTrack.attach();
            console.log('[Stage] Rendering remote video (fallback) on stage');
          }
        }
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

      const nameContainer = document.createElement('div');
      nameContainer.className = 'participant-name-container';

      const name = document.createElement('div');
      name.className = 'remote-participant-name';
      name.textContent = displayName;
      nameContainer.appendChild(name);

      const micIndicator = document.createElement('div');
      micIndicator.className = 'mic-indicator muted';
      micIndicator.innerHTML = '<svg stroke="currentColor" fill="none" stroke-width="2" viewBox="0 0 24 24" stroke-linecap="round" stroke-linejoin="round" height="1em" width="1em"><line x1="1" y1="1" x2="23" y2="23"></line><path d="M9 9v3a3 3 0 0 0 5.12 2.12M15 9.34V4a3 3 0 0 0-5.94-.6"></path><path d="M17 16.95A7 7 0 0 1 5 12v-2m14 0v2a7 7 0 0 1-.11 1.23"></path><line x1="12" y1="19" x2="12" y2="23"></line><line x1="8" y1="23" x2="16" y2="23"></line></svg>';
      micIndicator.style.display = 'none';
      nameContainer.appendChild(micIndicator);

      header.appendChild(nameContainer);
      container.appendChild(header);

      const videoSlot = document.createElement('div');
      videoSlot.className = 'remote-participant-video';
      videoSlot.dataset.slot = participant.sid;
      container.appendChild(videoSlot);

      videoContainerRef.current?.appendChild(container);
    }
    return container;
  };

  const updateMicIndicator = useCallback((participantSid: string, isMuted: boolean) => {
    const container = document.getElementById(`participant-${participantSid}`);
    if (!container) return;

    const micIndicator = container.querySelector<HTMLElement>('.mic-indicator');
    if (micIndicator) {
      if (isMuted) {
        micIndicator.style.display = 'flex';
        micIndicator.classList.add('muted');
      } else {
        micIndicator.style.display = 'none';
        micIndicator.classList.remove('muted');
      }
    }
  }, []);

  const hideParticipantVideo = (participantSid: string) => {
    const container = document.getElementById(`participant-${participantSid}`);
    if (container) {
      const videoSlot = container.querySelector<HTMLDivElement>('.remote-participant-video');
      if (videoSlot) {
        videoSlot.style.display = 'none';
        videoSlot.innerHTML = '';
        console.log(`[Hide Video] Hidden video for participant ${participantSid}`);
      }
    }
  };

  const showParticipantVideo = (participantSid: string) => {
    const container = document.getElementById(`participant-${participantSid}`);
    if (container) {
      const videoSlot = container.querySelector<HTMLDivElement>('.remote-participant-video');
      if (videoSlot) {
        videoSlot.style.display = '';
        console.log(`[Show Video] Shown video for participant ${participantSid}`);
      }
    }
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

  // Check if mobile device
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth <= 768);
    };

    checkMobile();
    window.addEventListener('resize', checkMobile);

    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Initialize participants sidebar collapsed on mobile
  useEffect(() => {
    if (isMobile) {
      setIsParticipantsCollapsed(true);
    }
  }, [isMobile]);

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

        // Show indicators based only on runtime state (is_recording / is_transcribing)
        // The needs_* fields are only used for initial auto-start
        setIsRecording(meeting.is_recording || false);
        setIsTranscribing(meeting.is_transcribing || false);

        console.log('[Meeting Settings] needs_video_record:', meeting.needs_video_record);
        console.log('[Meeting Settings] needs_audio_record:', meeting.needs_audio_record);
        console.log('[Meeting Settings] needs_transcription:', meeting.needs_transcription);
        console.log('[Meeting Settings] is_recording:', meeting.is_recording);
        console.log('[Meeting Settings] is_transcribing:', meeting.is_transcribing);
      } catch (err) {
        console.error('Failed to load meeting details:', err);
      }
    };

    loadMeetingTitle();

    // Periodically refresh recording/transcription status every 10 seconds
    // Only update state if values actually changed to prevent flickering
    const statusInterval = setInterval(() => {
      if (meetingId) {
        getMeeting(meetingId)
          .then(meeting => {
            const newIsRecording = meeting.is_recording || false;
            const newIsTranscribing = meeting.is_transcribing || false;

            // Only update if values changed
            setIsRecording(prev => prev !== newIsRecording ? newIsRecording : prev);
            setIsTranscribing(prev => prev !== newIsTranscribing ? newIsTranscribing : prev);
          })
          .catch(err => {
            console.error('Failed to refresh meeting status:', err);
          });
      }
    }, 10000); // Update every 10 seconds

    return () => clearInterval(statusInterval);
  }, [meetingId, t]);

  useEffect(() => {
    const pageTitle = meetingTitle || tokenData?.roomName || t('meetingRoom.pageTitle');
    document.title = `Recontext - ${pageTitle}`;
  }, [meetingTitle, tokenData, t]);

  // Auto-hide controls disabled - causes flickering
  // Controls are always visible now
  useEffect(() => {
    setShowControls(true);
  }, []);

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

      if (track.kind === Track.Kind.Video) {
        participantVideoTracks.current.delete(participant.sid);
        // Hide video when video is unsubscribed
        hideParticipantVideo(participant.sid);
      }

      if (stageParticipantId === participant.sid) {
        renderStageVideo();
      }
    };

    const handleTrackPublished = async (
      publication: RemoteTrackPublication,
      participant: RemoteParticipant,
    ) => {
      const displayName = getParticipantDisplayName(participant);
      console.log(`[Track Published] Participant: ${displayName}, Track: ${publication.kind}, Source: ${publication.source}`);
      console.log(`  - Is subscribed: ${publication.isSubscribed}`);
      console.log(`  - Is enabled: ${publication.isEnabled}`);
      console.log(`  - Track exists: ${publication.track ? 'yes' : 'no'}`);

      // If track is not subscribed yet, explicitly subscribe to it
      if (!publication.isSubscribed && publication.kind !== Track.Kind.Unknown) {
        console.log(`  - Track not subscribed, attempting to subscribe...`);
        try {
          await publication.setSubscribed(true);
          console.log(`  - Successfully subscribed to ${publication.kind} track`);
        } catch (err) {
          console.error(`  - Failed to subscribe to ${publication.kind} track:`, err);
        }
      }

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

      // Hide video when video is unpublished
      if (publication.kind === Track.Kind.Video) {
        hideParticipantVideo(participant.sid);
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
        // Hide video when local video is unpublished
        hideParticipantVideo(room.localParticipant.sid);
        if (stageParticipantId === room.localParticipant?.sid) {
          renderStageVideo();
        }
      }
    };

    const handleActiveSpeakerChange = (speakers: Participant[]) => {
      setActiveSpeakers(speakers);
    };

    const handleTrackMuted = (publication: TrackPublication, participant: Participant) => {
      if (publication.kind === Track.Kind.Audio) {
        console.log(`[Track Muted] ${getParticipantDisplayName(participant)} muted their microphone`);
        updateMicIndicator(participant.sid, true);
      } else if (publication.kind === Track.Kind.Video) {
        console.log(`[Track Muted] ${getParticipantDisplayName(participant)} muted their camera`);
        // Hide video when video is muted
        hideParticipantVideo(participant.sid);
      }
    };

    const handleTrackUnmuted = (publication: TrackPublication, participant: Participant) => {
      if (publication.kind === Track.Kind.Audio) {
        console.log(`[Track Unmuted] ${getParticipantDisplayName(participant)} unmuted their microphone`);
        updateMicIndicator(participant.sid, false);
      } else if (publication.kind === Track.Kind.Video) {
        console.log(`[Track Unmuted] ${getParticipantDisplayName(participant)} unmuted their camera`);
        // Show video and re-attach video when unmuted
        showParticipantVideo(participant.sid);
        if (participant instanceof RemoteParticipant) {
          const videoPublication = participant.getTrackPublication(Track.Source.Camera);
          if (videoPublication?.track) {
            const element = videoPublication.track.attach();
            element.classList.add('meeting-video-element');
            attachTrackToTile(participant, element);
          }
        } else if (participant === room.localParticipant) {
          // For local participant, re-render which will show video
          renderLocalPreview();
        }
      }
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

      // Always create tile for participant
      ensureParticipantTile(participant);

      // Hide video initially if no video track is published
      const hasVideoTrack = Array.from(participant.videoTrackPublications.values()).some(
        pub => pub.isSubscribed || pub.track
      );
      if (!hasVideoTrack) {
        hideParticipantVideo(participant.sid);
      }

      // Check if participant already has published tracks that we need to subscribe to
      participant.trackPublications.forEach(async (publication, trackSid) => {
        console.log(`[Existing Track] Track ${trackSid}: ${publication.kind}, subscribed: ${publication.isSubscribed}, track: ${publication.track ? 'exists' : 'null'}`);

        // If track exists but is not subscribed, explicitly subscribe to it
        if (!publication.isSubscribed && publication.kind !== Track.Kind.Unknown) {
          console.log(`[Manual Subscribe] Attempting to subscribe to ${publication.kind} track`);
          try {
            await publication.setSubscribed(true);
            console.log(`[Manual Subscribe] Successfully subscribed to ${publication.kind} track`);
          } catch (err) {
            console.error(`[Manual Subscribe] Failed to subscribe to ${publication.kind} track:`, err);
          }
        }

        // If already subscribed and has track, attach it immediately
        if (publication.isSubscribed && publication.track) {
          console.log(`[Existing Track] Track already subscribed, manually attaching...`);
          const track = publication.track as RemoteTrack;
          const element = track.attach();
          element.id = `${participant.sid}-${track.kind}`;

          if (track.kind === Track.Kind.Video) {
            participantVideoTracks.current.set(participant.sid, track);
            element.classList.add('meeting-video-element');
            attachTrackToTile(participant, element);
            if (stageParticipantId === participant.sid) {
              renderStageVideo(participant.sid);
            }
          } else if (track.kind === Track.Kind.Audio) {
            if (element instanceof HTMLAudioElement) {
              element.volume = volumeRef.current / 100;
            }
            element.style.display = 'none';
            attachTrackToTile(participant, element);
          }
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
      .on(RoomEvent.TrackMuted, handleTrackMuted)
      .on(RoomEvent.TrackUnmuted, handleTrackUnmuted)
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
    updateMicIndicator,
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

        room.remoteParticipants.forEach(async (participant) => {
          const displayName = getParticipantDisplayName(participant);
          console.log(`[Existing Participant] ${displayName} (${participant.sid})`);
          console.log(`  - Audio tracks: ${participant.audioTrackPublications.size}`);
          console.log(`  - Video tracks: ${participant.videoTrackPublications.size}`);

          setParticipants(prev => new Map(prev).set(participant.sid, participant));
          ensureParticipantTile(participant);

          // Process existing tracks
          participant.trackPublications.forEach(async (publication, trackSid) => {
            console.log(`  - Track ${trackSid}: ${publication.kind}, subscribed: ${publication.isSubscribed}, enabled: ${publication.isEnabled}`);

            // If track is not subscribed yet, subscribe to it
            if (!publication.isSubscribed && publication.kind !== Track.Kind.Unknown) {
              console.log(`  - Track not subscribed yet, subscribing to ${publication.kind} track...`);
              try {
                await publication.setSubscribed(true);
                console.log(`  - Successfully subscribed to ${publication.kind} track`);
              } catch (err) {
                console.error(`  - Failed to subscribe to ${publication.kind} track:`, err);
              }
            }

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


  const confirmLeave = useCallback(() => {
    console.log('[Leave] Disconnecting from room and navigating to dashboard...');
    setShowLeaveConfirm(false);

    // Disconnect from room
    try {
      room.disconnect();
    } catch (err) {
      console.error('[Leave] Error disconnecting:', err);
    }

    // Use window.location for full page reload to dashboard
    window.location.href = '/dashboard';
  }, [room]);

  // Aggressive track subscription - force subscribe to all tracks
  const forceSubscribeToAllTracks = useCallback(async () => {
    if (!isConnected) return;

    console.log('[Force Subscribe] Starting aggressive track subscription...');

    room.remoteParticipants.forEach(async (participant) => {
      const displayName = getParticipantDisplayName(participant);

      participant.trackPublications.forEach(async (publication) => {
        // Skip if already subscribed or not a valid track
        if (publication.isSubscribed || publication.kind === Track.Kind.Unknown) {
          return;
        }

        console.log(`[Force Subscribe] Subscribing ${displayName} to ${publication.kind} track`);

        try {
          await publication.setSubscribed(true);
          console.log(`[Force Subscribe] ✓ Successfully subscribed to ${publication.kind} track`);

          // Wait a bit for track to be ready
          setTimeout(() => {
            if (publication.track) {
              const track = publication.track as RemoteTrack;
              const element = track.attach();
              element.id = `${participant.sid}-${track.kind}`;

              if (track.kind === Track.Kind.Video) {
                console.log(`[Force Subscribe] Attaching video for ${displayName}`);
                participantVideoTracks.current.set(participant.sid, track);
                element.classList.add('meeting-video-element');
                attachTrackToTile(participant, element);

                if (stageParticipantId === participant.sid) {
                  renderStageVideo(participant.sid);
                }
              } else if (track.kind === Track.Kind.Audio) {
                console.log(`[Force Subscribe] Attaching audio for ${displayName}`);
                if (element instanceof HTMLAudioElement) {
                  element.volume = volumeRef.current / 100;
                }
                element.style.display = 'none';
                attachTrackToTile(participant, element);
              }
            }
          }, 500);
        } catch (err) {
          console.error(`[Force Subscribe] ✗ Failed to subscribe to ${publication.kind} track:`, err);
        }
      });
    });
  }, [isConnected, room, stageParticipantId, renderStageVideo]);

  // Force subscription timers removed - causes flickering
  // Subscription now happens only via event handlers

  // Touch handlers for swipe gesture on participant sidebar
  const minSwipeDistance = 50;

  const onTouchStart = (e: React.TouchEvent) => {
    setTouchEnd(null);
    setTouchStart(e.targetTouches[0].clientY);
  };

  const onTouchMove = (e: React.TouchEvent) => {
    setTouchEnd(e.targetTouches[0].clientY);
  };

  const onTouchEnd = () => {
    if (!touchStart || !touchEnd) return;

    const distance = touchStart - touchEnd;
    const isUpSwipe = distance > minSwipeDistance;
    const isDownSwipe = distance < -minSwipeDistance;

    if (isUpSwipe && isParticipantsCollapsed) {
      setIsParticipantsCollapsed(false);
    } else if (isDownSwipe && !isParticipantsCollapsed) {
      setIsParticipantsCollapsed(true);
    }
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
        // Hide video when camera is disabled
        hideParticipantVideo(room.localParticipant.sid);
      } else {
        await room.localParticipant.setCameraEnabled(true);
        setIsCameraEnabled(true);
        renderLocalPreview();
      }

      if (stageParticipantId === room.localParticipant?.sid) {
        renderStageVideo(room.localParticipant.sid);
      }
    } catch (err) {
      console.warn('Failed to toggle camera:', err);
      // Handle specific error types
      if (err instanceof Error) {
        if (err.name === 'NotFoundError') {
          setError(t('meetingRoom.errors.cameraNotFound') || 'Camera not found. Please check your device.');
        } else if (err.name === 'NotReadableError') {
          setError(t('meetingRoom.errors.cameraInUse') || 'Camera is being used by another application.');
        } else if (err.name === 'NotAllowedError') {
          setError(t('meetingRoom.errors.cameraPermissionDenied') || 'Camera permission denied.');
        } else {
          setError(err.message || t('meetingRoom.errors.toggleCamera'));
        }
      } else {
        setError(t('meetingRoom.errors.toggleCamera'));
      }
      // Ensure camera state reflects actual state
      setIsCameraEnabled(false);
      // Hide video when camera fails
      hideParticipantVideo(room.localParticipant.sid);
    }
  };

  const toggleMicrophone = async () => {
    try {
      if (isMicEnabled) {
        await room.localParticipant.setMicrophoneEnabled(false);
        setIsMicEnabled(false);
        updateMicIndicator(room.localParticipant.sid, true);
      } else {
        await room.localParticipant.setMicrophoneEnabled(true);
        setIsMicEnabled(true);
        updateMicIndicator(room.localParticipant.sid, false);
      }
    } catch (err) {
      console.warn('Failed to toggle microphone:', err);
      // Handle specific error types
      if (err instanceof Error) {
        if (err.name === 'NotFoundError') {
          setError(t('meetingRoom.errors.micNotFound') || 'Microphone not found. Please check your device.');
        } else if (err.name === 'NotReadableError') {
          setError(t('meetingRoom.errors.micInUse') || 'Microphone is being used by another application.');
        } else if (err.name === 'NotAllowedError') {
          setError(t('meetingRoom.errors.micPermissionDenied') || 'Microphone permission denied.');
        } else {
          setError(err.message || t('meetingRoom.errors.toggleMicrophone'));
        }
      } else {
        setError(t('meetingRoom.errors.toggleMicrophone'));
      }
      // Ensure microphone state reflects actual state
      setIsMicEnabled(false);
    }
  };

  const handleVolumeChange = useCallback((newVolume: number) => {
    setMicVolume(newVolume);
    // Note: Volume control affects the local participant's microphone gain
    // This is a UI-only control for now - actual audio processing would need
    // to be handled at the track level with Web Audio API if needed
  }, []);

  // Reserved for future use - manual start recording
  // @ts-ignore - unused variable kept for future use
  const handleStartRecording = async () => {
    if (!meetingId) return;
    try {
      setRecordingError(null);
      await startRecording(meetingId);
      setIsRecording(true);
    } catch (err) {
      console.error('Failed to start recording:', err);
      setRecordingError(err instanceof Error ? err.message : 'Failed to start recording');
    }
  };

  const handleStopRecording = async () => {
    if (!meetingId) return;
    try {
      setRecordingError(null);
      await stopRecording(meetingId);
      setIsRecording(false);
    } catch (err) {
      console.error('Failed to stop recording:', err);
      setRecordingError(err instanceof Error ? err.message : 'Failed to stop recording');
    }
  };

  // Reserved for future use - manual start transcription
  // @ts-ignore - unused variable kept for future use
  const handleStartTranscription = async () => {
    if (!meetingId) return;
    try {
      setRecordingError(null);
      await startTranscription(meetingId);
      setIsTranscribing(true);
    } catch (err) {
      console.error('Failed to start transcription:', err);
      setRecordingError(err instanceof Error ? err.message : 'Failed to start transcription');
    }
  };

  const handleStopTranscription = async () => {
    if (!meetingId) return;
    try {
      setRecordingError(null);
      await stopTranscription(meetingId);
      setIsTranscribing(false);
    } catch (err) {
      console.error('Failed to stop transcription:', err);
      setRecordingError(err instanceof Error ? err.message : 'Failed to stop transcription');
    }
  };

  const handleApplyMediaSettings = async (videoDeviceId: string, audioDeviceId: string, quality: 'low' | 'medium' | 'high') => {
    try {
      console.log('[Media Settings] Applying settings:', { videoDeviceId, audioDeviceId, screenShareQuality: quality });

      // Update selected devices and screen share quality
      setSelectedVideoDeviceId(videoDeviceId);
      setSelectedAudioDeviceId(audioDeviceId);
      setScreenShareQuality(quality);

      // If camera is enabled, switch to new video device
      if (isCameraEnabled && videoDeviceId) {
        await room.localParticipant.setCameraEnabled(false);
        await room.switchActiveDevice('videoinput', videoDeviceId);
        await room.localParticipant.setCameraEnabled(true);
        renderLocalPreview();
        if (stageParticipantId === room.localParticipant?.sid) {
          renderStageVideo(room.localParticipant.sid);
        }
      }

      // If microphone is enabled, switch to new audio device
      if (isMicEnabled && audioDeviceId) {
        await room.localParticipant.setMicrophoneEnabled(false);
        await room.switchActiveDevice('audioinput', audioDeviceId);
        await room.localParticipant.setMicrophoneEnabled(true);
      }

      console.log('[Media Settings] Settings applied successfully');
    } catch (err) {
      console.error('[Media Settings] Failed to apply settings:', err);
      setError(err instanceof Error ? err.message : 'Failed to apply media settings');
    }
  };

  const toggleScreenShare = async () => {
    try {
      if (isScreenSharing) {
        // Stop screen sharing
        console.log('[Screen Share] Stopping screen share...');
        await room.localParticipant.setScreenShareEnabled(false);
        setIsScreenSharing(false);
        setCurrentScreenSharer(null);

        // Notify other participants via WebSocket
        sendWSMessage('screen_share_stop');
        console.log('[Screen Share] Screen share stopped');

        // Update stage to show camera instead
        if (stageParticipantId === room.localParticipant?.sid) {
          renderStageVideo(room.localParticipant.sid);
        }
      } else {
        // Check if someone else is already sharing
        if (currentScreenSharer && currentScreenSharer !== room.localParticipant?.sid) {
          setError(t('meetingRoom.errors.screenShareInUse'));
          return;
        }

        // Get resolution based on selected quality
        let resolution;
        switch (screenShareQuality) {
          case 'low':
            resolution = VideoPresets.h720.resolution;
            console.log('[Screen Share] Starting screen share with 720p quality...');
            break;
          case 'medium':
            resolution = VideoPresets.h1080.resolution;
            console.log('[Screen Share] Starting screen share with 1080p quality...');
            break;
          case 'high':
            resolution = VideoPresets.h1440.resolution;
            console.log('[Screen Share] Starting screen share with 2K quality...');
            break;
          default:
            resolution = VideoPresets.h1080.resolution;
            console.log('[Screen Share] Starting screen share with default 1080p quality...');
        }

        await room.localParticipant.setScreenShareEnabled(true, {
          resolution,
        });
        setIsScreenSharing(true);
        setCurrentScreenSharer(room.localParticipant?.sid || null);

        // Notify other participants via WebSocket
        sendWSMessage('screen_share_start');
        console.log('[Screen Share] Screen share started');

        // Switch stage to show screen share
        if (room.localParticipant?.sid) {
          setStageParticipantId(room.localParticipant.sid);
          setTimeout(() => {
            renderStageVideo(room.localParticipant.sid);
          }, 100);
        }
      }
    } catch (err) {
      console.error('[Screen Share] Failed to toggle screen share:', err);
      if (err instanceof Error) {
        if (err.name === 'NotAllowedError') {
          setError(t('meetingRoom.errors.screenSharePermissionDenied'));
        } else {
          setError(err.message || t('meetingRoom.errors.screenShareFailed'));
        }
      } else {
        setError(t('meetingRoom.errors.screenShareFailed'));
      }
      setIsScreenSharing(false);
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
    <div className="meeting-room-page" style={{ cursor: showControls ? 'default' : 'none' }}>
      <div
        className={`meeting-room-header ${showControls ? '' : 'hidden'}`}
        aria-hidden={!showControls}
      >
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
              onClick={() => setIsParticipantsCollapsed(prev => !prev)}
              className="icon-circle-button"
              aria-label={isParticipantsCollapsed ? t('meetingRoom.controls.showParticipants') : t('meetingRoom.controls.hideParticipants')}
              title={isParticipantsCollapsed ? t('meetingRoom.controls.showParticipants') : t('meetingRoom.controls.hideParticipants')}
            >
              <LuUsers />
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

      {!showControls && (isRecording || isTranscribing) && (
        <div className="recording-status-floating">
          {isRecording && (
            <div className="status-pill recording">
              <LuCircle className="status-icon" />
              <span>{t('meetingRoom.indicators.recording')}</span>
            </div>
          )}
          {isTranscribing && (
            <div className="status-pill transcription">
              <LuFileText className="status-icon" />
              <span>{t('meetingRoom.indicators.transcribing')}</span>
            </div>
          )}
        </div>
      )}

      {error && isConnected && (
        <div className="alert alert-error meeting-room-inline-alert">
          {error}
        </div>
      )}

      {recordingError && (
        <div className="alert alert-error meeting-room-inline-alert">
          {recordingError}
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
            {showControls && <div className="stage-controls" style={{
              transition: 'opacity 0.3s ease',
            }}>
              <button
                onClick={toggleCamera}
                className="icon-circle-button"
                aria-label={isCameraEnabled ? t('meetingRoom.disableCamera') : t('meetingRoom.enableCamera')}
                title={isCameraEnabled ? t('meetingRoom.disableCamera') : t('meetingRoom.enableCamera')}
                disabled={!isConnected}
              >
                {isCameraEnabled ? <LuVideo /> : <LuVideoOff />}
              </button>
              <div
                className="mic-button-container"
                onMouseEnter={() => setShowVolumeSlider(true)}
                onMouseLeave={() => setShowVolumeSlider(false)}
              >
                <button
                  onClick={toggleMicrophone}
                  className="icon-circle-button"
                  aria-label={isMicEnabled ? t('meetingRoom.disableMic') : t('meetingRoom.enableMic')}
                  title={isMicEnabled ? t('meetingRoom.disableMic') : t('meetingRoom.enableMic')}
                  disabled={!isConnected}
                >
                  {isMicEnabled ? <LuMic /> : <LuMicOff />}
                </button>
                {showVolumeSlider && isMicEnabled && (
                  <div className="volume-slider-vertical">
                    <input
                      type="range"
                      min="0"
                      max="100"
                      value={micVolume}
                      onChange={(e) => handleVolumeChange(Number(e.target.value))}
                      className="vertical-slider"
                    />
                    <span className="volume-label">{micVolume}%</span>
                  </div>
                )}
              </div>
              <button
                onClick={toggleScreenShare}
                className={`icon-circle-button ${isScreenSharing ? 'active-share' : ''}`}
                aria-label={isScreenSharing ? t('meetingRoom.controls.screenShareStop') : t('meetingRoom.controls.screenShareStart')}
                title={isScreenSharing ? t('meetingRoom.controls.screenShareStop') : t('meetingRoom.controls.screenShareStart')}
                disabled={!isConnected || (currentScreenSharer !== null && currentScreenSharer !== room.localParticipant?.sid)}
              >
                {isScreenSharing ? <LuMonitorOff /> : <LuMonitor />}
              </button>
              <button
                onClick={() => setShowMediaSettings(true)}
                className="icon-circle-button"
                aria-label={t('meetingRoom.controls.mediaSettings')}
                title={t('meetingRoom.controls.mediaSettings')}
                disabled={!isConnected}
              >
                <LuSettings />
              </button>

              {/* Recording button */}
              <button
                onClick={isRecording ? handleStopRecording : handleStartRecording}
                className={`icon-circle-button ${isRecording ? 'recording' : ''}`}
                aria-label={isRecording ? t('meetingRoom.controls.stopRecording') : t('meetingRoom.controls.startRecording')}
                title={isRecording ? t('meetingRoom.controls.stopRecording') : t('meetingRoom.controls.startRecording')}
                disabled={!isConnected}
                style={isRecording ? { color: '#ef4444' } : {}}
              >
                <LuCircle />
              </button>

              {/* Transcription button */}
              <button
                onClick={isTranscribing ? handleStopTranscription : handleStartTranscription}
                className={`icon-circle-button ${isTranscribing ? 'transcribing' : ''}`}
                aria-label={isTranscribing ? t('meetingRoom.controls.stopTranscription') : t('meetingRoom.controls.startTranscription')}
                title={isTranscribing ? t('meetingRoom.controls.stopTranscription') : t('meetingRoom.controls.startTranscription')}
                disabled={!isConnected}
                style={isTranscribing ? { color: '#3b82f6' } : {}}
              >
                <LuFileText />
              </button>

              {isMobile && (
                <button
                  onClick={forceSubscribeToAllTracks}
                  className="icon-circle-button refresh-button"
                  aria-label="Refresh participants"
                  title="Refresh participants video"
                  disabled={!isConnected}
                >
                  <LuRefreshCw />
                </button>
              )}
            </div>}
          </div>
        </div>

        {showControls && <aside
          className={`participant-sidebar ${isParticipantsCollapsed ? 'collapsed' : ''}`}
          style={{
            transition: 'transform 0.3s ease, opacity 0.3s ease',
          }}
          onTouchStart={isMobile ? onTouchStart : undefined}
          onTouchMove={isMobile ? onTouchMove : undefined}
          onTouchEnd={isMobile ? onTouchEnd : undefined}
        >
          <div
            className="participant-sidebar-header"
            onClick={isMobile ? () => setIsParticipantsCollapsed(prev => !prev) : undefined}
          >
            {!isParticipantsCollapsed && (
              <div className="participant-sidebar-title">
                <span>{t('meetingRoom.remoteParticipants')}</span>
                <span className="participant-count-pill">
                  {participants.size + 1}
                </span>
              </div>
            )}
            {!isMobile && (
              <button
                className="icon-circle-button"
                onClick={() => setIsParticipantsCollapsed(prev => !prev)}
                aria-label={t('meetingRoom.toggleParticipants')}
                title={t('meetingRoom.toggleParticipants')}
              >
                <LuMenu />
              </button>
            )}
          </div>
          <div className="participant-sidebar-list" ref={videoContainerRef} />
          {participants.size === 0 && (
            <p className="empty-participants-text">{t('meetingRoom.waitingForParticipants')}</p>
          )}
        </aside>}
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

      <MediaSettingsModal
        isOpen={showMediaSettings}
        onClose={() => setShowMediaSettings(false)}
        onApplySettings={handleApplyMediaSettings}
        currentVideoDeviceId={selectedVideoDeviceId}
        currentAudioDeviceId={selectedAudioDeviceId}
        currentScreenShareQuality={screenShareQuality}
      />
    </div>
  );
}
