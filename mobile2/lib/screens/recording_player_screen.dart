import 'package:flutter/material.dart';
import 'package:video_player/video_player.dart';
import 'package:chewie/chewie.dart';
import '../models/recording.dart';
import '../utils/logger.dart';
import '../services/storage_service.dart';

class RecordingPlayerScreen extends StatefulWidget {
  final RoomRecording recording;
  final bool isTrack;
  final TrackRecording? track;

  const RecordingPlayerScreen({
    super.key,
    required this.recording,
    this.isTrack = false,
    this.track,
  });

  static Widget routeBuilder(BuildContext context, Object? arguments) {
    final args = arguments as Map<String, dynamic>;
    return RecordingPlayerScreen(
      recording: args['recording'] as RoomRecording,
      isTrack: args['isTrack'] as bool? ?? false,
      track: args['track'] as TrackRecording?,
    );
  }

  @override
  State<RecordingPlayerScreen> createState() => _RecordingPlayerScreenState();
}

class _RecordingPlayerScreenState extends State<RecordingPlayerScreen> {
  VideoPlayerController? _videoPlayerController;
  ChewieController? _chewieController;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _initializePlayer();
  }

  Future<void> _initializePlayer() async {
    try {
      setState(() {
        _isLoading = true;
        _error = null;
      });

      // Get the playlist URL
      String playlistUrl;
      if (widget.isTrack && widget.track != null) {
        playlistUrl = widget.track!.playlistUrl;
      } else {
        playlistUrl = widget.recording.playlistUrl ?? '';
      }

      if (playlistUrl.isEmpty) {
        throw Exception('No playlist URL available');
      }

      Logger.logInfo('Initializing video player', data: {'url': playlistUrl});

      // Get base URL - use environment variable or default
      const baseUrl = String.fromEnvironment('API_URL', defaultValue: 'https://recontext.online');

      // Build full URL with authentication header
      final fullUrl = playlistUrl.startsWith('http')
          ? playlistUrl
          : '$baseUrl$playlistUrl';

      Logger.logInfo('Full playlist URL', data: {'url': fullUrl});

      // Get authentication token
      final storageService = StorageService();
      final token = await storageService.getToken();

      // Initialize video player with HLS stream
      _videoPlayerController = VideoPlayerController.networkUrl(
        Uri.parse(fullUrl),
        httpHeaders: {
          if (token != null) 'Authorization': 'Bearer $token',
        },
      );

      await _videoPlayerController!.initialize();

      // Initialize Chewie controller for better UI
      _chewieController = ChewieController(
        videoPlayerController: _videoPlayerController!,
        autoPlay: true,
        looping: false,
        allowFullScreen: true,
        allowMuting: true,
        showControls: true,
        materialProgressColors: ChewieProgressColors(
          playedColor: Theme.of(context).primaryColor,
          handleColor: Theme.of(context).primaryColor,
          backgroundColor: Colors.grey,
          bufferedColor: Theme.of(context).primaryColor.withOpacity(0.3),
        ),
        placeholder: Container(
          color: Colors.black,
          child: const Center(
            child: CircularProgressIndicator(),
          ),
        ),
        errorBuilder: (context, errorMessage) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error, color: Colors.red, size: 48),
                const SizedBox(height: 16),
                Text(
                  'Error playing video',
                  style: const TextStyle(color: Colors.white, fontSize: 18),
                ),
                const SizedBox(height: 8),
                Text(
                  errorMessage,
                  style: const TextStyle(color: Colors.grey, fontSize: 14),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          );
        },
      );

      setState(() {
        _isLoading = false;
      });

      Logger.logSuccess('Video player initialized');
    } catch (e) {
      Logger.logError('Failed to initialize video player', error: e);
      setState(() {
        _isLoading = false;
        _error = e.toString();
      });
    }
  }

  @override
  void dispose() {
    _chewieController?.dispose();
    _videoPlayerController?.dispose();
    super.dispose();
  }

  String _getTitle() {
    if (widget.isTrack && widget.track != null) {
      return widget.track!.participantName;
    }
    return 'Room Recording';
  }

  String _getDuration() {
    if (widget.isTrack && widget.track != null) {
      final duration = widget.track!.duration;
      if (duration != null) {
        final hours = duration.inHours;
        final minutes = duration.inMinutes.remainder(60);
        final seconds = duration.inSeconds.remainder(60);
        if (hours > 0) {
          return '${hours}h ${minutes}m ${seconds}s';
        }
        return '${minutes}m ${seconds}s';
      }
    } else {
      final duration = widget.recording.duration;
      if (duration != null) {
        final hours = duration.inHours;
        final minutes = duration.inMinutes.remainder(60);
        final seconds = duration.inSeconds.remainder(60);
        if (hours > 0) {
          return '${hours}h ${minutes}m ${seconds}s';
        }
        return '${minutes}m ${seconds}s';
      }
    }
    return 'Unknown duration';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        title: Text(_getTitle()),
        backgroundColor: Colors.black,
      ),
      body: Column(
        children: [
          // Video player
          Expanded(
            child: Center(
              child: _isLoading
                  ? const CircularProgressIndicator()
                  : _error != null
                      ? _buildErrorWidget()
                      : _chewieController != null
                          ? Chewie(controller: _chewieController!)
                          : const SizedBox.shrink(),
            ),
          ),
          // Info section
          Container(
            color: Colors.black87,
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _getTitle(),
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    const Icon(Icons.access_time, color: Colors.grey, size: 16),
                    const SizedBox(width: 4),
                    Text(
                      _getDuration(),
                      style: const TextStyle(color: Colors.grey, fontSize: 14),
                    ),
                    const SizedBox(width: 16),
                    const Icon(Icons.calendar_today, color: Colors.grey, size: 16),
                    const SizedBox(width: 4),
                    Text(
                      _formatDate(widget.recording.startedAt),
                      style: const TextStyle(color: Colors.grey, fontSize: 14),
                    ),
                  ],
                ),
                if (widget.isTrack && widget.track?.participant != null) ...[
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Icon(Icons.person, color: Colors.grey, size: 16),
                      const SizedBox(width: 4),
                      Text(
                        widget.track!.participant!.displayName,
                        style: const TextStyle(color: Colors.grey, fontSize: 14),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorWidget() {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, color: Colors.red, size: 64),
          const SizedBox(height: 16),
          const Text(
            'Failed to load recording',
            style: TextStyle(color: Colors.white, fontSize: 18),
          ),
          const SizedBox(height: 8),
          Text(
            _error ?? 'Unknown error',
            style: const TextStyle(color: Colors.grey, fontSize: 14),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: _initializePlayer,
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }
}
