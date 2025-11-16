import 'package:flutter/material.dart';
import '../l10n/app_localizations.dart';
import 'package:intl/intl.dart';
import '../models/meeting.dart';
import '../models/recording.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/config_service.dart';
import '../widgets/error_display.dart';
import '../widgets/app_card.dart';
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
  late TabController _tabController;

  MeetingWithDetails? _meeting;
  bool _isLoading = true;
  String? _error;
  bool _isParticipantsExpanded = false;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _initService();
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  Future<void> _initService() async {
    final apiUrl = await _configService.getApiUrl();
    final apiClient = ApiClient(baseUrl: apiUrl);
    _meetingsService = MeetingsService(apiClient);
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
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: AppColors.primary500,
          labelColor: AppColors.primary600,
          unselectedLabelColor: AppColors.textSecondary,
          tabs: [
            Tab(text: l10n.tabInfo),
            Tab(text: l10n.tabRecordings),
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
            child: SurfaceCard(
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
            child: SurfaceCard(
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

    return SurfaceCard(
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
                _RecordingStatusChip(
                  status: recording.status,
                  color: _getRecordingStatusColor(recording.status),
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
            if (recording.playlistUrl != null) ...[
              const SizedBox(height: 16),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => RecordingPlayerScreen(
                          recording: recording,
                        ),
                      ),
                    );
                  },
                  icon: const Icon(
                    Icons.play_circle,
                    size: 20,
                  ),
                  label: Text(l10n.playRecording),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF26C6DA),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ),
            ],
          ],
        ),
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

  Widget _buildMeetingContent() {
    final l10n = AppLocalizations.of(context)!;
    final meeting = _meeting!;
    // Локализованные форматы даты и времени
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat.yMMMd(locale);
    final timeFormat = DateFormat.Hm(locale);

                Text(
                  meeting.title,
                  style: const TextStyle(
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
                if (meeting.subjectName != null) ...[
                  const SizedBox(height: 8),
                  Text(
                    meeting.subjectName!,
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.white.withValues(alpha: 0.9),
                    ),
                  ),
                ],
              ],
            ),
          ),

          // Meeting info
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildInfoSection(
                  icon: Icons.calendar_today,
                  title: l10n.dateAndTime,
                  content: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        dateFormat.format(meeting.scheduledAt),
                        style: const TextStyle(fontSize: 16),
                      ),
                      Text(
                        '${timeFormat.format(meeting.scheduledAt)} (${meeting.duration} ${l10n.minutes})',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.grey[600],
                        ),
                      ),
                      // Показываем recurrence только если он не null и не "none"
                      if (meeting.recurrence != null && meeting.recurrence!.toLowerCase() != 'none') ...[
                        const SizedBox(height: 4),
                        Text(
                          '${l10n.meetingRecurrence}: ${_getRecurrenceText(meeting.recurrence!)}',
                          style: TextStyle(
                            fontSize: 14,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                const Divider(height: 32),

                // Участники - сворачиваемые
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    InkWell(
                      onTap: () {
                        setState(() {
                          _isParticipantsExpanded = !_isParticipantsExpanded;
                        });
                      },
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        child: Row(
                          children: [
                            Icon(
                              Icons.people,
                              size: 20,
                              color: const Color(0xFF26C6DA),
                            ),
                            const SizedBox(width: 8),
                            Text(
                              '${l10n.participants} (${meeting.participants.length})',
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Spacer(),
                            Icon(
                              _isParticipantsExpanded
                                ? Icons.keyboard_arrow_up
                                : Icons.keyboard_arrow_down,
                              color: Colors.grey[600],
                            ),
                          ],
                        ),
                      ),
                    ),
                    if (_isParticipantsExpanded) ...[
                      const SizedBox(height: 12),
                      Padding(
                        padding: const EdgeInsets.only(left: 28),
                        child: meeting.participants.isEmpty
                            ? Text(
                                l10n.noParticipants,
                                style: TextStyle(color: Colors.grey[600]),
                              )
                            : Column(
                                children: meeting.participants.map((participant) {
                                  return Padding(
                                    padding: const EdgeInsets.symmetric(vertical: 4),
                                    child: Row(
                                      children: [
                                        CircleAvatar(
                                          radius: 20,
                                          backgroundColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
                                          child: Text(
                                            participant.displayName.substring(0, 1).toUpperCase(),
                                            style: const TextStyle(
                                              color: Color(0xFF00ACC1),
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
                                                participant.displayName,
                                                style: const TextStyle(
                                                  fontSize: 14,
                                                  fontWeight: FontWeight.w500,
                                                ),
                                              ),
                                              Text(
                                                '${participant.role} • ${participant.status}',
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
                                  );
                                }).toList(),
                              ),
                      ),
                    ],
                  ],
                ),

                if (meeting.departments.isNotEmpty) ...[
                  const Divider(height: 32),
                  _buildInfoSection(
                    icon: Icons.business,
                    title: l10n.departments,
                    content: Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: meeting.departments.map((dept) {
                        return Chip(
                          label: Text(dept),
                          backgroundColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
                          side: BorderSide(
                            color: const Color(0xFF26C6DA).withValues(alpha: 0.3),
                            width: 1,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                          labelStyle: const TextStyle(
                            color: Color(0xFF00ACC1),
                            fontWeight: FontWeight.w500,
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ],

                const Divider(height: 32),

                _buildInfoSection(
                  icon: Icons.settings,
                  title: l10n.settingsTitle,
                  content: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSettingRow(
                        l10n.type,
                        _getMeetingTypeText(meeting.type),
                      ),
                      _buildSettingRow(
                        l10n.videoRecording,
                        meeting.needsVideoRecord ? l10n.enabled : l10n.disabled,
                      ),
                      _buildSettingRow(
                        l10n.audioRecording,
                        meeting.needsAudioRecord ? l10n.enabled : l10n.disabled,
                      ),
                    ],
                  ),
                ),

                if (meeting.additionalNotes != null && meeting.additionalNotes!.isNotEmpty) ...[
                  const Divider(height: 32),
                  _buildInfoSection(
                    icon: Icons.notes,
                    title: l10n.notes,
                    content: Text(
                      meeting.additionalNotes!,
                      style: const TextStyle(fontSize: 14),
                    ),
                  ),
                ],

                const SizedBox(height: 80), // Space for FAB
              ],
            ),
          ),
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

  Widget _buildInfoSection({
    required IconData icon,
    required String title,
    required Widget content,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(
              icon,
              size: 20,
              color: const Color(0xFF26C6DA),
            ),
            const SizedBox(width: 8),
            Text(
              title,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Padding(
          padding: const EdgeInsets.only(left: 28),
          child: content,
        ),
      ],
    );
  }

  Widget _buildSettingRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 14,
                color: Colors.grey[600],
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
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
}
