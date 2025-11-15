import 'package:flutter/material.dart';
import 'package:fijkplayer/fijkplayer.dart';
import 'package:dio/dio.dart';
import 'package:path_provider/path_provider.dart';
import 'dart:io';
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
  bool _isDownloading = false;
  double _downloadProgress = 0.0;

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

      Logger.logSuccess('Video player initialized');
    } catch (e) {
      Logger.logError('Failed to initialize video player', error: e);
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

  Future<void> _downloadRecording() async {
    final l10n = AppLocalizations.of(context)!;

    try {
      setState(() {
        _isDownloading = true;
        _downloadProgress = 0.0;
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

      // Get base URL and build full URL
      final configService = ConfigService();
      final baseUrl = await configService.getApiUrl();
      final fullUrl = playlistUrl.startsWith('http')
          ? playlistUrl
          : '$baseUrl$playlistUrl';

      // Get authentication token
      final storageService = StorageService();
      final token = await storageService.getToken();

      // Get download directory
      Directory? directory;
      if (Platform.isAndroid) {
        directory = Directory('/storage/emulated/0/Download');
        if (!await directory.exists()) {
          directory = await getExternalStorageDirectory();
        }
      } else {
        directory = await getApplicationDocumentsDirectory();
      }

      // Generate filename
      final timestamp = widget.recording.startedAt.toIso8601String().split('T')[0];
      final title = _getTitle().replaceAll(RegExp(r'[^\w\s-]'), '');
      final filename = 'recording_${title}_$timestamp.m3u8';
      final savePath = '${directory?.path}/$filename';

      // Download with Dio
      final dio = Dio();
      await dio.download(
        fullUrl,
        savePath,
        options: Options(
          headers: {
            if (token != null) 'Authorization': 'Bearer $token',
          },
        ),
        onReceiveProgress: (received, total) {
          if (total != -1 && mounted) {
            setState(() {
              _downloadProgress = received / total;
            });
          }
        },
      );

      if (mounted) {
        setState(() {
          _isDownloading = false;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.recordingSaved),
            backgroundColor: Colors.green,
          ),
        );
      }

      Logger.logSuccess('Recording downloaded successfully to: $savePath');
    } catch (e) {
      Logger.logError('Failed to download recording', error: e);

      if (mounted) {
        setState(() {
          _isDownloading = false;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.downloadFailed),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
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
        actions: [
          if (_isDownloading)
            Center(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: SizedBox(
                  width: 24,
                  height: 24,
                  child: CircularProgressIndicator(
                    value: _downloadProgress > 0 ? _downloadProgress : null,
                    strokeWidth: 2,
                    valueColor: const AlwaysStoppedAnimation<Color>(Colors.white),
                  ),
                ),
              ),
            )
          else
            IconButton(
              icon: const Icon(Icons.download),
              tooltip: l10n.downloadRecording,
              onPressed: _downloadRecording,
            ),
        ],
      ),
      body: Column(
        children: [
          // Video player
          Expanded(
            child: _isLoading
                ? const Center(
                    child: CircularProgressIndicator(
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : _error != null
                    ? _buildErrorWidget()
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
