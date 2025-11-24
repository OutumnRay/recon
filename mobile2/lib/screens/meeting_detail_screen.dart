import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import '../main.dart';
import '../models/meeting.dart';
import '../models/recording.dart';
import '../models/task.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/config_service.dart';
import '../services/task_service.dart';
import '../services/auth_service.dart';
import '../widgets/error_display.dart';
import '../widgets/task_card.dart';
import '../theme/app_colors.dart';
import 'video_call_screen.dart';
import 'recording_player_screen.dart';

class MeetingDetailScreen extends StatefulWidget {
  final String meetingId;

  const MeetingDetailScreen({
    super.key,
    required this.meetingId,
  });

  @override
  State<MeetingDetailScreen> createState() => _MeetingDetailScreenState();
}

class _MeetingDetailScreenState extends State<MeetingDetailScreen>
    with SingleTickerProviderStateMixin {
  final _configService = ConfigService();
  late MeetingsService _meetingsService;
  late ApiClient _apiClient;
  late TabController _tabController;
  String? _publicBaseUrl;

  MeetingWithDetails? _meeting;
  bool _isLoading = true;
  String? _error;
  bool _isParticipantsExpanded = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _initService();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _initService() async {
    final apiUrl = await _configService.getApiUrl();
    _publicBaseUrl = apiUrl.replaceAll('/api/v1', '');
    _apiClient = ApiClient(baseUrl: apiUrl, navigatorKey: navigatorKey);
    _meetingsService = MeetingsService(_apiClient);
    _loadMeeting();
  }

  Future<void> _loadMeeting() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final meeting = await _meetingsService.getMeeting(widget.meetingId);
      if (mounted) {
        setState(() {
          _meeting = meeting;
          _isLoading = false;
        });
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = e.message;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  Future<void> _joinMeeting() async {
    if (_meeting == null) return;

    // Show loading indicator
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => const Center(
        child: CircularProgressIndicator(
          color: Color(0xFF26C6DA),
        ),
      ),
    );

    try {
      // Get LiveKit token
      final liveKitToken = await _meetingsService.getLiveKitToken(widget.meetingId);

      if (!mounted) return;

      // Close loading dialog
      Navigator.of(context).pop();

      // Navigate to video call screen
      await Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => VideoCallScreen(
            token: liveKitToken.token,
            url: liveKitToken.url,
            meetingTitle: _meeting!.title,
            participants: _meeting!.participants,
          ),
        ),
      );

      // Refresh meeting details after call ends
      _loadMeeting();
    } on ApiException catch (e) {
      if (!mounted) return;

      final l10n = AppLocalizations.of(context)!;

      // Close loading dialog
      Navigator.of(context).pop();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${l10n.failedToJoin}: ${e.message}'),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 5),
        ),
      );
    } catch (e) {
      if (!mounted) return;

      final l10n = AppLocalizations.of(context)!;

      // Close loading dialog
      Navigator.of(context).pop();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${l10n.error}: $e'),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 5),
        ),
      );
    }
  }

  Future<void> _copyAnonymousLink() async {
    final meeting = _meeting;
    if (meeting == null) return;

    final l10n = AppLocalizations.of(context)!;

    try {
      var baseUrl = _publicBaseUrl;
      if (baseUrl == null || baseUrl.isEmpty) {
        final apiUrl = await _configService.getApiUrl();
        baseUrl = apiUrl.replaceAll('/api/v1', '');
        _publicBaseUrl = baseUrl;
      }

      final anonymousLink = '$baseUrl/meeting/${meeting.id}/join';
      await Clipboard.setData(ClipboardData(text: anonymousLink));

      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(l10n.anonymousLinkCopied),
          backgroundColor: AppColors.success,
        ),
      );
    } catch (e) {
      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${l10n.error}: $e'),
          backgroundColor: AppColors.danger,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    final canJoin = _meeting != null &&
        (_meeting!.isPermanent ||
            _meeting!.recurrence == 'permanent' ||
            _meeting!.status == 'active' ||
            _meeting!.status == 'scheduled' ||
            _meeting!.status == 'in_progress') &&
        _meeting!.status != 'cancelled';

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.meetingDetailsTitle),
        backgroundColor: Colors.white,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        scrolledUnderElevation: 0,
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppColors.primary500,
          labelColor: AppColors.primary600,
          unselectedLabelColor: AppColors.textSecondary,
          tabs: [
            Tab(text: l10n.tabInfo),
            Tab(text: l10n.tabRecordings),
            const Tab(text: 'Tasks'),
          ],
        ),
      ),
      body: SafeArea(
        child: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? FullScreenError(
                    error: _error!,
                    onRetry: _loadMeeting,
                    title: l10n.failedToLoadMeetings,
                  )
                : _meeting == null
                    ? Center(child: Text(l10n.noMeetingsFound))
                    : TabBarView(
                        controller: _tabController,
                        children: [
                          _buildMeetingContent(),
                          _buildRecordingsTab(),
                          _buildTasksTab(),
                        ],
                      ),
      ),
      floatingActionButton: canJoin
          ? FloatingActionButton.extended(
              onPressed: _joinMeeting,
              icon: const Icon(Icons.video_call_rounded),
              label: Text(l10n.joinMeeting),
              backgroundColor: AppColors.success,
            )
          : null,
    );
  }

  Widget _buildRecordingsTab() {
    final l10n = AppLocalizations.of(context)!;

    return FutureBuilder<List<RoomRecording>>(
      future: _meetingsService.getMeetingRecordings(widget.meetingId),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: _buildCard(
              padding: const EdgeInsets.all(24),
              margin: const EdgeInsets.all(20),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.error_outline, size: 48, color: AppColors.danger),
                  const SizedBox(height: 12),
                  Text(
                    l10n.failedToLoadRecordings,
                    style: Theme.of(context).textTheme.titleMedium,
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 8),
                  Text(snapshot.error.toString(),
                      textAlign: TextAlign.center,
                      style: Theme.of(context)
                          .textTheme
                          .bodySmall
                          ?.copyWith(color: AppColors.textSecondary)),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => setState(() {}),
                    child: Text(l10n.retryLoadRecordings),
                  ),
                ],
              ),
            ),
          );
        }

        final recordings = snapshot.data ?? [];
        if (recordings.isEmpty) {
          return Center(
            child: _buildCard(
              padding: const EdgeInsets.all(32),
              margin: const EdgeInsets.all(20),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.videocam_off,
                      size: 56, color: AppColors.textTertiary),
                  const SizedBox(height: 12),
                  Text(l10n.noRecordingsFound,
                      style: Theme.of(context).textTheme.titleMedium),
                  const SizedBox(height: 8),
                  Text(
                    l10n.recordingsHint,
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: AppColors.textSecondary),
                    textAlign: TextAlign.center,
                  ),
                ],
              ),
            ),
          );
        }

        return ListView.builder(
          padding: const EdgeInsets.fromLTRB(20, 24, 20, 120),
          itemCount: recordings.length,
          itemBuilder: (context, index) =>
              _buildRecordingCard(recordings[index], index + 1),
        );
      },
    );
  }

  Widget _buildRecordingCard(RoomRecording recording, int sessionNumber) {
    final l10n = AppLocalizations.of(context)!;
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat.yMMMd(locale);
    final timeFormat = DateFormat.Hm(locale);

    final startedAt = recording.startedAt;
    final endedAt = recording.endedAt;

    int? durationMinutes;
    if (endedAt != null) {
      durationMinutes = endedAt.difference(startedAt).inMinutes;
    }

    return _buildCard(
      margin: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(
                  color: AppColors.primary50,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  l10n.sessionNumber(sessionNumber),
                  style: const TextStyle(
                    color: AppColors.primary600,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                decoration: BoxDecoration(
                  color: _getRecordingStatusColor(recording.status).withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  _getLocalizedStatus(recording.status, l10n),
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: _getRecordingStatusColor(recording.status),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
                Icon(Icons.calendar_today, size: 16, color: Colors.grey[600]),
                const SizedBox(width: 8),
                Text(
                  dateFormat.format(startedAt),
                  style: const TextStyle(fontSize: 14),
                ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
                Icon(Icons.access_time, size: 16, color: Colors.grey[600]),
                const SizedBox(width: 8),
                Text(
                  '${timeFormat.format(startedAt)} ${endedAt != null ? '- ${timeFormat.format(endedAt)}' : ''}',
                  style: const TextStyle(fontSize: 14),
                ),
                if (durationMinutes != null) ...[
                  const SizedBox(width: 16),
                  Text(
                    l10n.recordingDuration(durationMinutes),
                    style: TextStyle(fontSize: 14, color: Colors.grey[600]),
                  ),
                ],
            ],
          ),
          // Show participant tracks if available
          if (recording.tracks.isNotEmpty) ...[
            const SizedBox(height: 16),
            // Group tracks by participant
            ...() {
              final participantTracks = <String, List<TrackRecording>>{};
              for (final track in recording.tracks) {
                final key = track.participant?.displayName ?? 'Unknown';
                participantTracks.putIfAbsent(key, () => []).add(track);
              }

              return participantTracks.entries.map((entry) {
                final participantName = entry.key;
                final tracks = entry.value;
                final hasVideo = tracks.any((t) => t.isVideo);
                final hasAudio = tracks.any((t) => t.isAudioOnly);

                return Container(
                  margin: const EdgeInsets.only(bottom: 8),
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.surface,
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: AppColors.border),
                  ),
                  child: Row(
                    children: [
                      CircleAvatar(
                        radius: 16,
                        backgroundColor: AppColors.primary100,
                        child: Text(
                          participantName.isNotEmpty ? participantName[0].toUpperCase() : '?',
                          style: const TextStyle(
                            color: AppColors.primary600,
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              participantName,
                              style: const TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                            const SizedBox(height: 4),
                            Row(
                              children: [
                                if (hasVideo)
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: AppColors.primary50,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(Icons.videocam, size: 12, color: AppColors.primary600),
                                        const SizedBox(width: 2),
                                        Text(
                                          'Video',
                                          style: TextStyle(
                                            fontSize: 10,
                                            color: AppColors.primary600,
                                            fontWeight: FontWeight.w500,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                if (hasVideo && hasAudio) const SizedBox(width: 4),
                                if (hasAudio)
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(
                                      color: AppColors.primary50,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(Icons.mic, size: 12, color: AppColors.primary600),
                                        const SizedBox(width: 2),
                                        Text(
                                          'Audio',
                                          style: TextStyle(
                                            fontSize: 10,
                                            color: AppColors.primary600,
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
                    ],
                  ),
                );
              }).toList();
            }(),
            const SizedBox(height: 12),
            // Compact "View Session" button
            Align(
              alignment: Alignment.centerRight,
              child: ElevatedButton.icon(
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => RecordingPlayerScreen(
                        recording: recording,
                        initialTabIndex: 0,
                      ),
                    ),
                  );
                },
                icon: const Icon(Icons.arrow_forward, size: 16),
                label: Text(l10n.viewSession),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.primary50,
                  foregroundColor: AppColors.primary600,
                  elevation: 0,
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  minimumSize: Size.zero,
                  tapTargetSize: MaterialTapTargetSize.shrinkWrap,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                    side: BorderSide(color: AppColors.primary200),
                  ),
                ),
              ),
            ),
          ] else if (recording.playlistUrl != null) ...[
            // Fallback: Show icon buttons for room recording only (no tracks)
            const SizedBox(height: 16),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => RecordingPlayerScreen(
                          recording: recording,
                          initialTabIndex: 0,
                        ),
                      ),
                    );
                  },
                  icon: const Icon(Icons.play_circle_outlined),
                  iconSize: 32,
                  color: AppColors.primary500,
                  style: IconButton.styleFrom(
                    backgroundColor: AppColors.primary100,
                    padding: const EdgeInsets.all(12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  tooltip: l10n.tabPlayer,
                ),
                const SizedBox(width: 12),
                IconButton(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => RecordingPlayerScreen(
                          recording: recording,
                          initialTabIndex: 1,
                        ),
                      ),
                    );
                  },
                  icon: const Icon(Icons.text_snippet_outlined),
                  iconSize: 32,
                  color: AppColors.primary500,
                  style: IconButton.styleFrom(
                    backgroundColor: AppColors.primary100,
                    padding: const EdgeInsets.all(12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  tooltip: l10n.tabTranscript,
                ),
                const SizedBox(width: 12),
                IconButton(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => RecordingPlayerScreen(
                          recording: recording,
                          initialTabIndex: 2,
                        ),
                      ),
                    );
                  },
                  icon: const Icon(Icons.auto_awesome_outlined),
                  iconSize: 32,
                  color: AppColors.primary500,
                  style: IconButton.styleFrom(
                    backgroundColor: AppColors.primary100,
                    padding: const EdgeInsets.all(12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  tooltip: l10n.tabMemo,
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Color _getRecordingStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'recording':
        return const Color(0xFF991B1B); // Red for recording
      case 'completed':
        return const Color(0xFF059669); // Green
      case 'processing':
        return const Color(0xFF92400E); // Amber/brown
      case 'failed':
        return const Color(0xFF991B1B); // Red
      default:
        return const Color(0xFF6B7280); // Gray
    }
  }

  String _getLocalizedStatus(String status, AppLocalizations l10n) {
    switch (status.toLowerCase()) {
      case 'recording':
        return l10n.sessionStatusRecording;
      case 'completed':
        return l10n.sessionStatusCompleted;
      case 'processing':
        return l10n.sessionStatusProcessing;
      case 'failed':
        return l10n.sessionStatusFailed;
      case 'finished':
        return l10n.sessionStatusFinished;
      default:
        return status.toUpperCase();
    }
  }

  Widget _buildMeetingContent() {
    final l10n = AppLocalizations.of(context)!;
    final meeting = _meeting!;
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat.yMMMMEEEEd(locale);
    final timeFormat = DateFormat.Hm(locale);
    final timestampFormat = DateFormat.yMMMd(locale).add_Hm();
    final recurrenceLabel = meeting.isPermanent ||
            (meeting.recurrence?.toLowerCase() == 'permanent')
        ? l10n.recurrencePermanent
        : (meeting.recurrence != null &&
                meeting.recurrence!.toLowerCase() != 'none'
            ? _getRecurrenceText(meeting.recurrence!)
            : l10n.recurrenceNone);
    final participantsSummary = l10n.participantsSummary(
      meeting.participants.length,
      meeting.activeParticipantsCount,
    );
    final heroStats = <Widget>[
      _buildHeroStat(
        context,
        icon: Icons.schedule_rounded,
        label: l10n.meetingDuration,
        value: '${meeting.duration} ${l10n.minutes}',
      ),
      _buildHeroStat(
        context,
        icon: Icons.people_outline_rounded,
        label: l10n.participants,
        value: participantsSummary,
      ),
      _buildHeroStat(
        context,
        icon: Icons.loop_rounded,
        label: l10n.meetingRecurrence,
        value: recurrenceLabel,
      ),
      _buildHeroStat(
        context,
        icon: Icons.category_outlined,
        label: l10n.type,
        value: _getMeetingTypeText(meeting.type),
      ),
    ];
    if (meeting.allowAnonymous) {
      heroStats.add(
        _buildHeroStat(
          context,
          icon: Icons.lock_open_rounded,
          label: l10n.allowsAnonymousJoin,
          value: l10n.enabled,
        ),
      );
    }

    final anonymousLink = _publicBaseUrl != null
        ? '${_publicBaseUrl!}/meeting/${meeting.id}/join'
        : null;

    return RefreshIndicator(
      color: AppColors.primary500,
      onRefresh: _loadMeeting,
      child: ListView(
        physics: const AlwaysScrollableScrollPhysics(
          parent: BouncingScrollPhysics(),
        ),
        padding: const EdgeInsets.fromLTRB(20, 24, 20, 120),
        children: [
          gradientHeroCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            meeting.title,
                            style: Theme.of(context)
                                    .textTheme
                                    .headlineSmall
                                    ?.copyWith(
                                      color: AppColors.textPrimary,
                                      fontWeight: FontWeight.w700,
                                    ) ??
                                const TextStyle(
                                  fontSize: 24,
                                  fontWeight: FontWeight.w700,
                                ),
                          ),
                          if (meeting.subjectName?.isNotEmpty ?? false) ...[
                            const SizedBox(height: 6),
                            Text(
                              meeting.subjectName!,
                              style: Theme.of(context)
                                  .textTheme
                                  .bodyMedium
                                  ?.copyWith(color: AppColors.textSecondary),
                            ),
                          ],
                          if (!meeting.isPermanent) ...[
                            const SizedBox(height: 12),
                            Text(
                              '${dateFormat.format(meeting.scheduledAt)} • ${timeFormat.format(meeting.scheduledAt)}',
                              style: Theme.of(context)
                                  .textTheme
                                  .bodyMedium
                                  ?.copyWith(color: AppColors.textSecondary),
                            ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(width: 16),
                    if (!meeting.isPermanent) _buildStatusChip(meeting.status),
                  ],
                ),
                const SizedBox(height: 20),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: heroStats,
                ),
              ],
            ),
          ),
          if (!meeting.isPermanent) ...[
            const SizedBox(height: 20),
            surfaceCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    context,
                    icon: Icons.calendar_today_rounded,
                    title: l10n.dateAndTime,
                  ),
                  const SizedBox(height: 16),
                  _buildDetailRow(
                    context,
                    label: l10n.meetingDate,
                    value: dateFormat.format(meeting.scheduledAt),
                  ),
                  _buildDetailRow(
                    context,
                    label: l10n.meetingTime,
                    value: timeFormat.format(meeting.scheduledAt),
                  ),
                  _buildDetailRow(
                    context,
                    label: l10n.meetingDuration,
                    value: '${meeting.duration} ${l10n.minutes}',
                  ),
                  _buildDetailRow(
                    context,
                    label: l10n.meetingRecurrence,
                    value: recurrenceLabel,
                  ),
                  _buildDetailRow(
                    context,
                    label: l10n.type,
                    value: _getMeetingTypeText(meeting.type),
                  ),
                ],
              ),
            ),
          ],
          if (meeting.allowAnonymous) ...[
            const SizedBox(height: 20),
            surfaceCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    context,
                    icon: Icons.link_rounded,
                    title: l10n.anonymousLinkLabel,
                    subtitle: l10n.allowsAnonymousJoin,
                  ),
                  const SizedBox(height: 12),
                  if (anonymousLink != null)
                    SelectableText(
                      anonymousLink,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppColors.textPrimary,
                            fontWeight: FontWeight.w600,
                          ),
                    )
                  else
                    Text(
                      l10n.loading,
                      style: Theme.of(context)
                          .textTheme
                          .bodyMedium
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                  const SizedBox(height: 12),
                  SizedBox(
                    width: double.infinity,
                    child: OutlinedButton.icon(
                      onPressed: anonymousLink == null ? null : _copyAnonymousLink,
                      icon: const Icon(Icons.copy_rounded),
                      label: Text(l10n.copyLink),
                    ),
                  ),
                ],
              ),
            ),
          ],
          const SizedBox(height: 20),
          surfaceCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildSectionHeader(
                  context,
                  icon: Icons.settings_rounded,
                  title: l10n.settingsTitle,
                ),
                const SizedBox(height: 16),
                Wrap(
                  spacing: 12,
                  runSpacing: 12,
                  children: [
                    // Отображение статуса записи (аудио и видео объединены в один флаг)
                    _buildFeatureChip(
                      icon: Icons.videocam_outlined,
                      label: l10n.videoRecording,
                      isActive: meeting.needsRecord,
                    ),
                    _buildFeatureChip(
                      icon: Icons.mic_none_rounded,
                      label: l10n.audioRecording,
                      isActive: meeting.needsRecord,
                    ),
                    _buildFeatureChip(
                      icon: Icons.description_outlined,
                      label: l10n.transcription,
                      isActive: meeting.needsTranscription,
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _buildDetailRow(
                  context,
                  label: l10n.allowsAnonymousJoin,
                  value: meeting.allowAnonymous ? l10n.enabled : l10n.disabled,
                ),
                _buildDetailRow(
                  context,
                  label: l10n.roomId,
                  value: meeting.liveKitRoomId ?? '—',
                ),
              ],
            ),
          ),
          const SizedBox(height: 20),
          surfaceCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                InkWell(
                  onTap: meeting.participants.isEmpty
                      ? null
                      : () {
                          setState(() {
                            _isParticipantsExpanded = !_isParticipantsExpanded;
                          });
                        },
                  borderRadius: BorderRadius.circular(24),
                  child: Row(
                    children: [
                      Expanded(
                        child: _buildSectionHeader(
                          context,
                          icon: Icons.people_alt_rounded,
                          title:
                              '${l10n.participants} (${meeting.participants.length})',
                          subtitle: participantsSummary,
                        ),
                      ),
                      if (meeting.participants.isNotEmpty)
                        Icon(
                          _isParticipantsExpanded
                              ? Icons.keyboard_arrow_up_rounded
                              : Icons.keyboard_arrow_down_rounded,
                          color: AppColors.textSecondary,
                        ),
                    ],
                  ),
                ),
                const SizedBox(height: 12),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _buildSummaryChip(
                      icon: Icons.wifi_tethering,
                      label: l10n.onlineCount(meeting.activeParticipantsCount),
                    ),
                    if (meeting.anonymousGuestsCount > 0)
                      _buildSummaryChip(
                        icon: Icons.person_outline,
                        label: l10n.anonymousGuestsCount(
                          meeting.anonymousGuestsCount,
                        ),
                      ),
                  ],
                ),
                const SizedBox(height: 12),
                if (meeting.participants.isEmpty)
                  Text(
                    l10n.noParticipants,
                    style: Theme.of(context)
                        .textTheme
                        .bodyMedium
                        ?.copyWith(color: AppColors.textSecondary),
                  )
                else if (_isParticipantsExpanded)
                  Column(
                    children: [
                      for (var i = 0; i < meeting.participants.length; i++) ...[
                        if (i > 0)
                          const Divider(
                            height: 1,
                            color: AppColors.border,
                          ),
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 12),
                          child: _buildParticipantTile(
                            context,
                            meeting.participants[i],
                          ),
                        ),
                      ],
                    ],
                  ),
              ],
            ),
          ),
          if (meeting.departments.isNotEmpty) ...[
            const SizedBox(height: 20),
            surfaceCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    context,
                    icon: Icons.business_outlined,
                    title: l10n.departments,
                  ),
                  const SizedBox(height: 16),
                  Wrap(
                    spacing: 10,
                    runSpacing: 10,
                    children: meeting.departments
                        .map(_buildDepartmentChip)
                        .toList(),
                  ),
                ],
              ),
            ),
          ],
          if (meeting.additionalNotes != null &&
              meeting.additionalNotes!.isNotEmpty) ...[
            const SizedBox(height: 20),
            surfaceCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildSectionHeader(
                    context,
                    icon: Icons.notes_rounded,
                    title: l10n.notes,
                  ),
                  const SizedBox(height: 12),
                  Text(
                    meeting.additionalNotes!,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                ],
              ),
            ),
          ],
          const SizedBox(height: 20),
          surfaceCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildSectionHeader(
                  context,
                  icon: Icons.info_outline_rounded,
                  title: l10n.metadata,
                ),
                const SizedBox(height: 12),
                _buildDetailRow(
                  context,
                  label: l10n.createdBy,
                  value: meeting.createdBy,
                ),
                _buildDetailRow(
                  context,
                  label: l10n.created,
                  value: timestampFormat.format(meeting.createdAt),
                ),
                _buildDetailRow(
                  context,
                  label: l10n.updated,
                  value: timestampFormat.format(meeting.updatedAt),
                ),
              ],
            ),
          ),
          const SizedBox(height: 40),
        ],
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    final l10n = AppLocalizations.of(context)!;
    Color backgroundColor;
    Color textColor;
    String label;

    switch (status.toLowerCase()) {
      case 'active':
      case 'in_progress':
        backgroundColor = const Color(0xFFFEF3C7); // Warning light background
        textColor = const Color(0xFF92400E); // Warning text color from frontend
        label = status.toLowerCase() == 'active' ? l10n.active : l10n.statusInProgress;
        break;
      case 'scheduled':
        backgroundColor = const Color(0xFFDBEAFE); // Primary-50 background
        textColor = const Color(0xFF2563EB); // Primary-700 text color from frontend
        label = l10n.scheduled;
        break;
      case 'completed':
        backgroundColor = const Color(0xFFD1FAE5); // Success light background
        textColor = const Color(0xFF059669); // Success text color from frontend
        label = l10n.completed;
        break;
      case 'cancelled':
        backgroundColor = const Color(0xFFFEE2E2); // Error light background
        textColor = const Color(0xFF991B1B); // Error text color from frontend
        label = l10n.cancelled;
        break;
      default:
        backgroundColor = const Color(0xFFF3F4F6); // Gray light
        textColor = const Color(0xFF6B7280); // Gray text
        label = status;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: textColor.withValues(alpha: 0.3),
          width: 1,
        ),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: textColor,
          fontSize: 12,
          fontWeight: FontWeight.bold,
        ),
      ),
    );
  }

  Widget _buildHeroStat(
    BuildContext context, {
    required IconData icon,
    required String label,
    required String value,
  }) {
    final textTheme = Theme.of(context).textTheme;
    return ConstrainedBox(
      constraints: const BoxConstraints(minWidth: 140),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha: 0.9),
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, size: 20, color: AppColors.primary600),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    label,
                    style: textTheme.labelMedium?.copyWith(
                      color: AppColors.textSecondary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    value,
                    style: textTheme.bodyMedium?.copyWith(
                      color: AppColors.textPrimary,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(
    BuildContext context, {
    required IconData icon,
    required String title,
    String? subtitle,
  }) {
    final textTheme = Theme.of(context).textTheme;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: AppColors.primary50,
            borderRadius: BorderRadius.circular(18),
          ),
          child: Icon(
            icon,
            color: AppColors.primary600,
            size: 22,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                ),
              ),
              if (subtitle != null) ...[
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildDetailRow(
    BuildContext context, {
    required String label,
    required String value,
  }) {
    final textTheme = Theme.of(context).textTheme;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              label,
              style: textTheme.bodySmall?.copyWith(
                color: AppColors.textSecondary,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: textTheme.bodyMedium?.copyWith(
                color: AppColors.textPrimary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFeatureChip({
    required IconData icon,
    required String label,
    required bool isActive,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: isActive ? AppColors.primary50 : AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(
          color: isActive ? AppColors.primary300 : AppColors.border,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 18,
            color: isActive ? AppColors.primary600 : AppColors.textSecondary,
          ),
          const SizedBox(width: 8),
          Text(
            label,
            style: TextStyle(
              color: isActive ? AppColors.primary700 : AppColors.textSecondary,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryChip({
    required IconData icon,
    required String label,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.border),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 16, color: AppColors.textSecondary),
          const SizedBox(width: 6),
          Text(
            label,
            style: const TextStyle(
              color: AppColors.textSecondary,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildParticipantTile(
    BuildContext context,
    MeetingParticipant participant,
  ) {
    final textTheme = Theme.of(context).textTheme;
    final displayName =
        participant.displayName.isNotEmpty ? participant.displayName : participant.userId;
    final initials = displayName.isNotEmpty
        ? displayName.substring(0, 1).toUpperCase()
        : '?';
    final subtitleParts = <String>[];
    if (participant.role.isNotEmpty) {
      subtitleParts.add(participant.role);
    }
    if (participant.status.isNotEmpty) {
      subtitleParts.add(participant.status);
    }

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        CircleAvatar(
          radius: 22,
          backgroundColor: AppColors.primary50,
          child: Text(
            initials,
            style: const TextStyle(
              color: AppColors.primary600,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                displayName,
                style: textTheme.bodyLarge?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w600,
                ),
              ),
              if (subtitleParts.isNotEmpty)
                Text(
                  subtitleParts.join(' • '),
                  style: textTheme.bodySmall?.copyWith(
                    color: AppColors.textSecondary,
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildDepartmentChip(String department) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: AppColors.border),
      ),
      child: Text(
        department,
        style: const TextStyle(
          color: AppColors.textSecondary,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  String _getMeetingTypeText(String type) {
    final l10n = AppLocalizations.of(context)!;
    switch (type.toLowerCase()) {
      case 'conference':
        return l10n.typeConference;
      case 'presentation':
        return l10n.typePresentation;
      case 'training':
        return l10n.typeTraining;
      case 'discussion':
        return l10n.typeDiscussion;
      default:
        return type;
    }
  }

  String _getRecurrenceText(String recurrence) {
    final l10n = AppLocalizations.of(context)!;
    switch (recurrence.toLowerCase()) {
      case 'daily':
        return l10n.recurrenceDaily;
      case 'weekly':
        return l10n.recurrenceWeekly;
      case 'monthly':
        return l10n.recurrenceMonthly;
      case 'permanent':
        return l10n.recurrencePermanent;
      case 'none':
        return l10n.recurrenceNone;
      default:
        return recurrence;
    }
  }

  // Helper widget to create Aurora-styled cards
  Widget _buildCard({
    required Widget child,
    EdgeInsets? padding,
    EdgeInsets? margin,
  }) {
    return Container(
      margin: margin ?? const EdgeInsets.only(bottom: 20),
      padding: padding ?? const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24), // radius-xl
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            blurRadius: 30,
            offset: const Offset(0, 12),
          ),
        ],
        border: Border.all(color: AppColors.border),
      ),
      child: child,
    );
  }

  Widget _buildTasksTab() {
    if (_meeting == null) {
      return const Center(child: Text('No meeting data'));
    }

    return FutureBuilder<List<Task>>(
      future: _loadMeetingTasks(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 64, color: Colors.red),
                const SizedBox(height: 16),
                Text('Error: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () {
                    setState(() {}); // Trigger rebuild to retry
                  },
                  child: const Text('Retry'),
                ),
              ],
            ),
          );
        }

        final tasks = snapshot.data ?? [];

        if (tasks.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.task_alt, size: 64, color: Colors.grey[400]),
                const SizedBox(height: 16),
                Text(
                  'No tasks yet',
                  style: TextStyle(fontSize: 18, color: Colors.grey[600]),
                ),
                const SizedBox(height: 8),
                Text(
                  'Tasks will appear here after the meeting',
                  style: TextStyle(fontSize: 14, color: Colors.grey[500]),
                ),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () async {
            setState(() {}); // Trigger rebuild to reload tasks
          },
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: tasks.length,
            itemBuilder: (context, index) {
              final task = tasks[index];
              return TaskCard(
                task: task,
                onTap: () => _showTaskDetails(task),
              );
            },
          ),
        );
      },
    );
  }

  Future<List<Task>> _loadMeetingTasks() async {
    if (_meeting == null) return [];

    try {
      final apiUrl = await _configService.getApiUrl();
      final authService = AuthService(_apiClient);
      final taskService = TaskService(
        baseUrl: apiUrl.replaceAll('/api/v1', ''),
        authService: authService,
      );

      return await taskService.getMeetingTasks(_meeting!.id);
    } catch (e) {
      throw Exception('Failed to load tasks: $e');
    }
  }

  void _showTaskDetails(Task task) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.7,
        minChildSize: 0.5,
        maxChildSize: 0.95,
        expand: false,
        builder: (context, scrollController) => _buildTaskDetailSheet(
          task,
          scrollController,
        ),
      ),
    );
  }

  Widget _buildTaskDetailSheet(Task task, ScrollController scrollController) {
    return Container(
      padding: const EdgeInsets.all(24),
      child: ListView(
        controller: scrollController,
        children: [
          // Handle bar
          Center(
            child: Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: Colors.grey[300],
                borderRadius: BorderRadius.circular(2),
              ),
            ),
          ),
          const SizedBox(height: 20),

          // Title
          Text(
            task.title,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 16),

          // Description
          if (task.description != null && task.description!.isNotEmpty) ...[
            Text(
              'Description',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Text(task.description!),
            const SizedBox(height: 16),
          ],

          // Hint
          if (task.hint != null && task.hint!.isNotEmpty) ...[
            Text(
              'Hint',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.blue.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.blue.withOpacity(0.3)),
              ),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(Icons.lightbulb_outline, color: Colors.blue[700]),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      task.hint!,
                      style: TextStyle(color: Colors.blue[700]),
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 16),
          ],

          // Assigned to
          if (task.assignedToUser != null) ...[
            Text(
              'Assigned to',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                CircleAvatar(
                  backgroundColor: AppColors.primary100,
                  child: Text(
                    _getUserInitials(task.assignedToUser!),
                    style: const TextStyle(color: AppColors.primary600),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _getUserDisplayName(task.assignedToUser!),
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                      if (task.assignedToUser!.email != null)
                        Text(
                          task.assignedToUser!.email!,
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.grey[600],
                          ),
                        ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
          ],

          // AI Source (if present)
          if (task.extractedByAi && task.sourceSegment != null) ...[
            Text(
              'AI Extracted From',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.purple.withOpacity(0.05),
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.purple.withOpacity(0.2)),
              ),
              child: Text(
                task.sourceSegment!,
                style: TextStyle(
                  color: Colors.purple[900],
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
          ],
        ],
      ),
    );
  }

  String _getUserDisplayName(dynamic user) {
    if (user.firstName != null && user.firstName!.isNotEmpty) {
      return '${user.firstName} ${user.lastName ?? ''}'.trim();
    }
    return user.username ?? user.email ?? 'Unknown';
  }

  String _getUserInitials(dynamic user) {
    if (user.firstName != null && user.firstName!.isNotEmpty) {
      final first = user.firstName![0].toUpperCase();
      final last = user.lastName != null && user.lastName!.isNotEmpty
          ? user.lastName![0].toUpperCase()
          : '';
      return '$first$last';
    }
    if (user.username != null && user.username!.isNotEmpty) {
      return user.username![0].toUpperCase();
    }
    return '?';
  }

  // Вспомогательные методы для совместимости (используют lowerCamelCase согласно Dart naming conventions)
  Widget surfaceCard({
    required Widget child,
    EdgeInsets? padding,
    EdgeInsets? margin,
  }) {
    return _buildCard(child: child, padding: padding, margin: margin);
  }

  Widget gradientHeroCard({required Widget child}) {
    return _buildCard(child: child);
  }
}
