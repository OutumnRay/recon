import 'package:flutter/material.dart';
import 'package:fijkplayer/fijkplayer.dart';
import '../models/recording.dart';
import '../utils/logger.dart';
import '../services/storage_service.dart';
import '../services/config_service.dart';
import '../l10n/app_localizations.dart';

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
  final FijkPlayer _player = FijkPlayer();
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

      Logger.logInfo('Initializing media player', data: {'url': playlistUrl});

      // Get base URL from config service
      final configService = ConfigService();
      final baseUrl = await configService.getApiUrl();

      // Build full URL
      final fullUrl = playlistUrl.startsWith('http')
          ? playlistUrl
          : '$baseUrl$playlistUrl';

      Logger.logInfo('Full playlist URL', data: {'url': fullUrl});

      // Get authentication token
      final storageService = StorageService();
      final token = await storageService.getToken();

      if (!mounted) return;

      // Set headers with authentication
      // FijkPlayer expects headers as a string in format "key:value\r\nkey:value"
      if (token != null) {
        final headersString = 'Authorization:Bearer $token';
        await _player.setOption(FijkOption.formatCategory, "headers", headersString);
      }

      // Initialize FijkPlayer
      await _player.setDataSource(fullUrl, autoPlay: true);

      setState(() {
        _isLoading = false;
      });

      Logger.logSuccess('Media player initialized');
    } catch (e) {
      Logger.logError('Failed to initialize media player', error: e);
      if (mounted) {
        setState(() {
          _isLoading = false;
          _error = e.toString();
        });
      }
    }
  }

  @override
  void dispose() {
    _player.release();
    super.dispose();
  }

  String _getTitle() {
    final l10n = AppLocalizations.of(context)!;
    if (widget.isTrack && widget.track != null) {
      return widget.track!.participantName;
    }
    return l10n.roomRecording;
  }

  String _getDuration() {
    final l10n = AppLocalizations.of(context)!;
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
    return l10n.unknownError;
  }

  bool get _isAudioOnly {
    if (widget.isTrack && widget.track != null) {
      return widget.track!.isAudioOnly;
    }
    return false;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        title: Text(_getTitle()),
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        leading: IconButton(
          icon: const Icon(Icons.close),
          onPressed: () => Navigator.of(context).pop(),
          tooltip: l10n.close,
        ),
      ),
      body: Column(
        children: [
          // Media player or audio-only indicator
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : _error != null
                    ? _buildErrorWidget()
                    : _isAudioOnly
                        ? _buildAudioOnlyWidget()
                        : FijkView(
                            player: _player,
                            color: Colors.black,
                            fit: FijkFit.contain,
                            panelBuilder: fijkPanel2Builder(
                              onBack: () => Navigator.of(context).pop(),
                            ),
                          ),
          ),
          // Info section
          Container(
            color: Colors.black87,
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Text(
                        _getTitle(),
                        style: const TextStyle(
                          color: Colors.white,
                          fontSize: 18,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    if (_isAudioOnly)
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                        decoration: BoxDecoration(
                          color: const Color(0xFF7C3AED),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            const Icon(Icons.audiotrack, size: 12, color: Colors.white),
                            const SizedBox(width: 4),
                            Text(
                              l10n.audioOnlyRecording,
                              style: const TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                            ),
                          ],
                        ),
                      ),
                  ],
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

  Widget _buildAudioOnlyWidget() {
    final l10n = AppLocalizations.of(context)!;
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              color: const Color(0xFF7C3AED).withValues(alpha: 0.2),
              shape: BoxShape.circle,
            ),
            child: const Icon(
              Icons.audiotrack,
              size: 64,
              color: Color(0xFF7C3AED),
            ),
          ),
          const SizedBox(height: 24),
          Text(
            l10n.audioOnlyRecording,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            l10n.playRecording,
            style: const TextStyle(
              color: Colors.grey,
              fontSize: 14,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorWidget() {
    final l10n = AppLocalizations.of(context)!;
    return Container(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, color: Colors.red, size: 64),
          const SizedBox(height: 16),
          Text(
            l10n.failedToLoadRecording,
            style: const TextStyle(color: Colors.white, fontSize: 18),
          ),
          const SizedBox(height: 8),
          Text(
            _error ?? l10n.unknownError,
            style: const TextStyle(color: Colors.grey, fontSize: 14),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: _initializePlayer,
            child: Text(l10n.retry),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }
}
