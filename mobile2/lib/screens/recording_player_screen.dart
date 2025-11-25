import 'dart:async';
import 'package:flutter/material.dart';
import 'package:fijkplayer/fijkplayer.dart';
import 'package:provider/provider.dart';
import '../main.dart';
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
  final String? meetingId;

  const RecordingPlayerScreen({
    super.key,
    required this.recording,
    this.isTrack = false,
    this.track,
    this.initialTabIndex = 0,
    this.meetingId,
  });

  static Widget routeBuilder(BuildContext context, Object? arguments) {
    final args = arguments as Map<String, dynamic>;
    return RecordingPlayerScreen(
      recording: args['recording'] as RoomRecording,
      isTrack: args['isTrack'] as bool? ?? false,
      track: args['track'] as TrackRecording?,
      initialTabIndex: args['initialTabIndex'] as int? ?? 0,
      meetingId: args['meetingId'] as String?,
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
  bool _isDragging = false;
  double _dragPosition = 0.0;
  Timer? _positionTimer;
  int _lastKnownPosition = 0;

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

    // Start timer to update UI periodically (needed for smooth progress bar)
    _positionTimer = Timer.periodic(const Duration(milliseconds: 500), (_) {
      if (mounted && !_isDragging) {
        final currentPos = _player.currentPos.inMilliseconds;
        if (currentPos != _lastKnownPosition) {
          setState(() {
            _lastKnownPosition = currentPos;
          });
        }
      }
    });
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
      final apiClient = ApiClient(baseUrl: baseUrl, navigatorKey: navigatorKey);
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

  Future<void> _generateSummary() async {
    if (widget.meetingId == null) {
      Logger.logWarning('Cannot generate summary: meetingId is null');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.failedToGenerateSummary),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    try {
      final configService = ConfigService();
      final baseUrl = await configService.getApiUrl();
      final apiClient = ApiClient(baseUrl: baseUrl, navigatorKey: navigatorKey);
      final meetingsService = MeetingsService(apiClient);

      Logger.logInfo('Generating summary for meeting', data: {'meetingId': widget.meetingId});

      // Call the API to generate summary
      await meetingsService.generateMeetingSummary(widget.meetingId!);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.summaryGenerationStarted),
            backgroundColor: Colors.green,
          ),
        );

        // Reload transcripts after a short delay to get the updated summary
        await Future.delayed(const Duration(seconds: 2));
        await _loadTranscripts();
      }
    } catch (e) {
      Logger.logError('Failed to generate summary', error: e);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(AppLocalizations.of(context)!.failedToGenerateSummary),
            backgroundColor: Colors.red,
          ),
        );
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

      // If no composite video but has individual tracks, don't initialize player
      // The placeholder will be shown instead
      if (playlistUrl.isEmpty) {
        if (!widget.isTrack && widget.recording.tracks.isNotEmpty) {
          // This is expected - composite video not available, show placeholder
          setState(() {
            _isLoading = false;
          });
          return;
        }
        // No tracks available at all
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
    _positionTimer?.cancel();
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
    // Check if composite video is available
    final hasCompositeVideo = !widget.isTrack &&
                               widget.recording.playlistUrl != null &&
                               widget.recording.playlistUrl!.isNotEmpty;
    final hasIndividualTracks = !widget.isTrack && widget.recording.tracks.isNotEmpty;

    // If no composite video but has individual tracks, show placeholder
    final shouldShowPlaceholder = !widget.isTrack &&
                                   !hasCompositeVideo &&
                                   hasIndividualTracks;

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
                    : shouldShowPlaceholder
                        ? _buildCompositeVideoPlaceholder(l10n)
                        : Stack(
                            children: [
                              // Always show FijkView for both audio and video
                              FijkView(
                                player: _player,
                                color: Colors.black,
                                fit: FijkFit.contain,
                                panelBuilder: (FijkPlayer player, FijkData data, BuildContext context, Size viewSize, Rect texturePos) {
                                  return _buildCustomPanel(player, data, context, viewSize, texturePos);
                                },
                              ),
                              // Overlay audio-only indicator when it's audio
                              if (_isAudioOnly)
                                Center(
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
                                    ],
                                  ),
                                ),
                            ],
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
      return RefreshIndicator(
        onRefresh: _loadTranscripts,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          child: SizedBox(
            height: MediaQuery.of(context).size.height - 200,
            child: Center(
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
            ),
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadTranscripts,
      child: TranscriptViewer(
        transcripts: _transcripts!,
        onSeekToTime: (double time) async {
          await _player.seekTo((time * 1000).toInt());
          await _player.start();
          // Switch back to player tab
          _tabController.animateTo(0);
        },
      ),
    );
  }

  Widget _buildMemoView(LocaleService localeService) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoadingTranscripts) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_transcripts == null) {
      return RefreshIndicator(
        onRefresh: _loadTranscripts,
        child: SingleChildScrollView(
          physics: const AlwaysScrollableScrollPhysics(),
          child: SizedBox(
            height: MediaQuery.of(context).size.height - 200,
            child: Center(
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
            ),
          ),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadTranscripts,
      child: MemoViewer(
        transcripts: _transcripts!,
        languageCode: localeService.locale.languageCode,
        onGenerateSummary: widget.meetingId != null ? _generateSummary : null,
      ),
    );
  }

  Widget _buildCompositeVideoPlaceholder(AppLocalizations l10n) {
    // Group tracks by participant
    final participantTracks = <String, List<TrackRecording>>{};
    for (final track in widget.recording.tracks) {
      final key = track.participant?.displayName ?? track.participantId;
      participantTracks.putIfAbsent(key, () => []).add(track);
    }

    return Container(
      color: Colors.black,
      padding: const EdgeInsets.all(24),
      child: SingleChildScrollView(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Icon
            Container(
              width: 100,
              height: 100,
              decoration: BoxDecoration(
                color: const Color(0xFF2563EB).withValues(alpha: 0.2),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.video_library_outlined,
                size: 56,
                color: Color(0xFF2563EB),
              ),
            ),
            const SizedBox(height: 24),

            // Title
            Text(
              l10n.compositeVideoNotAvailable,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),

            // Description
            Text(
              l10n.viewIndividualTracksDescription,
              style: const TextStyle(
                color: Colors.grey,
                fontSize: 14,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),

            // Individual tracks list
            Text(
              l10n.availableParticipantTracks,
              style: const TextStyle(
                color: Colors.white70,
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 16),

            // Participant tracks
            ...participantTracks.entries.map((entry) {
              final participantName = entry.key;
              final tracks = entry.value;
              final hasVideo = tracks.any((t) => t.type == 'video');
              final hasAudio = tracks.any((t) => t.type == 'audio');

              return Container(
                margin: const EdgeInsets.only(bottom: 12),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: Colors.white.withValues(alpha: 0.2),
                  ),
                ),
                child: Row(
                  children: [
                    // Avatar
                    CircleAvatar(
                      radius: 24,
                      backgroundColor: const Color(0xFF2563EB).withValues(alpha: 0.3),
                      child: Text(
                        participantName.isNotEmpty
                            ? participantName[0].toUpperCase()
                            : '?',
                        style: const TextStyle(
                          color: Color(0xFF2563EB),
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    const SizedBox(width: 16),

                    // Participant info
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            participantName,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(height: 6),
                          Row(
                            children: [
                              if (hasVideo)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: const Color(0xFF2563EB).withValues(alpha: 0.3),
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: const Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(
                                        Icons.videocam,
                                        size: 14,
                                        color: Color(0xFF93C5FD),
                                      ),
                                      SizedBox(width: 4),
                                      Text(
                                        'Video',
                                        style: TextStyle(
                                          fontSize: 12,
                                          color: Color(0xFF93C5FD),
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              if (hasVideo && hasAudio) const SizedBox(width: 6),
                              if (hasAudio)
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 8,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: const Color(0xFF2563EB).withValues(alpha: 0.3),
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: const Row(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Icon(
                                        Icons.mic,
                                        size: 14,
                                        color: Color(0xFF93C5FD),
                                      ),
                                      SizedBox(width: 4),
                                      Text(
                                        'Audio',
                                        style: TextStyle(
                                          fontSize: 12,
                                          color: Color(0xFF93C5FD),
                                          fontWeight: FontWeight.w500,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                            ],
                          ),
                        ],
                      ),
                    ),

                    // Play button
                    IconButton(
                      onPressed: () {
                        // Prioritize video track, fallback to audio
                        TrackRecording? trackToPlay;
                        if (hasVideo) {
                          trackToPlay = tracks.firstWhere(
                            (t) => t.type == 'video',
                            orElse: () => tracks.first,
                          );
                        } else if (hasAudio) {
                          trackToPlay = tracks.firstWhere(
                            (t) => t.type == 'audio',
                            orElse: () => tracks.first,
                          );
                        } else {
                          trackToPlay = tracks.isNotEmpty ? tracks.first : null;
                        }

                        if (trackToPlay == null) return;

                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) => RecordingPlayerScreen(
                              recording: widget.recording,
                              isTrack: true,
                              track: trackToPlay,
                              meetingId: widget.meetingId,
                            ),
                          ),
                        );
                      },
                      icon: const Icon(
                        Icons.play_circle_filled,
                        color: Color(0xFF2563EB),
                        size: 36,
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ),
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
    final duration = player.value.duration.inMilliseconds;
    final position = _lastKnownPosition > 0 ? _lastKnownPosition : player.currentPos.inMilliseconds;

    // Use drag position if dragging, otherwise use actual position
    final displayProgress = _isDragging ? _dragPosition : (duration > 0 ? position / duration : 0.0);
    final displayPosition = _isDragging ? (_dragPosition * duration).toInt() : position;

    String formatDuration(Duration d) {
      String twoDigits(int n) => n.toString().padLeft(2, '0');
      final hours = d.inHours;
      final minutes = d.inMinutes.remainder(60);
      final seconds = d.inSeconds.remainder(60);
      return hours > 0 ? '$hours:${twoDigits(minutes)}:${twoDigits(seconds)}' : '$minutes:${twoDigits(seconds)}';
    }

    return Positioned.fill(
          child: GestureDetector(
            onTap: () {
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
                  // Progress bar
                  Positioned(
                    left: 0,
                    right: 0,
                    bottom: 72,
                    child: Container(
                      color: Colors.black.withValues(alpha: 0.5),
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(formatDuration(Duration(milliseconds: displayPosition)), style: const TextStyle(color: Colors.white, fontSize: 12)),
                              Text(formatDuration(player.value.duration), style: const TextStyle(color: Colors.white, fontSize: 12)),
                            ],
                          ),
                          SliderTheme(
                            data: SliderThemeData(
                              trackHeight: 3,
                              thumbShape: const RoundSliderThumbShape(enabledThumbRadius: 6),
                              overlayShape: const RoundSliderOverlayShape(overlayRadius: 12),
                              activeTrackColor: Colors.blue,
                              inactiveTrackColor: Colors.white.withValues(alpha: 0.3),
                              thumbColor: Colors.blue,
                              overlayColor: Colors.blue.withValues(alpha: 0.3),
                            ),
                            child: Slider(
                              value: displayProgress.clamp(0.0, 1.0),
                              onChangeStart: (value) {
                                setState(() {
                                  _isDragging = true;
                                  _dragPosition = value;
                                });
                              },
                              onChanged: (value) {
                                setState(() {
                                  _dragPosition = value;
                                });
                              },
                              onChangeEnd: (value) async {
                                final seekPosition = (value * duration).toInt();
                                await player.seekTo(seekPosition);
                                // Update last known position immediately
                                _lastKnownPosition = seekPosition;
                                // Small delay to ensure player position updates before releasing drag state
                                await Future.delayed(const Duration(milliseconds: 200));
                                if (mounted) {
                                  setState(() {
                                    _isDragging = false;
                                  });
                                }
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  // Control buttons
                  Positioned(
                    right: 16,
                    bottom: 16,
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Container(
                          decoration: BoxDecoration(color: Colors.black.withValues(alpha: 0.6), shape: BoxShape.circle),
                          child: IconButton(
                            icon: Icon(player.state == FijkState.started ? Icons.pause : Icons.play_arrow, color: Colors.white),
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
                        Container(
                          decoration: BoxDecoration(color: Colors.black.withValues(alpha: 0.6), shape: BoxShape.circle),
                          child: IconButton(
                            icon: Icon(_isFullScreen ? Icons.fullscreen_exit : Icons.fullscreen, color: Colors.white),
                            iconSize: 32,
                            onPressed: () {
                              if (_isFullScreen) {
                                player.exitFullScreen();
                                setState(() => _isFullScreen = false);
                              } else {
                                player.enterFullScreen();
                                setState(() => _isFullScreen = true);
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
