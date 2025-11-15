import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
import 'package:livekit_client/livekit_client.dart';
import '../utils/logger.dart';
import '../models/meeting.dart';

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
  Room? _room;
  List<ParticipantTrack> _participantTracks = [];
  bool _isMicEnabled = true;
  bool _isCameraEnabled = true;
  bool _isConnecting = true;
  String? _error;

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
        'token': widget.token.substring(0, 20) + '...',
      });

      // Create room with options
      final roomOptions = RoomOptions(
        adaptiveStream: true,
        dynacast: true,
        defaultCameraCaptureOptions: const CameraCaptureOptions(
          maxFrameRate: 30,
          params: VideoParametersPresets.h720_169,
        ),
        defaultAudioPublishOptions: const AudioPublishOptions(
          name: 'microphone',
        ),
        defaultVideoPublishOptions: const VideoPublishOptions(
          name: 'camera',
        ),
      );

      final room = Room(roomOptions: roomOptions);

      // Set up event listeners
      room.addListener(_onRoomUpdate);

      // Prepare connection for faster connection
      await room.prepareConnection(widget.url, widget.token);

      // Connect to room
      await room.connect(widget.url, widget.token);

      Logger.logSuccess('Connected to LiveKit room');

      if (mounted) {
        setState(() {
          _room = room;
          _isConnecting = false;
          _updateParticipantTracks();
        });
      }

      // Enable camera and microphone with error handling
      try {
        // Video will fail when running in iOS simulator
        await room.localParticipant?.setCameraEnabled(true);
        setState(() {
          _isCameraEnabled = true;
        });
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
        setState(() {
          _isMicEnabled = true;
        });
      } catch (error) {
        Logger.logWarning('Could not enable microphone: $error');
        if (mounted) {
          setState(() {
            _isMicEnabled = false;
          });
        }
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

    final localParticipant = _room!.localParticipant!;

    // Find the local video track
    for (final publication in localParticipant.videoTrackPublications) {
      if (publication.track is LocalVideoTrack) {
        final track = publication.track as LocalVideoTrack;
        // Toggle camera
        await track.setCameraPosition(CameraPosition.back);
        break;
      }
    }
  }

  Future<void> _disconnect() async {
    Logger.logInfo('Disconnecting from room');
    await _room?.disconnect();
    if (mounted) {
      Navigator.of(context).pop();
    }
  }

  @override
  void dispose() {
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
          IconButton(
            icon: const Icon(Icons.flip_camera_ios),
            onPressed: _switchCamera,
            tooltip: l10n.switchCamera,
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
              : _buildVideoGrid(),
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
                    icon: Icons.flip_camera_ios,
                    label: l10n.flipCamera,
                    color: Colors.white,
                    onPressed: _isCameraEnabled ? _switchCamera : null,
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

  Widget _buildVideoGrid() {
    final l10n = AppLocalizations.of(context)!;

    if (_participantTracks.isEmpty) {
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

    if (_participantTracks.length == 1) {
      return _buildParticipantView(_participantTracks[0]);
    }

    return ListView.builder(
      padding: const EdgeInsets.all(8),
      itemCount: _participantTracks.length,
      itemBuilder: (context, index) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 8),
          child: _buildParticipantView(_participantTracks[index]),
        );
      },
    );
  }

  Widget _buildParticipantView(ParticipantTrack participantTrack) {
    final l10n = AppLocalizations.of(context)!;
    final displayName = participantTrack.isLocal
        ? l10n.you
        : _getDisplayName(participantTrack.participant.identity);

    // Get initials for avatar
    String getInitials(String name) {
      final parts = name.trim().split(' ');
      if (parts.isEmpty) return '?';
      if (parts.length == 1) return parts[0][0].toUpperCase();
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }

    return Container(
      height: 70,
      decoration: BoxDecoration(
        color: Colors.grey[900]?.withOpacity(0.05),
        border: Border.all(
          color: Colors.white.withOpacity(0.1),
          width: 1,
        ),
        borderRadius: BorderRadius.circular(12),
      ),
      padding: const EdgeInsets.all(8),
      child: Row(
        children: [
          // Video/Avatar section (left)
          Container(
            width: 60,
            height: 60,
            decoration: BoxDecoration(
              color: Colors.black.withOpacity(0.3),
              borderRadius: BorderRadius.circular(8),
            ),
            clipBehavior: Clip.antiAlias,
            child: participantTrack.participant.isCameraEnabled()
                ? VideoTrackRenderer(
                    participantTrack.track as VideoTrack,
                    fit: VideoViewFit.cover,
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
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.3),
                            blurRadius: 8,
                            offset: const Offset(0, 2),
                          ),
                        ],
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
          // Participant info (right)
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
                          fontSize: 13,
                          fontWeight: FontWeight.w500,
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    // Muted indicator
                    if (!participantTrack.participant.isMicrophoneEnabled())
                      Container(
                        padding: const EdgeInsets.all(4),
                        decoration: BoxDecoration(
                          color: Colors.red.withOpacity(0.8),
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
              ],
            ),
          ),
        ],
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
                  ? Colors.red.withOpacity(0.9)
                  : const Color(0xFF4CAF50).withOpacity(0.15),
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.2),
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
  final Participant participant;
  final Track track;
  final bool isLocal;

  ParticipantTrack({
    required this.participant,
    required this.track,
    required this.isLocal,
  });
}
