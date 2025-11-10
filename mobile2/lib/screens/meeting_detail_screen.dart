import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/meeting.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/config_service.dart';
import '../widgets/error_display.dart';
import 'video_call_screen.dart';

class MeetingDetailScreen extends StatefulWidget {
  final String meetingId;

  const MeetingDetailScreen({
    super.key,
    required this.meetingId,
  });

  @override
  State<MeetingDetailScreen> createState() => _MeetingDetailScreenState();
}

class _MeetingDetailScreenState extends State<MeetingDetailScreen> {
  final _configService = ConfigService();
  late MeetingsService _meetingsService;

  MeetingWithDetails? _meeting;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _initService();
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
        child: CircularProgressIndicator(),
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

      // Close loading dialog
      Navigator.of(context).pop();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Failed to join: ${e.message}'),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 5),
        ),
      );
    } catch (e) {
      if (!mounted) return;

      // Close loading dialog
      Navigator.of(context).pop();

      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Error: $e'),
          backgroundColor: Colors.red,
          duration: const Duration(seconds: 5),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Meeting Details'),
        backgroundColor: const Color(0xFF46afba),
        foregroundColor: Colors.white,
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? FullScreenError(
                  error: _error!,
                  onRetry: _loadMeeting,
                  title: 'Failed to Load Meeting',
                )
              : _meeting == null
                  ? const Center(child: Text('Meeting not found'))
                  : _buildMeetingContent(),
      floatingActionButton: _meeting != null &&
              (_meeting!.status == 'active' || _meeting!.status == 'scheduled')
          ? FloatingActionButton.extended(
              onPressed: _joinMeeting,
              backgroundColor: const Color(0xFF46afba),
              icon: const Icon(Icons.video_call),
              label: const Text('Join Meeting'),
            )
          : null,
    );
  }

  Widget _buildMeetingContent() {
    final meeting = _meeting!;
    final dateFormat = DateFormat('MMM d, y');
    final timeFormat = DateFormat('HH:mm');

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with title and status
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  const Color(0xFF46afba),
                  const Color(0xFF46afba).withValues(alpha: 0.7),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildStatusChip(meeting.status),
                const SizedBox(height: 12),
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
                  title: 'Date & Time',
                  content: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        dateFormat.format(meeting.scheduledAt),
                        style: const TextStyle(fontSize: 16),
                      ),
                      Text(
                        '${timeFormat.format(meeting.scheduledAt)} (${meeting.duration} min)',
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.grey[600],
                        ),
                      ),
                      if (meeting.recurrence != null) ...[
                        const SizedBox(height: 4),
                        Text(
                          'Recurrence: ${meeting.recurrence}',
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

                _buildInfoSection(
                  icon: Icons.people,
                  title: 'Participants (${meeting.participants.length})',
                  content: meeting.participants.isEmpty
                      ? Text(
                          'No participants',
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
                                    backgroundColor: const Color(0xFF46afba).withValues(alpha: 0.2),
                                    child: Text(
                                      participant.displayName.substring(0, 1).toUpperCase(),
                                      style: const TextStyle(
                                        color: Color(0xFF46afba),
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

                if (meeting.departments.isNotEmpty) ...[
                  const Divider(height: 32),
                  _buildInfoSection(
                    icon: Icons.business,
                    title: 'Departments',
                    content: Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: meeting.departments.map((dept) {
                        return Chip(
                          label: Text(dept),
                          backgroundColor: Colors.grey[200],
                        );
                      }).toList(),
                    ),
                  ),
                ],

                const Divider(height: 32),

                _buildInfoSection(
                  icon: Icons.settings,
                  title: 'Settings',
                  content: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSettingRow(
                        'Type',
                        meeting.type,
                      ),
                      _buildSettingRow(
                        'Video Recording',
                        meeting.needsVideoRecord ? 'Enabled' : 'Disabled',
                      ),
                      _buildSettingRow(
                        'Audio Recording',
                        meeting.needsAudioRecord ? 'Enabled' : 'Disabled',
                      ),
                      if (meeting.liveKitRoomId != null)
                        _buildSettingRow(
                          'Room ID',
                          meeting.liveKitRoomId!,
                        ),
                    ],
                  ),
                ),

                if (meeting.additionalNotes != null) ...[
                  const Divider(height: 32),
                  _buildInfoSection(
                    icon: Icons.notes,
                    title: 'Notes',
                    content: Text(
                      meeting.additionalNotes!,
                      style: const TextStyle(fontSize: 14),
                    ),
                  ),
                ],

                const Divider(height: 32),

                _buildInfoSection(
                  icon: Icons.info_outline,
                  title: 'Metadata',
                  content: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSettingRow(
                        'Created',
                        dateFormat.format(meeting.createdAt),
                      ),
                      _buildSettingRow(
                        'Updated',
                        dateFormat.format(meeting.updatedAt),
                      ),
                      _buildSettingRow(
                        'Created By',
                        meeting.createdBy,
                      ),
                    ],
                  ),
                ),

                const SizedBox(height: 80), // Space for FAB
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    Color backgroundColor;
    Color textColor;
    String label;

    switch (status.toLowerCase()) {
      case 'active':
        backgroundColor = Colors.green;
        textColor = Colors.white;
        label = 'Active';
        break;
      case 'scheduled':
        backgroundColor = Colors.blue;
        textColor = Colors.white;
        label = 'Scheduled';
        break;
      case 'completed':
        backgroundColor = Colors.grey;
        textColor = Colors.white;
        label = 'Completed';
        break;
      case 'cancelled':
        backgroundColor = Colors.red;
        textColor = Colors.white;
        label = 'Cancelled';
        break;
      default:
        backgroundColor = Colors.grey;
        textColor = Colors.white;
        label = status;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(12),
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
              color: const Color(0xFF46afba),
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
}
