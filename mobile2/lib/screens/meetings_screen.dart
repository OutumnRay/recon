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
  String _filter = 'scheduled'; // scheduled, in_progress, completed, all
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
          _error = 'Connection Error:\n${e.toString()}\n\nPlease check:\n• Network connection\n• API server is running\n• You are logged in';
          _isLoading = false;
        });
      }
    } catch (e, stackTrace) {
      if (mounted) {
        setState(() {
          _error = 'Unexpected Error:\n${e.toString()}\n\nStack Trace:\n${stackTrace.toString().split('\n').take(5).join('\n')}';
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
        return Icons.people;
      case 'presentation':
        return Icons.present_to_all;
      case 'training':
        return Icons.school;
      default:
        return Icons.event;
    }
  }

  Color _getStatusColor(String status) {
    switch (status) {
      case 'scheduled':
        return const Color(0xFF2563EB); // Blue from frontend (primary-700)
      case 'in_progress':
        return const Color(0xFF92400E); // Amber/brown from frontend (warning)
      case 'completed':
        return const Color(0xFF059669); // Green from frontend (success/secondary-700)
      case 'cancelled':
        return const Color(0xFF991B1B); // Red from frontend (error)
      default:
        return const Color(0xFF6B7280); // Gray default
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

  @override
  Widget build(BuildContext context) {
    super.build(context); // Required for AutomaticKeepAliveClientMixin
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: Text(l10n.meetings),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF26C6DA),
        elevation: 1,
        shadowColor: Colors.black.withValues(alpha: 0.1),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadMeetings,
            tooltip: l10n.refresh,
          ),
        ],
        bottom: PreferredSize(
          preferredSize: const Size.fromHeight(48),
          child: SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
            child: Row(
              children: [
                _buildFilterChip(l10n.filterAll, 'all'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterScheduled, 'scheduled'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterInProgress, 'in_progress'),
                const SizedBox(width: 8),
                _buildFilterChip(l10n.filterCompleted, 'completed'),
              ],
            ),
          ),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? FullScreenError(
                  error: _error!,
                  onRetry: _loadMeetings,
                  title: l10n.failedToLoadMeetings,
                )
              : _meetings == null || _meetings!.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.event_busy,
                              size: 64, color: Colors.grey.shade400),
                          const SizedBox(height: 16),
                          Text(
                            l10n.noMeetingsFound,
                            style: Theme.of(context).textTheme.titleLarge,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _filter == 'all'
                                ? l10n.createFirstMeeting
                                : l10n.tryChangingFilter,
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ],
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: _loadMeetings,
                      child: ListView.builder(
                        itemCount: _meetings!.length,
                        padding: const EdgeInsets.all(8),
                        itemBuilder: (context, index) {
                          final meeting = _meetings![index];
                          return _buildMeetingCard(meeting, context);
                        },
                      ),
                    ),
      floatingActionButton: FloatingActionButton.extended(
        heroTag: 'meetings_fab',
        onPressed: () async {
          final result = await Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => CreateMeetingScreen(
                apiClient: widget.apiClient,
              ),
            ),
          );

          // Reload meetings if a meeting was created
          if (result == true) {
            _loadMeetings();
          }
        },
        backgroundColor: const Color(0xFF26C6DA),
        foregroundColor: Colors.white,
        elevation: 4,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(20),
        ),
        icon: const Icon(Icons.add),
        label: Text(l10n.newMeeting),
      ),
    );
  }

  Widget _buildFilterChip(String label, String value) {
    final isSelected = _filter == value;
    return FilterChip(
      label: Text(label),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          setState(() {
            _filter = value;
          });
          _loadMeetings();
        }
      },
      showCheckmark: false,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ),
      backgroundColor: Colors.white,
      selectedColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
      side: BorderSide(
        color: isSelected ? const Color(0xFF26C6DA) : Colors.grey.shade300,
        width: 1.5,
      ),
      labelStyle: TextStyle(
        color: isSelected ? const Color(0xFF00ACC1) : Colors.grey.shade700,
        fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
      ),
    );
  }

  Future<void> _copyAnonymousLink(String meetingId) async {
    final l10n = AppLocalizations.of(context)!;

    try {
      final configService = ConfigService();
      final apiUrl = await configService.getApiUrl();
      // Remove /api/v1 from the API URL to get base URL
      final baseUrl = apiUrl.replaceAll('/api/v1', '');
      final anonymousLink = '$baseUrl/meeting/$meetingId/join';

      await Clipboard.setData(ClipboardData(text: anonymousLink));

      if (mounted) {
        setState(() {
          _copiedMeetingId = meetingId;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(l10n.anonymousLinkCopied),
            backgroundColor: Colors.green,
            duration: const Duration(seconds: 2),
          ),
        );

        // Reset copied state after 2 seconds
        Future.delayed(const Duration(seconds: 2), () {
          if (mounted) {
            setState(() {
              _copiedMeetingId = null;
            });
          }
        });
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to copy link: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Widget _buildMeetingCard(MeetingWithDetails meeting, BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final totalParticipants = meeting.participants.length;
    final onlineParticipants = meeting.activeParticipantsCount;
    final anonymousGuests = meeting.anonymousGuestsCount;

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      elevation: 2,
      child: InkWell(
        onTap: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => MeetingDetailScreen(meetingId: meeting.id),
            ),
          );
        },
        borderRadius: BorderRadius.circular(20),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header row with icon, title, and status badge
              Row(
                children: [
                  CircleAvatar(
                    radius: 24,
                    backgroundColor: const Color(0xFF26C6DA).withValues(alpha: 0.15),
                    child: Icon(
                      _getMeetingIcon(meeting.type),
                      color: const Color(0xFF00ACC1),
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          meeting.title,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                        const SizedBox(height: 2),
                        Text(
                          _formatDateTime(meeting.scheduledAt, context),
                          style: TextStyle(
                            fontSize: 13,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  // Status or Permanent badge
                  if (meeting.isPermanent || meeting.recurrence == 'permanent')
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: const Color(0xFF7C3AED),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.all_inclusive, size: 12, color: Colors.white),
                          const SizedBox(width: 4),
                          Text(
                            l10n.permanent,
                            style: const TextStyle(
                              fontSize: 10,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: _getStatusColor(meeting.status).withValues(alpha: 0.15),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(
                          color: _getStatusColor(meeting.status).withValues(alpha: 0.3),
                          width: 1,
                        ),
                      ),
                      child: Text(
                        _getStatusText(meeting.status, context),
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: _getStatusColor(meeting.status),
                        ),
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 12),
              // Participant info row
              Row(
                children: [
                  Icon(Icons.timer, size: 14, color: Colors.grey[600]),
                  const SizedBox(width: 4),
                  Text(
                    '${meeting.duration} ${l10n.minutes}',
                    style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                  ),
                  const SizedBox(width: 16),
                  Icon(Icons.people, size: 14, color: Colors.grey[600]),
                  const SizedBox(width: 4),
                  Text(
                    l10n.participantsSummary(totalParticipants, onlineParticipants),
                    style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                  ),
                ],
              ),
              // Online indicators
              if (onlineParticipants > 0 || (meeting.allowAnonymous && anonymousGuests > 0)) ...[
                const SizedBox(height: 8),
                Wrap(
                  spacing: 12,
                  runSpacing: 6,
                  children: [
                    if (onlineParticipants > 0)
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: const BoxDecoration(
                              color: Color(0xFF059669),
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            l10n.onlineCount(onlineParticipants),
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF059669),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                    if (meeting.allowAnonymous && anonymousGuests > 0)
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Container(
                            width: 8,
                            height: 8,
                            decoration: const BoxDecoration(
                              color: Color(0xFF2563EB),
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            l10n.anonymousGuestsCount(anonymousGuests),
                            style: const TextStyle(
                              fontSize: 12,
                              color: Color(0xFF2563EB),
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ],
                      ),
                  ],
                ),
              ],
              // Anonymous link copy button
              if (meeting.allowAnonymous) ...[
                const SizedBox(height: 12),
                ElevatedButton.icon(
                  onPressed: () => _copyAnonymousLink(meeting.id),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: _copiedMeetingId == meeting.id
                        ? const Color(0xFF059669)
                        : const Color(0xFF26C6DA),
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    elevation: 2,
                  ),
                  icon: Icon(
                    _copiedMeetingId == meeting.id ? Icons.check : Icons.link,
                    size: 18,
                  ),
                  label: Text(
                    _copiedMeetingId == meeting.id ? l10n.linkCopied : l10n.copyLink,
                    style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
