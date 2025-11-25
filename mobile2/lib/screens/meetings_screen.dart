import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
// Удалён неиспользуемый import intl - форматирование дат выполняется через date_utils.dart
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/config_service.dart';
import '../services/storage_service.dart';
import '../models/meeting.dart';
import '../widgets/error_display.dart';
import 'meeting_detail_screen.dart';
import 'create_meeting_screen.dart';
import '../l10n/app_localizations.dart';
import '../theme/app_colors.dart';
import '../utils/date_utils.dart';

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
  List<MeetingWithDetails>? _filteredMeetings;
  bool _isLoading = true;
  String? _error;
  String _filter = 'all';
  String? _copiedMeetingId;
  final TextEditingController _searchController = TextEditingController();
  String _searchQuery = '';
  String? _currentUserId;
  String? _currentUserRole;

  @override
  bool get wantKeepAlive => true;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    _meetingsService = MeetingsService(widget.apiClient);
    _loadCurrentUser();
    _loadMeetings();
  }

  Future<void> _loadCurrentUser() async {
    try {
      final storageService = StorageService();
      final userData = await storageService.getUserData();
      setState(() {
        _currentUserId = userData['userId'];
        _currentUserRole = userData['role'];
      });
    } catch (e) {
      // Ignore error, user data will remain null
    }
  }

  bool _canCancelMeeting(MeetingWithDetails meeting) {
    if (meeting.status == 'cancelled') return false;

    // Admin can cancel any meeting
    if (_currentUserRole == 'admin') return true;

    // Organization admin can cancel any meeting in their organization
    if (_currentUserRole == 'organization_admin') return true;

    // Meeting creator can cancel their own meeting
    if (_currentUserId == meeting.createdBy) return true;

    return false;
  }

  Future<void> _cancelMeetingFromList(MeetingWithDetails meeting) async {
    final l10n = AppLocalizations.of(context)!;

    // Show confirmation dialog
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(l10n.cancelMeetingTitle),
        content: Text(l10n.cancelMeetingConfirm),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            style: FilledButton.styleFrom(
              backgroundColor: AppColors.danger,
            ),
            child: Text(l10n.confirm),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      await _meetingsService.cancelMeeting(meeting.id);

      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(l10n.meetingCancelledSuccess),
          backgroundColor: AppColors.success,
        ),
      );

      // Reload meetings
      _loadMeetings();
    } on ApiException catch (e) {
      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${l10n.failedToCancelMeeting}: ${e.message}'),
          backgroundColor: AppColors.danger,
          duration: const Duration(seconds: 5),
        ),
      );
    } catch (e) {
      if (!mounted) return;

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${l10n.error}: $e'),
          backgroundColor: AppColors.danger,
          duration: const Duration(seconds: 5),
        ),
      );
    }
  }

  Future<void> _loadMeetings() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      // Map filter to backend status parameter
      String? statusParam;
      switch (_filter) {
        case 'all':
          // Backend will exclude cancelled by default
          statusParam = null;
          break;
        case 'scheduled':
        case 'in_progress':
        case 'completed':
        case 'cancelled':
          statusParam = _filter;
          break;
        case 'permanent':
          // For permanent, we'll get all and filter client-side
          statusParam = null;
          break;
      }

      final meetings = await _meetingsService.getMeetings(
        status: statusParam,
        pageSize: 100,
      );

      if (mounted) {
        setState(() {
          _meetings = meetings;
          _applyFilters();
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

  void _applyFilters() {
    if (_meetings == null) {
      _filteredMeetings = null;
      return;
    }

    var filtered = _meetings!;

    // Apply additional client-side filtering for special cases
    switch (_filter) {
      case 'in_progress':
        // Filter to only show meetings with active participants
        filtered = filtered.where((m) => m.activeParticipantsCount > 0).toList();
        break;
      case 'permanent':
        // Filter to only permanent meetings (not cancelled)
        filtered = filtered.where((m) =>
          (m.isPermanent || m.recurrence == 'permanent') &&
          m.status != 'cancelled'
        ).toList();
        break;
      case 'completed':
        // Filter to only non-permanent completed meetings
        filtered = filtered.where((m) =>
          m.status == 'completed' &&
          !m.isPermanent &&
          m.recurrence != 'permanent'
        ).toList();
        break;
    }

    // Apply search filter
    if (_searchQuery.isNotEmpty) {
      final query = _searchQuery.toLowerCase();
      filtered = filtered.where((meeting) {
        return meeting.title.toLowerCase().contains(query) ||
            (meeting.subjectName?.toLowerCase().contains(query) ?? false) ||
            (meeting.additionalNotes?.toLowerCase().contains(query) ?? false);
      }).toList();
    }

    _filteredMeetings = filtered;
  }

  void _onSearchChanged(String query) {
    setState(() {
      _searchQuery = query;
      _applyFilters();
    });
  }

  String _formatDateTime(DateTime dateTime) {
    final locale = Localizations.localeOf(context).toString();
    final l10n = AppLocalizations.of(context)!;
    return AppDateUtils.formatRelativeDateTime(
      dateTime,
      locale: locale,
      todayLabel: l10n.today,
      tomorrowLabel: l10n.tomorrow,
    );
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

  bool _isMeetingNow(MeetingWithDetails meeting) {
    if (meeting.status != 'in_progress') return false;
    final now = DateTime.now();
    final scheduled = meeting.scheduledAt;
    final endTime = scheduled.add(Duration(minutes: meeting.duration));
    return now.isAfter(scheduled) && now.isBefore(endTime);
  }

  bool _isMeetingPast(MeetingWithDetails meeting) {
    final now = DateTime.now();
    final endTime = meeting.scheduledAt.add(Duration(minutes: meeting.duration));
    return now.isAfter(endTime);
  }

  Future<void> _copyAnonymousLink(String meetingId) async {
    final l10n = AppLocalizations.of(context)!;
    try {
      final config = ConfigService();
      final apiUrl = await config.getApiUrl();
      final baseUrl = apiUrl.replaceAll(RegExp(r'/api/v\d+$'), '');
      final link = '$baseUrl/join/$meetingId';

      await Clipboard.setData(ClipboardData(text: link));

      if (mounted) {
        setState(() => _copiedMeetingId = meetingId);
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) setState(() => _copiedMeetingId = null);
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.anonymousLinkCopied),
            backgroundColor: AppColors.success,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to copy link: $e'),
            backgroundColor: AppColors.danger,
            behavior: SnackBarBehavior.floating,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
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

  @override
  Widget build(BuildContext context) {
    super.build(context);
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.meetings),
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded),
            onPressed: _loadMeetings,
            tooltip: l10n.refresh,
          ),
        ],
      ),
      body: SafeArea(
        child: RefreshIndicator(
          onRefresh: _loadMeetings,
          child: CustomScrollView(
            slivers: [
              // Search bar
              SliverToBoxAdapter(
                child: Container(
                  padding: const EdgeInsets.fromLTRB(20, 16, 20, 8),
                  child: TextField(
                    controller: _searchController,
                    onChanged: _onSearchChanged,
                    decoration: InputDecoration(
                      hintText: l10n.searchMeetings,
                      prefixIcon: const Icon(Icons.search, color: AppColors.textSecondary),
                      suffixIcon: _searchQuery.isNotEmpty
                          ? IconButton(
                              icon: const Icon(Icons.clear, color: AppColors.textSecondary),
                              onPressed: () {
                                _searchController.clear();
                                _onSearchChanged('');
                              },
                            )
                          : null,
                      filled: true,
                      fillColor: AppColors.surfaceMuted.withValues(alpha: 0.5),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(24),
                        borderSide: BorderSide.none,
                      ),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                    ),
                  ),
                ),
              ),

              // Status filter pills
              SliverToBoxAdapter(
                child: Container(
                  padding: const EdgeInsets.fromLTRB(20, 8, 20, 24),
                  child: SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: [
                        _buildFilterPill(l10n.filterAll, 'all'),
                        const SizedBox(width: 8),
                        _buildFilterPill(l10n.filterScheduled, 'scheduled'),
                        const SizedBox(width: 8),
                        _buildFilterPill(l10n.filterInProgress, 'in_progress'),
                        const SizedBox(width: 8),
                        _buildFilterPill(l10n.filterPermanent, 'permanent'),
                        const SizedBox(width: 8),
                        _buildFilterPill(l10n.filterCompleted, 'completed'),
                        const SizedBox(width: 8),
                        _buildFilterPill(l10n.filterCancelled, 'cancelled'),
                      ],
                    ),
                  ),
                ),
              ),

              // Loading/Error/Empty state
              if (_isLoading)
                const SliverFillRemaining(
                  child: Center(child: CircularProgressIndicator()),
                )
              else if (_error != null)
                SliverFillRemaining(
                  child: FullScreenError(error: _error!, onRetry: _loadMeetings),
                )
              else if (_filteredMeetings == null || _filteredMeetings!.isEmpty)
                SliverFillRemaining(
                  child: _EmptyState(onCreateMeeting: _openCreateMeeting, l10n: l10n),
                )
              else
                // Meetings list
                SliverPadding(
                  padding: const EdgeInsets.fromLTRB(20, 8, 20, 100),
                  sliver: SliverList(
                    delegate: SliverChildBuilderDelegate(
                      (context, index) {
                        final meeting = _filteredMeetings![index];
                        return _buildMeetingCard(meeting, context);
                      },
                      childCount: _filteredMeetings!.length,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        heroTag: 'meetings_fab',
        onPressed: _openCreateMeeting,
        backgroundColor: AppColors.primary500,
        foregroundColor: Colors.white,
        icon: const Icon(Icons.add_rounded),
        label: Text(l10n.newMeeting),
        elevation: 4,
      ),
    );
  }

  // Aurora Design: Horizontal meeting card matching web design
  Widget _buildMeetingCard(MeetingWithDetails meeting, BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final isNow = _isMeetingNow(meeting);
    final isPast = _isMeetingPast(meeting);
    final canJoin = meeting.isPermanent || meeting.status == 'scheduled' || meeting.status == 'in_progress';

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(36), // Web: radius-2xl
        border: Border.all(
          color: isNow ? AppColors.primary400 : AppColors.border,
          width: 1,
        ),
        boxShadow: [
          BoxShadow(
            color: isNow
                ? AppColors.primary500.withValues(alpha: 0.12)
                : Colors.black.withValues(alpha: 0.04),
            blurRadius: isNow ? 20 : 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () => _openMeetingDetails(meeting),
          borderRadius: BorderRadius.circular(36),
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Title row with chips and status
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            meeting.title,
                            style: const TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary,
                              height: 1.3,
                            ),
                          ),
                          const SizedBox(height: 8),
                          // Chips row
                          Wrap(
                            spacing: 8,
                            runSpacing: 6,
                            children: [
                              if (meeting.subjectName != null)
                                _buildChip(
                                  meeting.subjectName!,
                                  Icons.bookmark_outline,
                                ),
                              if (meeting.isPermanent)
                                _buildChip(
                                  l10n.permanent,
                                  Icons.all_inclusive,
                                ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    // Cancel button for admin/creator
                    if (_canCancelMeeting(meeting))
                      IconButton(
                        icon: const Icon(Icons.delete_outline, color: AppColors.danger),
                        onPressed: () => _cancelMeetingFromList(meeting),
                        tooltip: l10n.cancelMeetingTitle,
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                  ],
                ),

                const SizedBox(height: 16),

                // Metadata row
                Wrap(
                  spacing: 16,
                  runSpacing: 8,
                  children: [
                    if (!meeting.isPermanent)
                      _buildMetadata(
                        Icons.calendar_today,
                        _formatDateTime(meeting.scheduledAt),
                      ),
                    _buildMetadata(
                      Icons.schedule,
                      '${meeting.duration} ${l10n.minutes}',
                    ),
                    _buildMetadata(
                      Icons.people_outline,
                      '${meeting.participants.length}',
                    ),
                    if (!meeting.isPermanent)
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                        decoration: BoxDecoration(
                          color: _getStatusColor(meeting.status).withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(20),
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

                const SizedBox(height: 16),

                // Action buttons row
                Row(
                  children: [
                    if (meeting.allowAnonymous)
                      _buildActionButton(
                        icon: _copiedMeetingId == meeting.id ? Icons.check : Icons.link,
                        onPressed: () => _copyAnonymousLink(meeting.id),
                      ),
                    if (meeting.allowAnonymous) const SizedBox(width: 8),

                    // View recordings/sessions button with arrow - show for completed/past meetings
                    // This navigates to meeting details where user can view video, audio, transcripts
                    // On phones: show icon-only button to save space
                    // On tablets: show full button with label
                    if (isPast || meeting.status == 'completed')
                      MediaQuery.of(context).size.width <= 600
                        ? _buildActionButton(
                            icon: Icons.video_library,
                            onPressed: () => _openMeetingDetails(meeting),
                          )
                        : Container(
                            height: 40,
                            decoration: BoxDecoration(
                              color: AppColors.primary50,
                              borderRadius: BorderRadius.circular(24),
                              border: Border.all(color: AppColors.primary200),
                            ),
                            child: ElevatedButton.icon(
                              onPressed: () => _openMeetingDetails(meeting),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Colors.transparent,
                                foregroundColor: AppColors.primary600,
                                shadowColor: Colors.transparent,
                                padding: const EdgeInsets.symmetric(horizontal: 16),
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(24),
                                ),
                              ),
                              icon: const Icon(Icons.video_library, size: 18),
                              label: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Text(
                                    l10n.viewRecordings,
                                    style: const TextStyle(
                                      fontWeight: FontWeight.w600,
                                      fontSize: 14,
                                    ),
                                  ),
                                  const SizedBox(width: 4),
                                  const Icon(Icons.arrow_forward, size: 16),
                                ],
                              ),
                            ),
                          ),
                    if (isPast || meeting.status == 'completed') const SizedBox(width: 8),

                    const Spacer(),

                    // Для постоянных встреч всегда показываем кнопку присоединения
                    if (meeting.isPermanent || (canJoin && !isPast))
                      Container(
                        height: 40,
                        decoration: BoxDecoration(
                          gradient: const LinearGradient(
                            colors: [AppColors.primary500, AppColors.primary600],
                          ),
                          borderRadius: BorderRadius.circular(24),
                          boxShadow: [
                            BoxShadow(
                              color: AppColors.primary500.withValues(alpha: 0.3),
                              blurRadius: 8,
                              offset: const Offset(0, 2),
                            ),
                          ],
                        ),
                        child: ElevatedButton(
                          onPressed: () => _openMeetingDetails(meeting),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.transparent,
                            foregroundColor: Colors.white,
                            shadowColor: Colors.transparent,
                            padding: const EdgeInsets.symmetric(horizontal: 20),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(24),
                            ),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Icon(Icons.play_arrow_rounded, size: 20),
                              const SizedBox(width: 6),
                              Text(
                                l10n.join,
                                style: const TextStyle(
                                  fontWeight: FontWeight.w600,
                                  fontSize: 14,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildFilterPill(String label, String value) {
    final isSelected = _filter == value;
    return Container(
      decoration: BoxDecoration(
        color: isSelected ? AppColors.primary50 : Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: isSelected ? AppColors.primary200 : AppColors.border,
        ),
        boxShadow: isSelected
            ? [
                BoxShadow(
                  color: AppColors.primary500.withValues(alpha: 0.12),
                  blurRadius: 12,
                  offset: const Offset(0, 4),
                ),
              ]
            : [],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            setState(() => _filter = value);
            _loadMeetings();
          },
          borderRadius: BorderRadius.circular(24),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
            child: Text(
              label,
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: isSelected ? AppColors.primary700 : AppColors.textSecondary,
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildChip(String text, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: AppColors.primary50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.primary100),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: AppColors.primary600),
          const SizedBox(width: 4),
          Text(
            text,
            style: const TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w500,
              color: AppColors.primary700,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMetadata(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 16, color: AppColors.textSecondary),
        const SizedBox(width: 4),
        Text(
          text,
          style: const TextStyle(
            fontSize: 14,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildActionButton({
    required IconData icon,
    required VoidCallback onPressed,
  }) {
    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.border),
      ),
      child: IconButton(
        icon: Icon(icon, size: 20),
        color: AppColors.textSecondary,
        onPressed: onPressed,
        padding: EdgeInsets.zero,
      ),
    );
  }
}

// Empty state widget
class _EmptyState extends StatelessWidget {
  final VoidCallback onCreateMeeting;
  final AppLocalizations l10n;

  const _EmptyState({
    required this.onCreateMeeting,
    required this.l10n,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(40),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(32),
              decoration: BoxDecoration(
                color: AppColors.primary50,
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.event_note_outlined,
                size: 64,
                color: AppColors.primary500,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              l10n.noMeetingsFound,
              style: const TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            const Text(
              'Create a new meeting to get started',
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 14,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(height: 32),
            ElevatedButton.icon(
              onPressed: onCreateMeeting,
              icon: const Icon(Icons.add_rounded),
              label: Text(l10n.newMeeting),
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primary500,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 14),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(24),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
