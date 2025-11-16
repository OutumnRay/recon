import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../l10n/app_localizations.dart';
import '../models/meeting.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/users_service.dart';
import '../theme/app_colors.dart';
import '../widgets/app_card.dart';
import '../widgets/error_display.dart';

class CreateMeetingScreen extends StatefulWidget {
  final ApiClient apiClient;

  const CreateMeetingScreen({super.key, required this.apiClient});

  @override
  State<CreateMeetingScreen> createState() => _CreateMeetingScreenState();
}

class _CreateMeetingScreenState extends State<CreateMeetingScreen> {
  late final MeetingsService _meetingsService;
  late final UsersService _usersService;

  final _formKey = GlobalKey<FormState>();
  final _titleController = TextEditingController();
  final _notesController = TextEditingController();

  DateTime _scheduledDate = DateTime.now().add(const Duration(hours: 1));
  TimeOfDay _scheduledTime = TimeOfDay.now();
  int _duration = 60; // minutes
  String _type = 'conference';
  String? _subjectId;
  String? _recurrence;
  bool _needsVideoRecord = true;
  bool _needsAudioRecord = true;

  List<MeetingSubject>? _subjects;
  List<UserListItem>? _users;
  List<Department>? _departments;
  Set<String> _selectedParticipantIds = {};
  Set<String> _selectedDepartmentIds = {};
  String? _selectedSpeakerId;

  bool _isLoading = false;
  bool _isLoadingData = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _meetingsService = MeetingsService(widget.apiClient);
    _usersService = UsersService(widget.apiClient);
    _loadInitialData();

    // Set initial time
    final now = DateTime.now();
    _scheduledTime = TimeOfDay(hour: now.hour + 1, minute: 0);
  }

  @override
  void dispose() {
    _titleController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _loadInitialData() async {
    setState(() {
      _isLoadingData = true;
      _error = null;
    });

    try {
      final results = await Future.wait([
        _meetingsService.getMeetingSubjects(),
        _usersService.getUsers(),
        _usersService.getDepartments(),
      ]);

      if (mounted) {
        setState(() {
          _subjects = results[0] as List<MeetingSubject>;
          _users = results[1] as List<UserListItem>;
          _departments = results[2] as List<Department>;
          _isLoadingData = false;

          // Set default subject if available
          if (_subjects != null && _subjects!.isNotEmpty) {
            _subjectId = _subjects!.first.id;
          }
        });
      }
    } catch (e) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        setState(() {
          _error = '${l10n?.failedToLoadFormData ?? 'Failed to load data'}: ${e.toString()}';
          _isLoadingData = false;
        });
      }
    }
  }

  Future<void> _selectDate() async {
    final pickedDate = await showDatePicker(
      context: context,
      initialDate: _scheduledDate,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (pickedDate != null) {
      setState(() {
        _scheduledDate = pickedDate;
      });
    }
  }

  Future<void> _selectTime() async {
    final pickedTime = await showTimePicker(
      context: context,
      initialTime: _scheduledTime,
    );

    if (pickedTime != null) {
      setState(() {
        _scheduledTime = pickedTime;
      });
    }
  }

  Future<void> _selectParticipants() async {
    if (_users == null) return;

    final selected = await showDialog<Set<String>>(
      context: context,
      builder: (context) => _ParticipantSelectionDialog(
        users: _users!,
        selectedIds: _selectedParticipantIds,
        selectedSpeakerId: _selectedSpeakerId,
      ),
    );

    if (selected != null) {
      setState(() {
        _selectedParticipantIds = selected;
      });
    }
  }

  Future<void> _selectDepartments() async {
    if (_departments == null) return;

    final selected = await showDialog<Set<String>>(
      context: context,
      builder: (context) => _DepartmentSelectionDialog(
        departments: _departments!,
        selectedIds: _selectedDepartmentIds,
      ),
    );

    if (selected != null) {
      setState(() {
        _selectedDepartmentIds = selected;
      });
    }
  }

  Future<void> _selectSpeaker() async {
    if (_users == null) return;

    final selected = await showDialog<String>(
      context: context,
      builder: (context) => _SpeakerSelectionDialog(
        users: _users!,
        selectedSpeakerId: _selectedSpeakerId,
      ),
    );

    if (selected != null) {
      setState(() {
        _selectedSpeakerId = selected;
      });
    }
  }

  Future<void> _createMeeting() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final l10n = AppLocalizations.of(context)!;

    if (_subjectId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.pleaseSelectSubject)),
      );
      return;
    }

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      // Combine date and time
      final scheduledAt = DateTime(
        _scheduledDate.year,
        _scheduledDate.month,
        _scheduledDate.day,
        _scheduledTime.hour,
        _scheduledTime.minute,
      );

      final request = CreateMeetingRequest(
        title: _titleController.text,
        scheduledAt: scheduledAt,
        duration: _duration,
        recurrence: _recurrence,
        type: _type,
        subjectId: _subjectId!,
        needsVideoRecord: _needsVideoRecord,
        needsAudioRecord: _needsAudioRecord,
        additionalNotes:
            _notesController.text.isEmpty ? null : _notesController.text,
        participantIds: _selectedParticipantIds.toList(),
        departmentIds: _selectedDepartmentIds.toList(),
        speakerId: _selectedSpeakerId,
      );

      await _meetingsService.createMeeting(request);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.meetingCreatedSuccess)),
        );
        Navigator.pop(context, true); // Return true to indicate success
      }
    } on ApiException catch (e) {
      if (mounted) {
        setState(() {
          _error = '${l10n.meetingCreatedError}: ${e.message}';
          _isLoading = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_error!)),
        );
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = '${l10n.unexpectedError}: ${e.toString()}';
          _isLoading = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_error!)),
        );
      }
    }
  }

  String _formatDateTime(BuildContext context, DateTime date, TimeOfDay time) {
    final dateTime = DateTime(date.year, date.month, date.day, time.hour, time.minute);
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat.yMMMMEEEEd(locale);
    final timeFormat = DateFormat.Hm(locale);
    return '${dateFormat.format(dateTime)} ${timeFormat.format(dateTime)}';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoadingData) {
      return Scaffold(
        backgroundColor: AppColors.surface,
        appBar: AppBar(
          title: Text(l10n.createMeeting),
          backgroundColor: AppColors.surface,
          foregroundColor: AppColors.textPrimary,
          elevation: 0,
        ),
        body: const Center(
          child: CircularProgressIndicator(color: AppColors.primary500),
        ),
      );
    }

    if (_error != null && _subjects == null) {
      return Scaffold(
        backgroundColor: AppColors.surface,
        appBar: AppBar(
          title: Text(l10n.createMeeting),
          backgroundColor: AppColors.surface,
          foregroundColor: AppColors.textPrimary,
          elevation: 0,
        ),
        body: FullScreenError(
          error: _error!,
          onRetry: _loadInitialData,
          title: l10n.failedToLoadFormData,
        ),
      );
    }

    return Scaffold(
      backgroundColor: AppColors.surface,
      appBar: AppBar(
        title: Text(l10n.createMeeting),
        backgroundColor: AppColors.surface,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        actions: [
          if (_isLoading)
            const Padding(
              padding: EdgeInsets.all(16),
              child: SizedBox(
                width: 20,
                height: 20,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  color: AppColors.primary500,
                ),
              ),
            ),
        ],
      ),
      body: SafeArea(
        child: Form(
          key: _formKey,
          child: ListView(
            physics: const BouncingScrollPhysics(),
            padding: const EdgeInsets.fromLTRB(20, 24, 20, 140),
            children: [
              _buildHeroCard(context, l10n),
              const SizedBox(height: 20),
              _buildGeneralInfoCard(context, l10n),
              const SizedBox(height: 20),
              _buildScheduleCard(context, l10n),
              const SizedBox(height: 20),
              _buildAudienceCard(context, l10n),
              const SizedBox(height: 20),
              _buildRecordingCard(context, l10n),
              const SizedBox(height: 20),
              _buildNotesCard(context, l10n),
              const SizedBox(height: 24),
              FilledButton.icon(
                onPressed: _isLoading ? null : _createMeeting,
                icon: const Icon(Icons.check_circle_rounded),
                label: Text(l10n.createMeeting),
                style: FilledButton.styleFrom(
                  backgroundColor: AppColors.primary600,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 18),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHeroCard(BuildContext context, AppLocalizations l10n) {
    final stats = [
      _buildHeroStat(
        context,
        icon: Icons.calendar_month_rounded,
        label: l10n.dateAndTime,
        value: _formatDateTime(context, _scheduledDate, _scheduledTime),
      ),
      _buildHeroStat(
        context,
        icon: Icons.timer_outlined,
        label: l10n.meetingDuration,
        value: '${_duration} ${l10n.minutes}',
      ),
      _buildHeroStat(
        context,
        icon: Icons.people_alt_rounded,
        label: l10n.meetingParticipants,
        value: _selectedParticipantIds.isEmpty
            ? l10n.noParticipants
            : l10n.participantsSelected(_selectedParticipantIds.length),
      ),
    ];

    return GradientHeroCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.createMeeting,
            style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 6),
          Text(
            l10n.meetingDetails,
            style: Theme.of(context)
                .textTheme
                .bodyMedium
                ?.copyWith(color: AppColors.textSecondary),
          ),
          const SizedBox(height: 20),
          Wrap(
            spacing: 12,
            runSpacing: 12,
            children: stats,
          ),
        ],
      ),
    );
  }

  Widget _buildGeneralInfoCard(BuildContext context, AppLocalizations l10n) {
    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.badge_rounded,
            title: l10n.meetingDetailsTitle,
            subtitle: l10n.meetingDetails,
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _titleController,
            decoration: _inputDecoration(
              '${l10n.meetingTitle} ${l10n.required}',
              hint: l10n.enterTitle,
              icon: Icons.title_rounded,
            ),
            validator: (value) {
              if (value == null || value.isEmpty) {
                return l10n.pleaseEnterTitle;
              }
              return null;
            },
            enabled: !_isLoading,
          ),
          const SizedBox(height: 14),
          DropdownButtonFormField<String>(
            value: _subjectId,
            decoration: _inputDecoration(
              '${l10n.meetingSubject} ${l10n.required}',
              icon: Icons.category_rounded,
            ),
            items: (_subjects ?? [])
                .map(
                  (subject) => DropdownMenuItem(
                    value: subject.id,
                    child: Text(subject.name),
                  ),
                )
                .toList(),
            onChanged: _isLoading
                ? null
                : (value) => setState(() => _subjectId = value),
            validator: (value) {
              if (value == null) {
                return l10n.pleaseSelectSubject;
              }
              return null;
            },
          ),
          const SizedBox(height: 14),
          DropdownButtonFormField<String>(
            value: _type,
            decoration: _inputDecoration(
              l10n.meetingType,
              icon: Icons.event_note_rounded,
            ),
            items: [
              DropdownMenuItem(value: 'conference', child: Text(l10n.typeConference)),
              DropdownMenuItem(value: 'presentation', child: Text(l10n.typePresentation)),
              DropdownMenuItem(value: 'training', child: Text(l10n.typeTraining)),
              DropdownMenuItem(value: 'discussion', child: Text(l10n.typeDiscussion)),
            ],
            onChanged: _isLoading
                ? null
                : (value) {
                    if (value != null) {
                      setState(() => _type = value);
                    }
                  },
          ),
        ],
      ),
    );
  }

  Widget _buildScheduleCard(BuildContext context, AppLocalizations l10n) {
    final locale = Localizations.localeOf(context).toString();
    final dateFormat = DateFormat.yMMMd(locale);

    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.schedule_rounded,
            title: l10n.dateAndTime,
            subtitle: _formatDateTime(context, _scheduledDate, _scheduledTime),
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _isLoading ? null : _selectDate,
                  icon: const Icon(Icons.calendar_today_rounded),
                  label: Text(dateFormat.format(_scheduledDate)),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.primary600,
                    side: BorderSide(color: AppColors.border),
                    padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(18),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: OutlinedButton.icon(
                  onPressed: _isLoading ? null : _selectTime,
                  icon: const Icon(Icons.access_time_rounded),
                  label: Text(_scheduledTime.format(context)),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.primary600,
                    side: BorderSide(color: AppColors.border),
                    padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(18),
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          DropdownButtonFormField<int>(
            value: _duration,
            decoration: _inputDecoration(
              l10n.meetingDuration,
              icon: Icons.timer_outlined,
            ),
            items: [
              DropdownMenuItem(value: 15, child: Text(l10n.duration15)),
              DropdownMenuItem(value: 30, child: Text(l10n.duration30)),
              DropdownMenuItem(value: 45, child: Text(l10n.duration45)),
              DropdownMenuItem(value: 60, child: Text(l10n.duration60)),
              DropdownMenuItem(value: 90, child: Text(l10n.duration90)),
              DropdownMenuItem(value: 120, child: Text(l10n.duration120)),
              DropdownMenuItem(value: 180, child: Text(l10n.duration180)),
            ],
            onChanged: _isLoading
                ? null
                : (value) {
                    if (value != null) {
                      setState(() => _duration = value);
                    }
                  },
          ),
          const SizedBox(height: 14),
          DropdownButtonFormField<String?>(
            value: _recurrence,
            decoration: _inputDecoration(
              l10n.meetingRecurrence,
              icon: Icons.repeat_rounded,
            ),
            items: [
              DropdownMenuItem(value: null, child: Text(l10n.recurrenceNone)),
              DropdownMenuItem(value: 'daily', child: Text(l10n.recurrenceDaily)),
              DropdownMenuItem(value: 'weekly', child: Text(l10n.recurrenceWeekly)),
              DropdownMenuItem(value: 'monthly', child: Text(l10n.recurrenceMonthly)),
            ],
            onChanged: _isLoading
                ? null
                : (value) => setState(() => _recurrence = value),
          ),
        ],
      ),
    );
  }

  Widget _buildAudienceCard(BuildContext context, AppLocalizations l10n) {
    final participantNames = _selectedParticipantIds.map(_getUserName).toList();
    final departmentNames = _selectedDepartmentIds.map(_getDepartmentName).toList();
    final speakerName =
        _selectedSpeakerId != null ? _getUserName(_selectedSpeakerId!) : l10n.noSpeaker;

    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.group_work_rounded,
            title: l10n.meetingParticipants,
            subtitle: l10n.participantsSelected(_selectedParticipantIds.length),
          ),
          const SizedBox(height: 16),
          _buildSelectionTile(
            context,
            icon: Icons.person_outline_rounded,
            title: '${l10n.meetingSpeaker} ${l10n.optional}',
            subtitle: speakerName,
            onTap: _isLoading ? null : _selectSpeaker,
          ),
          const SizedBox(height: 12),
          _buildSelectionTile(
            context,
            icon: Icons.people_alt_rounded,
            title: l10n.meetingParticipants,
            subtitle: l10n.participantsSelected(_selectedParticipantIds.length),
            onTap: _isLoading ? null : _selectParticipants,
          ),
          if (participantNames.isNotEmpty) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _buildChipList(participantNames),
            ),
          ],
          const SizedBox(height: 12),
          _buildSelectionTile(
            context,
            icon: Icons.business_outlined,
            title: l10n.meetingDepartments,
            subtitle: l10n.departmentsSelected(_selectedDepartmentIds.length),
            onTap: _isLoading ? null : _selectDepartments,
          ),
          if (departmentNames.isNotEmpty) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: _buildChipList(departmentNames),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildRecordingCard(BuildContext context, AppLocalizations l10n) {
    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.settings_input_component_rounded,
            title: l10n.recordingOptions,
            subtitle: l10n.settingsTitle,
          ),
          _buildToggleTile(
            context,
            title: l10n.meetingVideoRecord,
            subtitle: _needsVideoRecord ? l10n.enabled : l10n.disabled,
            value: _needsVideoRecord,
            onChanged: (value) => setState(() => _needsVideoRecord = value),
          ),
          _buildToggleTile(
            context,
            title: l10n.meetingAudioRecord,
            subtitle: _needsAudioRecord ? l10n.enabled : l10n.disabled,
            value: _needsAudioRecord,
            onChanged: (value) => setState(() => _needsAudioRecord = value),
          ),
        ],
      ),
    );
  }

  Widget _buildNotesCard(BuildContext context, AppLocalizations l10n) {
    return SurfaceCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildSectionHeader(
            context,
            icon: Icons.notes_rounded,
            title: l10n.meetingNotes,
            subtitle: l10n.enterNotes,
          ),
          const SizedBox(height: 16),
          TextFormField(
            controller: _notesController,
            maxLines: 4,
            decoration: _inputDecoration(
              l10n.meetingNotes,
              hint: l10n.enterNotes,
              icon: Icons.description_outlined,
            ).copyWith(alignLabelWithHint: true),
            enabled: !_isLoading,
          ),
        ],
      ),
    );
  }
  Widget _buildSelectionTile(
    BuildContext context, {
    required IconData icon,
    required String title,
    required String subtitle,
    VoidCallback? onTap,
  }) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: AppColors.surfaceMuted,
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: AppColors.border),
          ),
          child: Row(
            children: [
              Icon(icon, color: AppColors.primary600),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      subtitle,
                      style: Theme.of(context)
                          .textTheme
                          .bodySmall
                          ?.copyWith(color: AppColors.textSecondary),
                    ),
                  ],
                ),
              ),
              Icon(
                Icons.chevron_right_rounded,
                color: onTap != null ? AppColors.textSecondary : AppColors.textTertiary,
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildToggleTile(
    BuildContext context, {
    required String title,
    required String subtitle,
    required bool value,
    required ValueChanged<bool> onChanged,
  }) {
    return Container(
      margin: const EdgeInsets.only(top: 12),
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.border),
      ),
      child: SwitchListTile.adaptive(
        value: value,
        onChanged: _isLoading ? null : onChanged,
        title: Text(title),
        subtitle: Text(
          subtitle,
          style: Theme.of(context)
              .textTheme
              .bodySmall
              ?.copyWith(color: AppColors.textSecondary),
        ),
        activeColor: AppColors.primary600,
        contentPadding: EdgeInsets.zero,
      ),
    );
  }

  List<Widget> _buildChipList(List<String> labels) {
    if (labels.isEmpty) return [];
    const maxVisible = 4;
    final chips = <Widget>[];
    for (var i = 0; i < labels.length && i < maxVisible; i++) {
      chips.add(_buildChip(labels[i]));
    }
    final remaining = labels.length - maxVisible;
    if (remaining > 0) {
      chips.add(_buildChip('+${remaining}'));
    }
    return chips;
  }

  Widget _buildChip(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: AppColors.primary50,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.primary200),
      ),
      child: Text(
        label,
        style: const TextStyle(
          color: AppColors.primary700,
          fontWeight: FontWeight.w600,
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
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
            color: AppColors.primary50,
            borderRadius: BorderRadius.circular(16),
          ),
          child: Icon(icon, color: AppColors.primary600, size: 22),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
              ),
              if (subtitle != null) ...[
                const SizedBox(height: 4),
                Text(
                  subtitle,
                  style: Theme.of(context)
                      .textTheme
                      .bodySmall
                      ?.copyWith(color: AppColors.textSecondary),
                ),
              ],
            ],
          ),
        ),
      ],
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
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.border),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Icon(icon, size: 18, color: AppColors.primary600),
            const SizedBox(width: 10),
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
                    style: textTheme.titleMedium?.copyWith(
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

  InputDecoration _inputDecoration(
    String label, {
    String? hint,
    IconData? icon,
  }) {
    return InputDecoration(
      labelText: label,
      hintText: hint,
      prefixIcon: icon != null ? Icon(icon, color: AppColors.primary500) : null,
      filled: true,
      fillColor: AppColors.surfaceMuted,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(20),
        borderSide: BorderSide(color: AppColors.border),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(20),
        borderSide: BorderSide(color: AppColors.border),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(20),
        borderSide: const BorderSide(color: AppColors.primary500, width: 2),
      ),
    );
  }

  String _getUserName(String id) {
    if (_users != null) {
      for (final user in _users!) {
        if (user.id == id) {
          return user.displayName;
        }
      }
    }
    return id;
  }

  String _getDepartmentName(String id) {
    if (_departments != null) {
      for (final dept in _departments!) {
        if (dept.id == id) {
          return dept.name;
        }
      }
    }
    return id;
  }

// Participant Selection Dialog
class _ParticipantSelectionDialog extends StatefulWidget {
  final List<UserListItem> users;
  final Set<String> selectedIds;
  final String? selectedSpeakerId;

  const _ParticipantSelectionDialog({
    required this.users,
    required this.selectedIds,
    this.selectedSpeakerId,
  });

  @override
  State<_ParticipantSelectionDialog> createState() =>
      _ParticipantSelectionDialogState();
}

class _ParticipantSelectionDialogState
    extends State<_ParticipantSelectionDialog> {
  late Set<String> _selected;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _selected = Set.from(widget.selectedIds);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    final filteredUsers = widget.users.where((user) {
      final query = _searchQuery.toLowerCase();
      return user.displayName.toLowerCase().contains(query) ||
          user.email.toLowerCase().contains(query);
    }).toList();

    return AlertDialog(
      backgroundColor: AppColors.surfaceCard,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      title: Text(
        l10n.selectParticipants,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w600,
            ),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: InputDecoration(
                hintText: l10n.searchUsers,
                prefixIcon:
                    const Icon(Icons.search_rounded, color: AppColors.primary600),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: AppColors.border),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: AppColors.border),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: AppColors.primary500, width: 2),
                ),
                filled: true,
                fillColor: AppColors.surfaceMuted,
              ),
              onChanged: (value) {
                setState(() {
                  _searchQuery = value;
                });
              },
            ),
            const SizedBox(height: 16),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: filteredUsers.length,
                itemBuilder: (context, index) {
                  final user = filteredUsers[index];
                  final isSelected = _selected.contains(user.id);
                  return CheckboxListTile(
                    title: Text(user.displayName),
                    subtitle: Text(user.email),
                    value: isSelected,
                    activeColor: AppColors.primary600,
                    onChanged: (value) {
                      setState(() {
                        if (value == true) {
                          _selected.add(user.id);
                        } else {
                          _selected.remove(user.id);
                        }
                      });
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          style: TextButton.styleFrom(foregroundColor: AppColors.textSecondary),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selected),
          style: FilledButton.styleFrom(
            backgroundColor: AppColors.primary600,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
          child: Text('${l10n.select} (${_selected.length})'),
        ),
      ],
    );
  }
}

// Department Selection Dialog
class _DepartmentSelectionDialog extends StatefulWidget {
  final List<Department> departments;
  final Set<String> selectedIds;

  const _DepartmentSelectionDialog({
    required this.departments,
    required this.selectedIds,
  });

  @override
  State<_DepartmentSelectionDialog> createState() =>
      _DepartmentSelectionDialogState();
}

class _DepartmentSelectionDialogState
    extends State<_DepartmentSelectionDialog> {
  late Set<String> _selected;

  @override
  void initState() {
    super.initState();
    _selected = Set.from(widget.selectedIds);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return AlertDialog(
      backgroundColor: AppColors.surfaceCard,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      title: Text(
        l10n.selectDepartments,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w600,
            ),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: ListView.builder(
          shrinkWrap: true,
          itemCount: widget.departments.length,
          itemBuilder: (context, index) {
            final dept = widget.departments[index];
            final isSelected = _selected.contains(dept.id);
            return CheckboxListTile(
              title: Text(dept.name),
              subtitle: dept.description != null ? Text(dept.description!) : null,
              value: isSelected,
              activeColor: AppColors.primary600,
              onChanged: (value) {
                setState(() {
                  if (value == true) {
                    _selected.add(dept.id);
                  } else {
                    _selected.remove(dept.id);
                  }
                });
              },
            );
          },
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          style: TextButton.styleFrom(foregroundColor: AppColors.textSecondary),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selected),
          style: FilledButton.styleFrom(
            backgroundColor: AppColors.primary600,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
          child: Text('${l10n.select} (${_selected.length})'),
        ),
      ],
    );
  }
}

// Speaker Selection Dialog
class _SpeakerSelectionDialog extends StatefulWidget {
  final List<UserListItem> users;
  final String? selectedSpeakerId;

  const _SpeakerSelectionDialog({
    required this.users,
    this.selectedSpeakerId,
  });

  @override
  State<_SpeakerSelectionDialog> createState() =>
      _SpeakerSelectionDialogState();
}

class _SpeakerSelectionDialogState extends State<_SpeakerSelectionDialog> {
  String? _selectedId;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _selectedId = widget.selectedSpeakerId;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    final filteredUsers = widget.users.where((user) {
      final query = _searchQuery.toLowerCase();
      return user.displayName.toLowerCase().contains(query) ||
          user.email.toLowerCase().contains(query);
    }).toList();

    return AlertDialog(
      backgroundColor: AppColors.surfaceCard,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      title: Text(
        l10n.selectSpeaker,
        style: Theme.of(context).textTheme.titleMedium?.copyWith(
              color: AppColors.textPrimary,
              fontWeight: FontWeight.w600,
            ),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: InputDecoration(
                hintText: l10n.searchUsers,
                prefixIcon:
                    const Icon(Icons.search_rounded, color: AppColors.primary600),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: AppColors.border),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: AppColors.border),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: AppColors.primary500, width: 2),
                ),
                filled: true,
                fillColor: AppColors.surfaceMuted,
              ),
              onChanged: (value) {
                setState(() {
                  _searchQuery = value;
                });
              },
            ),
            const SizedBox(height: 16),
            ListTile(
              title: Text(l10n.noSpeaker),
              leading: Radio<String?>(
                value: null,
                groupValue: _selectedId,
                activeColor: AppColors.primary600,
                onChanged: (value) {
                  setState(() {
                    _selectedId = value;
                  });
                },
              ),
              onTap: () {
                setState(() {
                  _selectedId = null;
                });
              },
            ),
            const Divider(),
            Flexible(
              child: ListView.builder(
                shrinkWrap: true,
                itemCount: filteredUsers.length,
                itemBuilder: (context, index) {
                  final user = filteredUsers[index];
                  return RadioListTile<String>(
                    title: Text(user.displayName),
                    subtitle: Text(user.email),
                    value: user.id,
                    groupValue: _selectedId,
                    activeColor: AppColors.primary600,
                    onChanged: (value) {
                      setState(() {
                        _selectedId = value;
                      });
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          style: TextButton.styleFrom(foregroundColor: AppColors.textSecondary),
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selectedId),
          style: FilledButton.styleFrom(
            backgroundColor: AppColors.primary600,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
          child: Text(l10n.select),
        ),
      ],
    );
  }
}
