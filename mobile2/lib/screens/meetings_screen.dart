import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
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
        shadowColor: Colors.black.withOpacity(0.1),
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
                          return Card(
                            margin: const EdgeInsets.symmetric(
                              vertical: 8,
                              horizontal: 12,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(20),
                            ),
                            elevation: 2,
                            child: ListTile(
                              contentPadding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 8,
                              ),
                              onTap: () {
                                Navigator.push(
                                  context,
                                  MaterialPageRoute(
                                    builder: (context) => MeetingDetailScreen(
                                      meetingId: meeting.id,
                                    ),
                                  ),
                                );
                              },
                              leading: CircleAvatar(
                                radius: 28,
                                backgroundColor: const Color(0xFF26C6DA).withOpacity(0.15),
                                child: Icon(
                                  _getMeetingIcon(meeting.type),
                                  color: const Color(0xFF00ACC1),
                                  size: 26,
                                ),
                              ),
                              title: Text(
                                meeting.title,
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                              subtitle: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  const SizedBox(height: 4),
                                  Row(
                                    children: [
                                      Icon(Icons.access_time,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Expanded(
                                        child: Text(
                                          _formatDateTime(meeting.scheduledAt, context),
                                          style: TextStyle(
                                              fontSize: 13,
                                              color: Colors.grey[600]),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 2),
                                  Row(
                                    children: [
                                      Icon(Icons.timer,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Text(
                                        '${meeting.duration} ${l10n.minutes}',
                                        style: TextStyle(
                                            fontSize: 13,
                                            color: Colors.grey[600]),
                                      ),
                                      const SizedBox(width: 12),
                                      Icon(Icons.people,
                                          size: 14, color: Colors.grey[600]),
                                      const SizedBox(width: 4),
                                      Text(
                                        '${meeting.participants.length}',
                                        style: TextStyle(
                                            fontSize: 13,
                                            color: Colors.grey[600]),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                              trailing: SizedBox(
                                width: 95,
                                child: Column(
                                  mainAxisAlignment: MainAxisAlignment.center,
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    // Show "Permanent" badge for permanent meetings instead of status
                                    if (meeting.isPermanent || meeting.recurrence == 'permanent')
                                      Container(
                                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                                        decoration: BoxDecoration(
                                          color: const Color(0xFF7C3AED), // Purple from frontend
                                          borderRadius: BorderRadius.circular(12),
                                        ),
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            const Icon(
                                              Icons.all_inclusive,
                                              size: 10,
                                              color: Colors.white,
                                            ),
                                            const SizedBox(width: 2),
                                            Flexible(
                                              child: Text(
                                                l10n.permanent,
                                                style: const TextStyle(
                                                  fontSize: 9,
                                                  fontWeight: FontWeight.bold,
                                                  color: Colors.white,
                                                ),
                                                maxLines: 1,
                                                overflow: TextOverflow.ellipsis,
                                              ),
                                            ),
                                          ],
                                        ),
                                      )
                                    else
                                      // Show status badge for non-permanent meetings
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
                              ),
                              isThreeLine: true,
                            ),
                          );
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
}
