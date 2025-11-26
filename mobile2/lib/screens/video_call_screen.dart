import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
import 'package:livekit_client/livekit_client.dart' as livekit;
import '../utils/logger.dart';
import '../models/meeting.dart';
import 'package:wakelock_plus/wakelock_plus.dart';

class VideoCallScreen extends StatefulWidget {
  final String token;
  final String url;
  final String meetingTitle;
  final List<MeetingParticipant> participants;

  const VideoCallScreen({
    super.key,
    required this.token,
    required this.url,
    required this.meetingTitle,
    this.participants = const [],
  });

  @override
  State<VideoCallScreen> createState() => _VideoCallScreenState();
}

class _VideoCallScreenState extends State<VideoCallScreen> {
  livekit.Room? _room;
  List<ParticipantTrack> _participantTracks = [];
  bool _isMicEnabled = true;
  bool _isCameraEnabled = true;
  bool _isConnecting = true;
  String? _error;
  bool _wasDisconnected = false;
  livekit.CameraPosition _cameraPosition = livekit.CameraPosition.front;

  // Stage participant (main view)
  String? _stageParticipantId;
  bool _showParticipantsList = false;
  bool _manuallySelectedParticipant = false; // Track if user manually selected a participant

  // Active speakers tracking
  List<livekit.Participant> _activeSpeakers = [];

  // Маппинг userId -> displayName
  Map<String, String> get _participantNames {
    final map = <String, String>{};
    for (final p in widget.participants) {
      map[p.userId] = p.displayName;
    }
    return map;
  }

  String _getDisplayName(String identity) {
    // identity в LiveKit это userId
    return _participantNames[identity] ?? identity;
  }

  @override
  void initState() {
    super.initState();
    _connectToRoom();
  }

  Future<void> _connectToRoom() async {
    try {
      setState(() {
        _isConnecting = true;
        _error = null;
      });

      Logger.logInfo('Connecting to LiveKit room', data: {
        'url': widget.url,
        'token': '${widget.token.substring(0, 20)}...',
      });

      // Create room with options
      final roomOptions = livekit.RoomOptions(
        adaptiveStream: true,
        dynacast: true,
        defaultCameraCaptureOptions: const livekit.CameraCaptureOptions(
          maxFrameRate: 30,
          params: livekit.VideoParametersPresets.h720_169,
        ),
        defaultAudioPublishOptions: const livekit.AudioPublishOptions(
          name: 'microphone',
        ),
        defaultVideoPublishOptions: const livekit.VideoPublishOptions(
          name: 'camera',
        ),
      );

      final room = livekit.Room(roomOptions: roomOptions);

      // Set up event listeners
      room.addListener(_onRoomUpdate);

      // Listen for disconnection events
      room.createListener().on<livekit.RoomDisconnectedEvent>((event) {
        _onRoomDisconnected();
      });

      // Listen for active speakers
      room.createListener().on<livekit.ActiveSpeakersChangedEvent>((event) {
        _onActiveSpeakersChanged(event.speakers);
      });

      // Prepare connection for faster connection
      await room.prepareConnection(widget.url, widget.token);

      // Connect to room
      await room.connect(widget.url, widget.token);

      Logger.logSuccess('Connected to LiveKit room');

      // Enable wakelock to prevent screen from turning off
      try {
        await WakelockPlus.enable();
        Logger.logInfo('Wakelock enabled - screen will stay on during meeting');
      } catch (e) {
        Logger.logWarning('Failed to enable wakelock: $e');
      }

      // Wait for connection state to be fully connected
      await Future.delayed(const Duration(milliseconds: 500));

      // Check if connection state is connected before publishing
      if (room.connectionState != livekit.ConnectionState.connected) {
        Logger.logWarning('Room not in connected state, waiting...');
        // Wait up to 5 seconds for connection
        for (int i = 0; i < 10; i++) {
          await Future.delayed(const Duration(milliseconds: 500));
          if (room.connectionState == livekit.ConnectionState.connected) {
            break;
          }
        }
      }

      if (mounted) {
        setState(() {
          _room = room;
          _isConnecting = false;
          _updateParticipantTracks();
        });
      }

      // Only enable camera and microphone if we're connected
      if (room.connectionState == livekit.ConnectionState.connected) {
        // Enable camera and microphone with error handling
        try {
          // Video will fail when running in iOS simulator
          await room.localParticipant?.setCameraEnabled(true);
          if (mounted) {
            setState(() {
              _isCameraEnabled = true;
            });
          }
        } catch (error) {
          Logger.logWarning('Could not publish video: $error');
          if (mounted) {
            setState(() {
              _isCameraEnabled = false;
            });
          }
        }

        try {
          await room.localParticipant?.setMicrophoneEnabled(true);
          if (mounted) {
            setState(() {
              _isMicEnabled = true;
            });
          }
        } catch (error) {
          Logger.logWarning('Could not enable microphone: $error');
          if (mounted) {
            setState(() {
              _isMicEnabled = false;
            });
          }
        }
      } else {
        Logger.logError('Failed to establish connection, not enabling tracks');
      }
    } catch (e, stackTrace) {
      Logger.logError('Failed to connect to room', error: e, stackTrace: stackTrace);
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isConnecting = false;
        });
      }
    }
  }

  void _onRoomUpdate() {
    if (mounted) {
      setState(() {
        _updateParticipantTracks();
      });
    }
  }

  void _onRoomDisconnected() {
    Logger.logWarning('Room disconnected by server');
    if (mounted && !_wasDisconnected) {
      setState(() {
        _wasDisconnected = true;
      });
      _showDisconnectedDialog();
    }
  }

  void _onActiveSpeakersChanged(List<livekit.Participant> speakers) {
    if (!mounted) return;

    setState(() {
      _activeSpeakers = speakers;
    });

    // Auto-switch to active speaker if not manually selected
    if (speakers.isNotEmpty && !_manuallySelectedParticipant) {
      // Find first remote speaker (exclude local participant)
      final remoteSpeakers = speakers.where(
        (s) => s.sid != _room?.localParticipant?.sid,
      ).toList();

      if (remoteSpeakers.isNotEmpty) {
        setState(() {
          _stageParticipantId = remoteSpeakers.first.sid;
        });
      }
      // If only local participant is speaking, don't auto-select
      // Let _buildMainSpeakerView handle showing remote participants
    }
  }

  void _updateParticipantTracks() {
    final room = _room;
    if (room == null) return;

    final tracks = <ParticipantTrack>[];

    // Add local participant
    final localParticipant = room.localParticipant;
    if (localParticipant != null) {
      for (final trackPub in localParticipant.videoTrackPublications) {
        if (trackPub.track != null && trackPub.subscribed) {
          tracks.add(ParticipantTrack(
            participant: localParticipant,
            track: trackPub.track!,
            isLocal: true,
          ));
        }
      }
    }

    // Add remote participants
    for (final participant in room.remoteParticipants.values) {
      for (final trackPub in participant.videoTrackPublications) {
        if (trackPub.track != null && trackPub.subscribed) {
          tracks.add(ParticipantTrack(
            participant: participant,
            track: trackPub.track!,
            isLocal: false,
          ));
        }
      }
    }

    setState(() {
      _participantTracks = tracks;
    });
  }

  Future<void> _toggleMicrophone() async {
    if (_room?.localParticipant == null) return;

    final enabled = !_isMicEnabled;
    await _room!.localParticipant!.setMicrophoneEnabled(enabled);

    setState(() {
      _isMicEnabled = enabled;
    });
  }

  Future<void> _toggleCamera() async {
    if (_room?.localParticipant == null) return;

    final enabled = !_isCameraEnabled;
    await _room!.localParticipant!.setCameraEnabled(enabled);

    setState(() {
      _isCameraEnabled = enabled;
    });
  }

  Future<void> _switchCamera() async {
    if (_room?.localParticipant == null) return;

    try {
      final localParticipant = _room!.localParticipant!;

      // Toggle camera position
      final newPosition = _cameraPosition == livekit.CameraPosition.front
          ? livekit.CameraPosition.back
          : livekit.CameraPosition.front;

      // Find the local video track
      for (final publication in localParticipant.videoTrackPublications) {
        if (publication.track is livekit.LocalVideoTrack) {
          final track = publication.track as livekit.LocalVideoTrack;

          // Switch camera
          await track.setCameraPosition(newPosition);

          if (mounted) {
            setState(() {
              _cameraPosition = newPosition;
            });
          }

          Logger.logInfo('Switched camera to $newPosition');
          break;
        }
      }
    } catch (e) {
      Logger.logError('Failed to switch camera', error: e);
      // Silently ignore camera switch errors as they're not critical
    }
  }

  Future<void> _disconnect() async {
    final l10n = AppLocalizations.of(context)!;

    // Show confirmation dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.confirmLeaveTitle),
        content: Text(l10n.confirmLeaveMessage),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: Colors.red,
            ),
            child: Text(l10n.confirmLeave),
          ),
        ],
      ),
    );

    // Only disconnect if user confirmed
    if (confirmed == true) {
      Logger.logInfo('Disconnecting from room');
      await _room?.disconnect();
      if (mounted) {
        // Navigate back to meetings list instead of just popping
        Navigator.of(context).popUntil((route) => route.isFirst);
      }
    }
  }

  void _showSettingsBottomSheet() async {
    final l10n = AppLocalizations.of(context)!;

    // Get available audio devices
    List<livekit.MediaDevice> audioDevices = [];
    livekit.MediaDevice? selectedAudioDevice;
    String? currentDeviceId;

    try {
      audioDevices = await livekit.Hardware.instance.enumerateDevices();

      // Filter audio input devices
      audioDevices = audioDevices.where((d) => d.kind == 'audioinput').toList();

      // Get current audio device
      final localParticipant = _room?.localParticipant;
      if (localParticipant != null) {
        for (final pub in localParticipant.audioTrackPublications) {
          if (pub.track is livekit.LocalAudioTrack) {
            final track = pub.track as livekit.LocalAudioTrack;
            currentDeviceId = track.currentOptions.deviceId;
            if (currentDeviceId != null) {
              try {
                selectedAudioDevice = audioDevices.firstWhere(
                  (d) => d.deviceId == currentDeviceId,
                );
              } catch (_) {
                selectedAudioDevice = audioDevices.isNotEmpty ? audioDevices.first : null;
              }
            } else {
              selectedAudioDevice = audioDevices.isNotEmpty ? audioDevices.first : null;
            }
          }
        }
      }
    } catch (e) {
      Logger.logError('Failed to get audio devices', error: e);
    }

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.grey[900],
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                child: Row(
                  children: [
                    Text(
                      l10n.settings,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.white),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              const Divider(color: Colors.white24),

              // Audio source selection
              if (audioDevices.isNotEmpty)
                ListTile(
                  leading: const Icon(Icons.mic, color: Colors.white),
                  title: const Text(
                    'Audio Source',
                    style: TextStyle(color: Colors.white),
                  ),
                  subtitle: Text(
                    selectedAudioDevice?.label ?? 'Default',
                    style: const TextStyle(color: Colors.white70),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  enabled: _isMicEnabled,
                  onTap: _isMicEnabled
                      ? () {
                          Navigator.of(context).pop();
                          _showAudioSourceSelector(audioDevices, selectedAudioDevice);
                        }
                      : null,
                ),

              // Camera flip option
              ListTile(
                leading: const Icon(Icons.flip_camera_ios, color: Colors.white),
                title: Text(
                  l10n.flipCamera,
                  style: const TextStyle(color: Colors.white),
                ),
                subtitle: Text(
                  _cameraPosition == livekit.CameraPosition.front
                      ? 'Front Camera'
                      : 'Back Camera',
                  style: const TextStyle(color: Colors.white70),
                ),
                enabled: _isCameraEnabled,
                onTap: _isCameraEnabled
                    ? () {
                        Navigator.of(context).pop();
                        _switchCamera();
                      }
                    : null,
              ),
            ],
          ),
        ),
      ),
    );
  }

  void _showAudioSourceSelector(List<livekit.MediaDevice> devices, livekit.MediaDevice? currentDevice) {
    showModalBottomSheet(
      context: context,
      backgroundColor: Colors.grey[900],
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                child: Row(
                  children: [
                    const Text(
                      'Select Audio Source',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.white),
                      onPressed: () => Navigator.of(context).pop(),
                    ),
                  ],
                ),
              ),
              const Divider(color: Colors.white24),

              // List of audio devices
              ...devices.map((device) => ListTile(
                leading: Icon(
                  Icons.mic,
                  color: device.deviceId == currentDevice?.deviceId
                      ? const Color(0xFF26C6DA)
                      : Colors.white70,
                ),
                title: Text(
                  device.label.isNotEmpty ? device.label : 'Unknown Device',
                  style: TextStyle(
                    color: device.deviceId == currentDevice?.deviceId
                        ? const Color(0xFF26C6DA)
                        : Colors.white,
                    fontWeight: device.deviceId == currentDevice?.deviceId
                        ? FontWeight.bold
                        : FontWeight.normal,
                  ),
                ),
                trailing: device.deviceId == currentDevice?.deviceId
                    ? const Icon(Icons.check, color: Color(0xFF26C6DA))
                    : null,
                onTap: () async {
                  Navigator.of(context).pop();
                  await _switchAudioSource(device);
                },
              )),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _switchAudioSource(livekit.MediaDevice device) async {
    if (_room?.localParticipant == null) return;
    final l10n = AppLocalizations.of(context)!;

    try {
      final localParticipant = _room!.localParticipant!;

      // Find the local audio track
      for (final publication in localParticipant.audioTrackPublications) {
        if (publication.track is livekit.LocalAudioTrack) {
          final track = publication.track as livekit.LocalAudioTrack;

          // Switch audio source
          await track.setDeviceId(device.deviceId);

          Logger.logInfo('Switched audio source to ${device.label}');

          if (mounted) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(l10n.audioSourceChanged(device.label)),
                duration: const Duration(seconds: 2),
              ),
            );
          }
          break;
        }
      }
    } catch (e) {
      Logger.logError('Failed to switch audio source', error: e);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.failedToSwitchAudio),
            duration: const Duration(seconds: 3),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  void _showDisconnectedDialog() {
    final l10n = AppLocalizations.of(context)!;

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        title: Text(l10n.roomDisconnected),
        content: Text(l10n.disconnectedMessage),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pop(); // Close dialog
              Navigator.of(context).popUntil((route) => route.isFirst); // Go to home
            },
            child: Text(l10n.goToHome),
          ),
        ],
      ),
    );
  }

  @override
  void dispose() {
    // Disable wakelock when leaving the video call
    WakelockPlus.disable();

    _room?.removeListener(_onRoomUpdate);
    _room?.disconnect();
    _room?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black.withValues(alpha: 0.5),
        foregroundColor: Colors.white,
        title: Text(widget.meetingTitle),
        actions: [
          // Participants count button
          if (_room != null)
            TextButton.icon(
              onPressed: () {
                setState(() {
                  _showParticipantsList = !_showParticipantsList;
                });
              },
              icon: const Icon(Icons.people, color: Colors.white),
              label: Text(
                '${(_room!.remoteParticipants.length + 1)}',
                style: const TextStyle(color: Colors.white),
              ),
            ),
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: () {
              // Show settings bottom sheet
              _showSettingsBottomSheet();
            },
            tooltip: l10n.settings,
          ),
        ],
      ),
      body: _isConnecting
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const CircularProgressIndicator(color: Colors.white),
                  const SizedBox(height: 16),
                  Text(
                    l10n.connectingToMeeting,
                    style: const TextStyle(color: Colors.white),
                  ),
                ],
              ),
            )
          : _error != null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(24),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const Icon(
                          Icons.error_outline,
                          color: Colors.red,
                          size: 64,
                        ),
                        const SizedBox(height: 16),
                        Text(
                          l10n.connectionFailed,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _error!,
                          style: const TextStyle(color: Colors.white70),
                          textAlign: TextAlign.center,
                        ),
                        const SizedBox(height: 24),
                        ElevatedButton(
                          onPressed: _connectToRoom,
                          child: Text(l10n.retry),
                        ),
                      ],
                    ),
                  ),
                )
              : Stack(
                  children: [
                    // Full-screen main speaker view
                    _buildMainSpeakerView(),
                    // Participants list overlay
                    if (_showParticipantsList) _buildParticipantsOverlay(),
                  ],
                ),
      bottomNavigationBar: _room != null
          ? Container(
              color: Colors.black.withValues(alpha: 0.8),
              padding: const EdgeInsets.symmetric(vertical: 16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  _buildControlButton(
                    icon: _isMicEnabled ? Icons.mic : Icons.mic_off,
                    label: l10n.mic,
                    color: _isMicEnabled ? Colors.white : Colors.red,
                    onPressed: _toggleMicrophone,
                  ),
                  _buildControlButton(
                    icon: _isCameraEnabled ? Icons.videocam : Icons.videocam_off,
                    label: l10n.camera,
                    color: _isCameraEnabled ? Colors.white : Colors.red,
                    onPressed: _toggleCamera,
                  ),
                  _buildControlButton(
                    icon: Icons.call_end,
                    label: l10n.leave,
                    color: Colors.red,
                    onPressed: _disconnect,
                  ),
                ],
              ),
            )
          : null,
    );
  }

  Widget _buildMainSpeakerView() {
    final room = _room;
    if (room == null) return const SizedBox();

    // Determine which participant to show
    livekit.Participant? stageParticipant;

    if (_stageParticipantId != null) {
      // Show manually selected participant
      if (_stageParticipantId == room.localParticipant?.sid) {
        stageParticipant = room.localParticipant;
      } else {
        stageParticipant = room.remoteParticipants[_stageParticipantId];
      }
    }

    // If no stage participant, prioritize remote participants, but show local if alone
    if (stageParticipant == null) {
      // If only local participant exists, show local participant
      if (room.remoteParticipants.isEmpty) {
        stageParticipant = room.localParticipant;
      } else {
        // First try to find a remote participant with video
        ParticipantTrack? remoteTrack;
        try {
          remoteTrack = _participantTracks.firstWhere(
            (pt) => !pt.isLocal && pt.participant.isCameraEnabled(),
          );
        } catch (_) {
          // Try to find any remote participant
          try {
            remoteTrack = _participantTracks.firstWhere((pt) => !pt.isLocal);
          } catch (_) {
            // If no remote tracks, use first available track (including local)
            if (_participantTracks.isNotEmpty) {
              remoteTrack = _participantTracks.first;
            }
          }
        }

        if (remoteTrack != null) {
          stageParticipant = remoteTrack.participant;
        } else {
          // No tracks available, show local participant if available
          stageParticipant = room.localParticipant;
        }
      }
    }

    // If still no participant, show waiting view
    if (stageParticipant == null) {
      return _buildWaitingView();
    }

    // Check if participant has camera enabled
    final videoTrack = stageParticipant.videoTrackPublications
        .where((pub) => pub.track != null && pub.subscribed)
        .map((pub) => pub.track as livekit.VideoTrack)
        .firstOrNull;

    final displayName = stageParticipant.sid == room.localParticipant?.sid
        ? AppLocalizations.of(context)!.you
        : _getDisplayName(stageParticipant.identity);

    final isSpeaking = _activeSpeakers.any((s) => s.sid == stageParticipant?.sid);

    return Container(
      color: Colors.black,
      child: Stack(
        children: [
          // Video or avatar
          if (videoTrack != null)
            Center(
              child: livekit.VideoTrackRenderer(
                videoTrack,
                fit: livekit.VideoViewFit.contain,
              ),
            )
          else
            _buildLargeAvatar(displayName),

          // Speaking indicator at top (always visible)
          Positioned(
            top: 16,
            left: 16,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: isSpeaking
                    ? const Color(0xFF26C6DA).withValues(alpha: 0.9)
                    : Colors.black.withValues(alpha: 0.7),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (isSpeaking) ...[
                    const Icon(Icons.mic, size: 16, color: Colors.white),
                    const SizedBox(width: 4),
                  ],
                  Text(
                    displayName,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWaitingView() {
    final l10n = AppLocalizations.of(context)!;

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.people_outline,
            color: Colors.white54,
            size: 64,
          ),
          const SizedBox(height: 16),
          Text(
            l10n.waitingForParticipants,
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.7),
              fontSize: 16,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            l10n.participantsInRoom(_room?.remoteParticipants.length ?? 0),
            style: TextStyle(
              color: Colors.white.withValues(alpha: 0.5),
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLargeAvatar(String displayName) {
    String getInitials(String name) {
      final parts = name.trim().split(' ');
      if (parts.isEmpty) return '?';
      if (parts.length == 1) return parts[0][0].toUpperCase();
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }

    return Center(
      child: Container(
        width: 120,
        height: 120,
        decoration: BoxDecoration(
          gradient: const LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [Color(0xFF667eea), Color(0xFF764ba2)],
          ),
          shape: BoxShape.circle,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.3),
              blurRadius: 20,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Center(
          child: Text(
            getInitials(displayName),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 48,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildParticipantsOverlay() {
    final room = _room;
    if (room == null) return const SizedBox();

    final allParticipants = <livekit.Participant>[
      if (room.localParticipant != null) room.localParticipant!,
      ...room.remoteParticipants.values,
    ];

    return GestureDetector(
      onTap: () {
        setState(() {
          _showParticipantsList = false;
        });
      },
      child: Container(
        color: Colors.black.withValues(alpha: 0.7),
        child: SafeArea(
          child: Column(
            children: [
              // Header
              Container(
                padding: const EdgeInsets.all(16),
                color: Colors.black.withValues(alpha: 0.9),
                child: Row(
                  children: [
                    Text(
                      AppLocalizations.of(context)!.participants,
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.white),
                      onPressed: () {
                        setState(() {
                          _showParticipantsList = false;
                        });
                      },
                    ),
                  ],
                ),
              ),
              // Participants list
              Expanded(
                child: ListView.builder(
                  padding: const EdgeInsets.all(8),
                  itemCount: allParticipants.length,
                  itemBuilder: (context, index) {
                    final participant = allParticipants[index];
                    return _buildParticipantTile(participant);
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildParticipantTile(livekit.Participant participant) {
    final room = _room;
    if (room == null) return const SizedBox();

    final isLocal = participant.sid == room.localParticipant?.sid;
    final displayName = isLocal
        ? AppLocalizations.of(context)!.you
        : _getDisplayName(participant.identity);

    final isStage = participant.sid == _stageParticipantId;
    final isSpeaking = _activeSpeakers.any((s) => s.sid == participant.sid);

    // Get video track
    final videoTrack = participant.videoTrackPublications
        .where((pub) => pub.track != null && pub.subscribed)
        .map((pub) => pub.track as livekit.VideoTrack)
        .firstOrNull;

    // Get initials for avatar
    String getInitials(String name) {
      final parts = name.trim().split(' ');
      if (parts.isEmpty) return '?';
      if (parts.length == 1) return parts[0][0].toUpperCase();
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }

    return GestureDetector(
      onTap: () {
        setState(() {
          _stageParticipantId = participant.sid;
          _manuallySelectedParticipant = true; // User manually selected a participant
          _showParticipantsList = false;
        });
      },
      child: Container(
        height: 80,
        margin: const EdgeInsets.only(bottom: 8),
        decoration: BoxDecoration(
          color: isStage
              ? const Color(0xFF26C6DA).withValues(alpha: 0.2)
              : Colors.grey[900]?.withValues(alpha: 0.3),
          border: Border.all(
            color: isStage
                ? const Color(0xFF26C6DA)
                : isSpeaking
                    ? const Color(0xFF26C6DA).withValues(alpha: 0.5)
                    : Colors.white.withValues(alpha: 0.1),
            width: isStage ? 2 : 1,
          ),
          borderRadius: BorderRadius.circular(12),
        ),
        padding: const EdgeInsets.all(8),
        child: Row(
          children: [
            // Video thumbnail or avatar
            Container(
              width: 60,
              height: 60,
              decoration: BoxDecoration(
                color: Colors.black.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(8),
              ),
              clipBehavior: Clip.antiAlias,
              child: videoTrack != null
                  ? livekit.VideoTrackRenderer(
                      videoTrack,
                      fit: livekit.VideoViewFit.cover,
                    )
                  : Center(
                      child: Container(
                        width: 40,
                        height: 40,
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            begin: Alignment.topLeft,
                            end: Alignment.bottomRight,
                            colors: [Color(0xFF667eea), Color(0xFF764ba2)],
                          ),
                          shape: BoxShape.circle,
                        ),
                        child: Center(
                          child: Text(
                            getInitials(displayName),
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ),
                    ),
            ),
            const SizedBox(width: 12),
            // Name and status
            Expanded(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          displayName,
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.w500,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      // Mic indicator
                      if (!participant.isMicrophoneEnabled())
                        Container(
                          padding: const EdgeInsets.all(4),
                          decoration: BoxDecoration(
                            color: Colors.red.withValues(alpha: 0.8),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            Icons.mic_off,
                            size: 14,
                            color: Colors.white,
                          ),
                        ),
                    ],
                  ),
                  if (isSpeaking)
                    const SizedBox(height: 4),
                  if (isSpeaking)
                    Row(
                      children: [
                        Icon(
                          Icons.graphic_eq,
                          size: 14,
                          color: const Color(0xFF26C6DA),
                        ),
                        const SizedBox(width: 4),
                        Text(
                          AppLocalizations.of(context)!.speaking,
                          style: TextStyle(
                            color: const Color(0xFF26C6DA),
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                ],
              ),
            ),
            // Active indicator
            if (isStage)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFF26C6DA),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  AppLocalizations.of(context)!.active,
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildControlButton({
    required IconData icon,
    required String label,
    required Color color,
    required VoidCallback? onPressed,
  }) {
    final isDisabled = onPressed == null;
    return Opacity(
      opacity: isDisabled ? 0.5 : 1.0,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            width: 64,
            height: 64,
            decoration: BoxDecoration(
              color: color == Colors.red
                  ? Colors.red.withValues(alpha: 0.9)
                  : const Color(0xFF4CAF50).withValues(alpha: 0.15),
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.2),
                  blurRadius: 8,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: IconButton(
              icon: Icon(icon),
              color: color == Colors.red ? Colors.white : color,
              iconSize: 32,
              onPressed: onPressed,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            label,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 13,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }
}

class ParticipantTrack {
  final livekit.Participant participant;
  final livekit.Track track;
  final bool isLocal;

  ParticipantTrack({
    required this.participant,
    required this.track,
    required this.isLocal,
  });
}
