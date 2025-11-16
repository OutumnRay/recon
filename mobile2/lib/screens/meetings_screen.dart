import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/config_service.dart';
import '../models/meeting.dart';
import '../widgets/error_display.dart';
import 'meeting_detail_screen.dart';
import 'create_meeting_screen.dart';
import '../l10n/app_localizations.dart';
import '../theme/app_colors.dart';
import '../widgets/app_card.dart';

class MeetingsScreen extends StatefulWidget {
  final ApiClient apiClient;

  const MeetingsScreen({super.key, required this.apiClient});

  @override
  State<MeetingsScreen> createState() => _MeetingsScreenState();
}

class _MeetingsScreenState extends State<MeetingsScreen>
    with AutomaticKeepAliveClientMixin {
  late final MeetingsService _meetingsService;

  List<MeetingWithDetails>? _meetings;
  bool _isLoading = true;
  String? _error;
  String _filter = 'scheduled';
  String? _copiedMeetingId;

  @override
  bool get wantKeepAlive => true;

  @override
  void initState() {
    super.initState();
    _meetingsService = MeetingsService(widget.apiClient);
    _loadMeetings();
  }

  Future<void> _loadMeetings() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final meetings = await _meetingsService.getMeetings(
        status: _filter == 'all' ? null : _filter,
        pageSize: 20,
      );
      if (mounted) {
        setState(() {
          _meetings = meetings;
          _isLoading = false;
        });
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = 'API Error: ${e.message}\nStatus Code: ${e.statusCode}';
          _isLoading = false;
        });
      }
    } on Exception catch (e) {
      if (mounted) {
        setState(() {
          _error = 'Connection Error:\n${e.toString()}';
          _isLoading = false;
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error =
              'Unexpected Error:\n${e.toString()}\n${stackTrace.toString().split('\n').take(3).join('\n')}';
          _isLoading = false;
        });
      }
    }
  }

  String _formatDateTime(DateTime dateTime, BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final meetingDate = DateTime(dateTime.year, dateTime.month, dateTime.day);

    if (meetingDate == today) {
      return '${l10n.today}, ${DateFormat.Hm().format(dateTime)}';
    } else if (meetingDate == today.add(const Duration(days: 1))) {
      return '${l10n.tomorrow}, ${DateFormat.Hm().format(dateTime)}';
    } else if (meetingDate.isAfter(today) &&
        meetingDate.isBefore(today.add(const Duration(days: 7)))) {
      return '${DateFormat.E().format(dateTime)}, ${DateFormat.Hm().format(dateTime)}';
    } else {
      return DateFormat('dd.MM.yyyy HH:mm').format(dateTime);
    }
  }

  IconData _getMeetingIcon(String type) {
    switch (type) {
      case 'conference':
        return Icons.people_outline_rounded;
      case 'presentation':
        return Icons.present_to_all;
      case 'training':
        return Icons.school_outlined;
      default:
        return Icons.event_note_outlined;
    }
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'scheduled':
        return AppColors.primary600;
      case 'in_progress':
        return AppColors.warning;
      case 'completed':
        return AppColors.success;
      case 'cancelled':
        return AppColors.danger;
      default:
        return AppColors.textTertiary;
    }
  }

  String _getStatusText(String status, BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    switch (status) {
      case 'scheduled':
        return l10n.statusScheduled;
      case 'in_progress':
        return l10n.statusInProgress;
      case 'completed':
        return l10n.statusCompleted;
      case 'cancelled':
        return l10n.statusCancelled;
      default:
        return status;
    }
  }

  Future<void> _copyAnonymousLink(String meetingId) async {
    final l10n = AppLocalizations.of(context)!;

    try {
      final configService = ConfigService();
      final apiUrl = await configService.getApiUrl();
      final baseUrl = apiUrl.replaceAll('/api/v1', '');
      final anonymousLink = '$baseUrl/meeting/$meetingId/join';

      await Clipboard.setData(ClipboardData(text: anonymousLink));

      if (mounted) {
        setState(() => _copiedMeetingId = meetingId);
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) setState(() => _copiedMeetingId = null);
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.anonymousLinkCopied),
            backgroundColor: AppColors.success,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to copy link: $e'),
            backgroundColor: AppColors.danger,
          ),
        );
      }
    }
  }

  Future<void> _openMeetingDetails(MeetingWithDetails meeting) async {
    await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => MeetingDetailScreen(meetingId: meeting.id),
      ),
    );
    _loadMeetings();
  }

  Future<void> _openCreateMeeting() async {
    final result = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => CreateMeetingScreen(apiClient: widget.apiClient),
      ),
    );
    if (result == true) {
      _loadMeetings();
    }
  }

  FloatingActionButton _buildFab(AppLocalizations l10n) {
    return FloatingActionButton.extended(
      heroTag: 'meetings_fab',
      onPressed: _openCreateMeeting,
      backgroundColor: AppColors.primary500,
      foregroundColor: Colors.white,
      icon: const Icon(Icons.add_rounded),
      label: Text(l10n.newMeeting),
    );
  }

  Widget _buildHeroSection(AppLocalizations l10n) {
    final total = _meetings?.length ?? 0;
    final inProgress =
        _meetings?.where((m) => m.status == 'in_progress').length ?? 0;
    final scheduled =
        _meetings?.where((m) => m.status == 'scheduled').length ?? 0;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        GradientHeroCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.meetings,
                style: Theme.of(context).textTheme.headlineMedium,
              ),
              const SizedBox(height: 8),
              Text(
                l10n.loginSubtitle,
                style: Theme.of(context)
                    .textTheme
                    .bodyMedium
                    ?.copyWith(color: AppColors.textSecondary),
              ),
              const SizedBox(height: 20),
              Row(
                children: [
                  Expanded(
                    child: _StatTile(
                      label: l10n.filterAll,
                      value: '$total',
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _StatTile(
                      label: l10n.filterInProgress,
                      value: '$inProgress',
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: _StatTile(
                      label: l10n.filterScheduled,
                      value: '$scheduled',
                    ),
                  ),
                ],
              )
            ],
          ),
        ),
        const SizedBox(height: 16),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            _buildFilterChip(l10n.filterAll, 'all'),
            _buildFilterChip(l10n.filterScheduled, 'scheduled'),
            _buildFilterChip(l10n.filterInProgress, 'in_progress'),
            _buildFilterChip(l10n.filterCompleted, 'completed'),
          ],
        )
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    super.build(context);
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.meetings),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadMeetings,
            tooltip: l10n.refresh,
          ),
        ],
      ),
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: _loadMeetings,
          child: ListView(
            padding: const EdgeInsets.fromLTRB(20, 24, 20, 120),
            children: [
              _buildHeroSection(l10n),
              const SizedBox(height: 24),
              if (_isLoading)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 60),
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (_error != null)
                FullScreenError(error: _error!, onRetry: _loadMeetings)
              else if (_meetings == null || _meetings!.isEmpty)
                _EmptyState(onCreateMeeting: _openCreateMeeting, l10n: l10n)
              else
                ..._meetings!.map((m) => _buildMeetingCard(m, context))
            ],
          ),
        ),
      ),
      floatingActionButton: _buildFab(l10n),
    );
  }

  Widget _buildMeetingCard(MeetingWithDetails meeting, BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isNow = meeting.status == 'in_progress';
    final canJoin = meeting.status == 'scheduled' || isNow;
    final hasRecording = meeting.needsVideoRecord || meeting.needsAudioRecord;

    return GestureDetector(
      onTap: () => _openMeetingDetails(meeting),
      child: SurfaceCard(
        margin: const EdgeInsets.only(bottom: 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: AppColors.primary50,
                    borderRadius: BorderRadius.circular(18),
                  ),
                  child: Icon(
                    _getMeetingIcon(meeting.type),
                    color: AppColors.primary600,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              meeting.title,
                              style: Theme.of(context)
                                  .textTheme
                                  .titleLarge
                                  ?.copyWith(fontWeight: FontWeight.w700),
                            ),
                          ),
                          Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 12, vertical: 6),
                            decoration: BoxDecoration(
                              color: _getStatusColor(meeting.status)
                                  .withValues(alpha: 0.12),
                              borderRadius: BorderRadius.circular(16),
                            ),
                            child: Text(
                              _getStatusText(meeting.status, context),
                              style: TextStyle(
                                fontSize: 12,
                                fontWeight: FontWeight.w600,
                                color: _getStatusColor(meeting.status),
                              ),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      Wrap(
                        spacing: 8,
                        runSpacing: 6,
                        children: [
                          if (meeting.subjectName != null)
                            _Pill(text: meeting.subjectName!),
                          if (meeting.isPermanent)
                            _Pill(
                              text: l10n.permanent,
                              icon: Icons.all_inclusive,
                            ),
                          if (meeting.allowAnonymous)
                            _Pill(
                              text: l10n.allowsAnonymousJoin,
                              icon: Icons.visibility_off_outlined,
                            ),
                          if (meeting.needsVideoRecord)
                            _Pill(text: l10n.meetingVideoRecord, icon: Icons.videocam),
                          if (meeting.needsAudioRecord)
                            _Pill(text: l10n.meetingAudioRecord, icon: Icons.mic_none),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),
            Row(
              children: [
                Icon(Icons.calendar_today, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 6),
                Text(
                  _formatDateTime(meeting.scheduledAt, context),
                  style: Theme.of(context)
                      .textTheme
                      .bodyMedium
                      ?.copyWith(color: AppColors.textSecondary),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.schedule, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 6),
                Text('${meeting.duration} ${l10n.minutes}'),
                const SizedBox(width: 12),
                Icon(Icons.people_outline, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 6),
                Text('${meeting.participants.length} ${l10n.meetingParticipants}')
              ],
            ),
            const SizedBox(height: 16),
            Wrap(
              spacing: 12,
              runSpacing: 8,
              alignment: WrapAlignment.spaceBetween,
              children: [
                OutlinedButton.icon(
                  onPressed: () => _openMeetingDetails(meeting),
                  icon: const Icon(Icons.info_outline),
                  label: Text(l10n.meetingDetailsTitle),
                ),
                if (meeting.allowAnonymous)
                  OutlinedButton.icon(
                    onPressed: () => _copyAnonymousLink(meeting.id),
                    icon: Icon(
                      _copiedMeetingId == meeting.id
                          ? Icons.check
                          : Icons.link_outlined,
                    ),
                    label: Text(
                      _copiedMeetingId == meeting.id
                          ? l10n.linkCopied
                          : l10n.copyLink,
                    ),
                  ),
                if (canJoin)
                  ElevatedButton.icon(
                    onPressed: () => _openMeetingDetails(meeting),
                    icon: const Icon(Icons.play_arrow_rounded),
                    label: Text(l10n.joinMeeting),
                  ),
                if (hasRecording)
                  TextButton.icon(
                    onPressed: () => _openMeetingDetails(meeting),
                    icon: const Icon(Icons.video_library_outlined),
                    label: Text(l10n.recordingOptions),
                  ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterChip(String label, String value) {
    final isSelected = _filter == value;
    return ChoiceChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          setState(() => _filter = value);
          _loadMeetings();
        }
      },
      labelStyle: TextStyle(
        color: isSelected ? AppColors.primary600 : AppColors.textSecondary,
        fontWeight: FontWeight.w600,
      ),
      side: BorderSide(
        color: isSelected ? AppColors.primary200 : AppColors.border,
      ),
      selectedColor: AppColors.primary50,
      backgroundColor: Colors.white,
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
    );
  }
}

class _StatTile extends StatelessWidget {
  final String label;
  final String value;

  const _StatTile({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return SurfaceCard(
      padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 6),
          Text(
            value,
            style: Theme.of(context)
                .textTheme
                .titleLarge
                ?.copyWith(fontWeight: FontWeight.bold),
          ),
        ],
      ),
    );
  }
}

class _Pill extends StatelessWidget {
  final String text;
  final IconData? icon;

  const _Pill({required this.text, this.icon});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[
            Icon(icon, size: 14, color: AppColors.textSecondary),
            const SizedBox(width: 4),
          ],
          Text(
            text,
            style: Theme.of(context)
                .textTheme
                .bodySmall
                ?.copyWith(fontWeight: FontWeight.w600),
          ),
        ],
      ),
    );
  }
}

class _EmptyState extends StatelessWidget {
  final VoidCallback onCreateMeeting;
  final AppLocalizations l10n;

  const _EmptyState({required this.onCreateMeeting, required this.l10n});

  @override
  Widget build(BuildContext context) {
    return SurfaceCard(
      padding: const EdgeInsets.all(32),
      child: Column(
        children: [
          const Icon(Icons.event_busy, size: 48, color: AppColors.textTertiary),
          const SizedBox(height: 12),
          Text(
            l10n.noMeetingsFound,
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Text(
            l10n.createFirstMeeting,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 16),
          ElevatedButton.icon(
            onPressed: onCreateMeeting,
            icon: const Icon(Icons.add_rounded),
            label: Text(l10n.createMeeting),
          )
        ],
      ),
    );
  }
}
