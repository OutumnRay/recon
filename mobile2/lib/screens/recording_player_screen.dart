import 'package:flutter/material.dart';
import 'package:fijkplayer/fijkplayer.dart';
import 'package:provider/provider.dart';
import '../models/recording.dart';
import '../models/transcript.dart';
import '../utils/logger.dart';
import '../services/storage_service.dart';
import '../services/config_service.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/locale_service.dart';
import '../widgets/transcript_viewer.dart';
import '../widgets/memo_viewer.dart';
import '../l10n/app_localizations.dart';

class RecordingPlayerScreen extends StatefulWidget {
  final RoomRecording recording;
  final bool isTrack;
  final TrackRecording? track;
  final int initialTabIndex;

  const RecordingPlayerScreen({
    super.key,
    required this.recording,
    this.isTrack = false,
    this.track,
    this.initialTabIndex = 0,
  });

  static Widget routeBuilder(BuildContext context, Object? arguments) {
    final args = arguments as Map<String, dynamic>;
    return RecordingPlayerScreen(
      recording: args['recording'] as RoomRecording,
      isTrack: args['isTrack'] as bool? ?? false,
      track: args['track'] as TrackRecording?,
      initialTabIndex: args['initialTabIndex'] as int? ?? 0,
    );
  }

  @override
  State<RecordingPlayerScreen> createState() => _RecordingPlayerScreenState();
}

class _RecordingPlayerScreenState extends State<RecordingPlayerScreen>
    with SingleTickerProviderStateMixin {
  final FijkPlayer _player = FijkPlayer();
  bool _isLoading = true;
  String? _error;
  late TabController _tabController;
  RoomTranscripts? _transcripts;
  bool _isLoadingTranscripts = false;
  bool _isFullScreen = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(
      length: 3,
      vsync: this,
      initialIndex: widget.initialTabIndex,
    );
    _initializePlayer();
    _loadTranscripts();
  }

  Future<void> _loadTranscripts() async {
    if (widget.isTrack) {
      // Don't load transcripts for individual tracks
      return;
    }

    try {
      setState(() {
        _isLoadingTranscripts = true;
      });

      final configService = ConfigService();
      final baseUrl = await configService.getApiUrl();
      final apiClient = ApiClient(baseUrl: baseUrl);
      final meetingsService = MeetingsService(apiClient);

      Logger.logInfo('Loading transcripts for room', data: {'roomSid': widget.recording.roomSid});

      final transcripts = await meetingsService.getRoomTranscripts(widget.recording.roomSid);

      if (mounted) {
        setState(() {
          _transcripts = transcripts;
          _isLoadingTranscripts = false;
        });
      }
    } catch (e) {
      Logger.logError('Failed to load transcripts', error: e);
      if (mounted) {
        setState(() {
          _isLoadingTranscripts = false;
        });
      }
    }
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

      // Initialize FijkPlayer (don't autoplay)
      await _player.setDataSource(fullUrl, autoPlay: false);

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
    _tabController.dispose();
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
    final localeService = Provider.of<LocaleService>(context);

    return Scaffold(
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      appBar: AppBar(
        title: Text(_getTitle()),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
          tooltip: l10n.close,
        ),
        bottom: widget.isTrack
            ? null
            : TabBar(
                controller: _tabController,
                tabs: [
                  Tab(text: l10n.tabPlayer),
                  Tab(text: l10n.tabTranscript),
                  Tab(text: l10n.tabMemo),
                ],
              ),
      ),
      body: widget.isTrack
          ? _buildPlayerView(l10n)
          : TabBarView(
              controller: _tabController,
              children: [
                _buildPlayerView(l10n),
                _buildTranscriptView(localeService),
                _buildMemoView(localeService),
              ],
            ),
    );
  }

  Widget _buildPlayerView(AppLocalizations l10n) {
    return Container(
      color: Colors.black,
      child: Column(
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
                            panelBuilder: (FijkPlayer player, FijkData data, BuildContext context, Size viewSize, Rect texturePos) {
                              return _buildCustomPanel(player, data, context, viewSize, texturePos);
                            },
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

  Widget _buildTranscriptView(LocaleService localeService) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoadingTranscripts) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_transcripts == null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.error_outline, size: 64, color: Colors.grey),
            const SizedBox(height: 16),
            Text(
              l10n.failedToLoadTranscript,
              style: const TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadTranscripts,
              child: Text(l10n.retry),
            ),
          ],
        ),
      );
    }

    return TranscriptViewer(
      transcripts: _transcripts!,
      onSeekToTime: (double time) async {
        await _player.seekTo((time * 1000).toInt());
        await _player.start();
        // Switch back to player tab
        _tabController.animateTo(0);
      },
    );
  }

  Widget _buildMemoView(LocaleService localeService) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoadingTranscripts) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_transcripts == null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.error_outline, size: 64, color: Colors.grey),
            const SizedBox(height: 16),
            Text(
              l10n.failedToLoadMemo,
              style: const TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loadTranscripts,
              child: Text(l10n.retry),
            ),
          ],
        ),
      );
    }

    return MemoViewer(
      transcripts: _transcripts!,
      languageCode: localeService.locale.languageCode,
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

  Widget _buildCustomPanel(FijkPlayer player, FijkData data, BuildContext context, Size viewSize, Rect texturePos) {
    return Positioned.fill(
      child: GestureDetector(
        onTap: () {
          // Toggle play/pause on tap
          if (player.state == FijkState.started) {
            player.pause();
          } else {
            player.start();
          }
        },
        child: Container(
          color: Colors.transparent,
          child: Stack(
            children: [
              // Play/Pause and Fullscreen buttons in lower right corner
              Positioned(
                right: 16,
                bottom: 16,
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Play/Pause button
                    Container(
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.6),
                        shape: BoxShape.circle,
                      ),
                      child: IconButton(
                        icon: Icon(
                          player.state == FijkState.started
                              ? Icons.pause
                              : Icons.play_arrow,
                          color: Colors.white,
                        ),
                        iconSize: 32,
                        onPressed: () {
                          if (player.state == FijkState.started) {
                            player.pause();
                          } else {
                            player.start();
                          }
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    // Fullscreen/Exit Fullscreen button
                    Container(
                      decoration: BoxDecoration(
                        color: Colors.black.withValues(alpha: 0.6),
                        shape: BoxShape.circle,
                      ),
                      child: IconButton(
                        icon: Icon(
                          _isFullScreen ? Icons.fullscreen_exit : Icons.fullscreen,
                          color: Colors.white,
                        ),
                        iconSize: 32,
                        onPressed: () {
                          if (_isFullScreen) {
                            player.exitFullScreen();
                            setState(() {
                              _isFullScreen = false;
                            });
                          } else {
                            player.enterFullScreen();
                            setState(() {
                              _isFullScreen = true;
                            });
                          }
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }
}
