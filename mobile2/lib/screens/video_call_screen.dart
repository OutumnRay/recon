import 'package:flutter/material.dart';
import 'package:livekit_client/livekit_client.dart';
import '../utils/logger.dart';

class VideoCallScreen extends StatefulWidget {
  final String token;
  final String url;
  final String meetingTitle;

  const VideoCallScreen({
    super.key,
    required this.token,
    required this.url,
    required this.meetingTitle,
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
            tooltip: 'Switch Camera',
          ),
        ],
      ),
      body: _isConnecting
          ? const Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  CircularProgressIndicator(color: Colors.white),
                  SizedBox(height: 16),
                  Text(
                    'Connecting to meeting...',
                    style: TextStyle(color: Colors.white),
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
                        const Text(
                          'Connection Failed',
                          style: TextStyle(
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
                          child: const Text('Retry'),
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
                    label: 'Mic',
                    color: _isMicEnabled ? Colors.white : Colors.red,
                    onPressed: _toggleMicrophone,
                  ),
                  _buildControlButton(
                    icon: _isCameraEnabled ? Icons.videocam : Icons.videocam_off,
                    label: 'Camera',
                    color: _isCameraEnabled ? Colors.white : Colors.red,
                    onPressed: _toggleCamera,
                  ),
                  _buildControlButton(
                    icon: Icons.call_end,
                    label: 'Leave',
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
              'Waiting for participants...',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.7),
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              '${_room?.remoteParticipants.length ?? 0} participant(s) in room',
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

    return GridView.builder(
      padding: const EdgeInsets.all(8),
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: _participantTracks.length <= 4 ? 2 : 3,
        crossAxisSpacing: 8,
        mainAxisSpacing: 8,
      ),
      itemCount: _participantTracks.length,
      itemBuilder: (context, index) {
        return _buildParticipantView(_participantTracks[index]);
      },
    );
  }

  Widget _buildParticipantView(ParticipantTrack participantTrack) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.grey[900],
        borderRadius: BorderRadius.circular(8),
      ),
      child: Stack(
        children: [
          // Video
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: VideoTrackRenderer(
              participantTrack.track as VideoTrack,
              fit: VideoViewFit.cover,
            ),
          ),
          // Participant name overlay
          Positioned(
            bottom: 8,
            left: 8,
            child: Container(
              padding: const EdgeInsets.symmetric(
                horizontal: 8,
                vertical: 4,
              ),
              decoration: BoxDecoration(
                color: Colors.black.withValues(alpha: 0.6),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (participantTrack.isLocal)
                    const Padding(
                      padding: EdgeInsets.only(right: 4),
                      child: Icon(
                        Icons.person,
                        size: 12,
                        color: Colors.white,
                      ),
                    ),
                  Text(
                    participantTrack.isLocal
                        ? 'You'
                        : participantTrack.participant.identity,
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ],
              ),
            ),
          ),
          // Muted indicator
          if (!participantTrack.participant.isMicrophoneEnabled())
            Positioned(
              top: 8,
              right: 8,
              child: Container(
                padding: const EdgeInsets.all(4),
                decoration: BoxDecoration(
                  color: Colors.red.withValues(alpha: 0.8),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.mic_off,
                  size: 16,
                  color: Colors.white,
                ),
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
    required VoidCallback onPressed,
  }) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.1),
            shape: BoxShape.circle,
          ),
          child: IconButton(
            icon: Icon(icon),
            color: color,
            iconSize: 32,
            onPressed: onPressed,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          label,
          style: const TextStyle(
            color: Colors.white,
            fontSize: 12,
          ),
        ),
      ],
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
