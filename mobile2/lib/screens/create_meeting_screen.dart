import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

import '../l10n/app_localizations.dart';
import '../models/meeting.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/users_service.dart';
import '../theme/app_colors.dart';
import '../utils/logger.dart';
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
  bool _needsTranscription = false;
  bool _forceEndAtDuration = false;
  bool _allowAnonymous = false;

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
          _error =
              '${l10n?.failedToLoadFormData ?? 'Failed to load data'}: ${e.toString()}';
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

  Future<void> _selectSubject() async {
    if (_subjects == null || _subjects!.isEmpty) return;

    final l10n = AppLocalizations.of(context)!;
    final selected = await showModalBottomSheet<String>(
      context: context,
      backgroundColor: Colors.transparent,
      isScrollControlled: true,
      builder: (context) => _buildSelectionBottomSheet(
        title: l10n.meetingSubject,
        items: _subjects!
            .map((s) => _SelectionItem(value: s.id, label: s.name))
            .toList(),
        selectedValue: _subjectId,
      ),
    );

    if (selected != null) {
      setState(() => _subjectId = selected);
    }
  }

  Future<void> _selectType() async {
    final l10n = AppLocalizations.of(context)!;
    final types = [
      _SelectionItem(value: 'conference', label: l10n.typeConference),
      _SelectionItem(value: 'presentation', label: l10n.typePresentation),
      _SelectionItem(value: 'training', label: l10n.typeTraining),
      _SelectionItem(value: 'discussion', label: l10n.typeDiscussion),
    ];

    final selected = await showModalBottomSheet<String>(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => _buildSelectionBottomSheet(
        title: l10n.meetingType,
        items: types,
        selectedValue: _type,
      ),
    );

    if (selected != null) {
      setState(() => _type = selected);
    }
  }

  Future<void> _selectDuration() async {
    final l10n = AppLocalizations.of(context)!;
    final durations = [
      _SelectionItem(value: 15, label: l10n.duration15),
      _SelectionItem(value: 30, label: l10n.duration30),
      _SelectionItem(value: 45, label: l10n.duration45),
      _SelectionItem(value: 60, label: l10n.duration60),
      _SelectionItem(value: 90, label: l10n.duration90),
      _SelectionItem(value: 120, label: l10n.duration120),
      _SelectionItem(value: 180, label: l10n.duration180),
    ];

    final selected = await showModalBottomSheet<int>(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => _buildSelectionBottomSheet<int>(
        title: l10n.meetingDuration,
        items: durations,
        selectedValue: _duration,
      ),
    );

    if (selected != null) {
      setState(() => _duration = selected);
    }
  }

  Future<void> _selectRecurrence() async {
    final l10n = AppLocalizations.of(context)!;
    final recurrences = [
      _SelectionItem<String?>(value: null, label: l10n.recurrenceNone),
      _SelectionItem<String?>(value: 'daily', label: l10n.recurrenceDaily),
      _SelectionItem<String?>(value: 'weekly', label: l10n.recurrenceWeekly),
      _SelectionItem<String?>(value: 'monthly', label: l10n.recurrenceMonthly),
      _SelectionItem<String?>(value: 'permanent', label: l10n.recurrencePermanent),
    ];

    final selected = await showModalBottomSheet<String?>(
      context: context,
      backgroundColor: Colors.transparent,
      builder: (context) => _buildSelectionBottomSheet<String?>(
        title: l10n.meetingRecurrence,
        items: recurrences,
        selectedValue: _recurrence,
      ),
    );

    if (selected != null || selected == null) {
      setState(() => _recurrence = selected);
    }
  }

  String _getTypeLabel(String type) {
    final l10n = AppLocalizations.of(context)!;
    switch (type) {
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

  String _getDurationLabel(int duration) {
    final l10n = AppLocalizations.of(context)!;
    switch (duration) {
      case 15:
        return l10n.duration15;
      case 30:
        return l10n.duration30;
      case 45:
        return l10n.duration45;
      case 60:
        return l10n.duration60;
      case 90:
        return l10n.duration90;
      case 120:
        return l10n.duration120;
      case 180:
        return l10n.duration180;
      default:
        return '$duration min';
    }
  }

  String _getRecurrenceLabel(String? recurrence) {
    final l10n = AppLocalizations.of(context)!;
    if (recurrence == null) return l10n.recurrenceNone;
    switch (recurrence) {
      case 'daily':
        return l10n.recurrenceDaily;
      case 'weekly':
        return l10n.recurrenceWeekly;
      case 'monthly':
        return l10n.recurrenceMonthly;
      case 'permanent':
        return l10n.recurrencePermanent;
      default:
        return recurrence;
    }
  }

  Future<void> _createMeeting() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final l10n = AppLocalizations.of(context)!;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      // Log all field values before creating the request
      Logger.logInfo('Creating meeting with the following data:');
      Logger.logInfo('  title: ${_titleController.text}');
      Logger.logInfo('  scheduledDate: $_scheduledDate');
      Logger.logInfo('  scheduledTime: $_scheduledTime');
      Logger.logInfo('  duration: $_duration');
      Logger.logInfo('  recurrence: $_recurrence');
      Logger.logInfo('  type: $_type');
      Logger.logInfo('  subjectId: $_subjectId');
      Logger.logInfo('  needsVideoRecord: $_needsVideoRecord');
      Logger.logInfo('  needsAudioRecord: $_needsAudioRecord');
      Logger.logInfo('  needsTranscription: $_needsTranscription');
      Logger.logInfo('  forceEndAtDuration: $_forceEndAtDuration');
      Logger.logInfo('  allowAnonymous: $_allowAnonymous');
      Logger.logInfo('  additionalNotes: ${_notesController.text}');
      Logger.logInfo('  participantIds: $_selectedParticipantIds');
      Logger.logInfo('  departmentIds: $_selectedDepartmentIds');
      Logger.logInfo('  speakerId: $_selectedSpeakerId');

      // Combine date and time
      final scheduledAt = DateTime(
        _scheduledDate.year,
        _scheduledDate.month,
        _scheduledDate.day,
        _scheduledTime.hour,
        _scheduledTime.minute,
      );

      Logger.logInfo('Combined scheduledAt: $scheduledAt');

      final request = CreateMeetingRequest(
        title: _titleController.text,
        scheduledAt: scheduledAt,
        duration: _duration,
        recurrence: _recurrence,
        type: _type,
        subjectId: _subjectId,
        needsVideoRecord: _needsVideoRecord,
        needsAudioRecord: _needsAudioRecord,
        needsTranscription: _needsTranscription,
        forceEndAtDuration: _forceEndAtDuration,
        allowAnonymous: _allowAnonymous,
        additionalNotes:
            _notesController.text.isEmpty ? null : _notesController.text,
        participantIds: _selectedParticipantIds.toList(),
        departmentIds: _selectedDepartmentIds.toList(),
        speakerId: _selectedSpeakerId,
      );

      Logger.logInfo('Request object created successfully');
      Logger.logInfo('Request JSON: ${jsonEncode(request.toJson())}');

      await _meetingsService.createMeeting(request);

      Logger.logInfo('Meeting created successfully');

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.meetingCreatedSuccess)),
        );
        Navigator.pop(context, true); // Return true to indicate success
      }
    } on ApiException catch (e, stackTrace) {
      Logger.logError('API Exception while creating meeting', error: e, stackTrace: stackTrace);
      debugPrint('API Exception: ${e.message}');
      debugPrint('Status Code: ${e.statusCode}');
      debugPrint('Stack trace: $stackTrace');

      if (mounted) {
        setState(() {
          _error = '${l10n.meetingCreatedError}: ${e.message}';
          _isLoading = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_error!)),
        );
      }
    } catch (e, stackTrace) {
      Logger.logError('Unexpected error while creating meeting', error: e, stackTrace: stackTrace);
      debugPrint('Error type: ${e.runtimeType}');
      debugPrint('Error: $e');
      debugPrint('Stack trace: $stackTrace');

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

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (_isLoadingData) {
      return Scaffold(
        backgroundColor: AppColors.surface,
        appBar: AppBar(
          title: Text(l10n.createMeeting),
          backgroundColor: Colors.white,
          foregroundColor: AppColors.textPrimary,
          elevation: 0,
          scrolledUnderElevation: 0,
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
          backgroundColor: Colors.white,
          foregroundColor: AppColors.textPrimary,
          elevation: 0,
          scrolledUnderElevation: 0,
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
        backgroundColor: Colors.white,
        foregroundColor: AppColors.textPrimary,
        elevation: 0,
        scrolledUnderElevation: 0,
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
            padding: const EdgeInsets.all(20),
            children: [
              // Single white card container matching web design
              Container(
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
                padding: const EdgeInsets.all(32),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Basic Information Section
                    _buildSectionTitle(
                      l10n.meetingDetailsTitle,
                      Icons.badge_rounded,
                    ),
                    const SizedBox(height: 24),
                    _buildTextField(
                      controller: _titleController,
                      label: '${l10n.meetingTitle} *',
                      hint: l10n.enterTitle,
                      icon: Icons.title_rounded,
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return l10n.pleaseEnterTitle;
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),
                    _buildModalSelectionField(
                      label: '${l10n.meetingSubject} ${l10n.optional}',
                      icon: Icons.category_rounded,
                      value: _subjectId != null
                          ? _subjects?.firstWhere((s) => s.id == _subjectId).name
                          : null,
                      hint: l10n.pleaseSelectSubject,
                      onTap: () => _selectSubject(),
                    ),
                    const SizedBox(height: 16),
                    _buildModalSelectionField(
                      label: l10n.meetingType,
                      icon: Icons.event_note_rounded,
                      value: _getTypeLabel(_type),
                      onTap: () => _selectType(),
                    ),
                    const SizedBox(height: 32),

                    // Schedule Section
                    _buildSectionTitle(
                      l10n.dateAndTime,
                      Icons.schedule_rounded,
                    ),
                    const SizedBox(height: 24),
                    Row(
                      children: [
                        Expanded(
                          child: _buildDateTimeButton(
                            onTap: _selectDate,
                            icon: Icons.calendar_today_rounded,
                            label: DateFormat.yMMMd(
                                    Localizations.localeOf(context).toString())
                                .format(_scheduledDate),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _buildDateTimeButton(
                            onTap: _selectTime,
                            icon: Icons.access_time_rounded,
                            label: _scheduledTime.format(context),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _buildModalSelectionField(
                      label: l10n.meetingDuration,
                      icon: Icons.timer_outlined,
                      value: _getDurationLabel(_duration),
                      onTap: () => _selectDuration(),
                    ),
                    const SizedBox(height: 16),
                    _buildModalSelectionField(
                      label: l10n.meetingRecurrence,
                      icon: Icons.repeat_rounded,
                      value: _getRecurrenceLabel(_recurrence),
                      onTap: () => _selectRecurrence(),
                    ),
                    const SizedBox(height: 32),

                    // Participants Section
                    _buildSectionTitle(
                      l10n.meetingParticipants,
                      Icons.group_work_rounded,
                    ),
                    const SizedBox(height: 24),
                    _buildSelectionTile(
                      icon: Icons.person_outline_rounded,
                      title: '${l10n.meetingSpeaker} ${l10n.optional}',
                      subtitle: _selectedSpeakerId != null
                          ? _getUserName(_selectedSpeakerId!)
                          : l10n.noSpeaker,
                      onTap: _selectSpeaker,
                    ),
                    const SizedBox(height: 12),
                    _buildSelectionTile(
                      icon: Icons.people_alt_rounded,
                      title: l10n.meetingParticipants,
                      subtitle:
                          l10n.participantsSelected(_selectedParticipantIds.length),
                      onTap: _selectParticipants,
                    ),
                    if (_selectedParticipantIds.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: _buildChipList(
                            _selectedParticipantIds.map(_getUserName).toList()),
                      ),
                    ],
                    const SizedBox(height: 12),
                    _buildSelectionTile(
                      icon: Icons.business_outlined,
                      title: l10n.meetingDepartments,
                      subtitle: l10n
                          .departmentsSelected(_selectedDepartmentIds.length),
                      onTap: _selectDepartments,
                    ),
                    if (_selectedDepartmentIds.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: _buildChipList(_selectedDepartmentIds
                            .map(_getDepartmentName)
                            .toList()),
                      ),
                    ],
                    const SizedBox(height: 32),

                    // Recording Options Section
                    _buildSectionTitle(
                      l10n.recordingOptions,
                      Icons.settings_input_component_rounded,
                    ),
                    const SizedBox(height: 16),
                    _buildSwitchTile(
                      title: l10n.meetingVideoRecord,
                      subtitle: _needsVideoRecord ? l10n.enabled : l10n.disabled,
                      value: _needsVideoRecord,
                      onChanged: (value) {
                        setState(() {
                          _needsVideoRecord = value;
                          // Auto-enable audio when video is enabled
                          if (value && !_needsAudioRecord) {
                            _needsAudioRecord = true;
                          }
                        });
                      },
                    ),
                    const SizedBox(height: 8),
                    _buildSwitchTile(
                      title: l10n.meetingAudioRecord,
                      subtitle: _needsVideoRecord
                          ? l10n.enabled
                          : (_needsAudioRecord ? l10n.enabled : l10n.disabled),
                      value: _needsAudioRecord,
                      onChanged: _needsVideoRecord
                          ? null // Disabled when video is enabled
                          : (value) => setState(() => _needsAudioRecord = value),
                    ),
                    const SizedBox(height: 8),
                    _buildSwitchTile(
                      title: l10n.meetingTranscription,
                      subtitle: _needsTranscription ? l10n.enabled : l10n.disabled,
                      value: _needsTranscription,
                      onChanged: (value) =>
                          setState(() => _needsTranscription = value),
                    ),
                    const SizedBox(height: 8),
                    _buildSwitchTile(
                      title: l10n.meetingForceEndAtDuration,
                      subtitle: _forceEndAtDuration ? l10n.enabled : l10n.disabled,
                      value: _forceEndAtDuration,
                      onChanged: (value) =>
                          setState(() => _forceEndAtDuration = value),
                    ),
                    const SizedBox(height: 8),
                    _buildSwitchTile(
                      title: l10n.meetingAllowAnonymous,
                      subtitle: _allowAnonymous ? l10n.enabled : l10n.disabled,
                      value: _allowAnonymous,
                      onChanged: (value) =>
                          setState(() => _allowAnonymous = value),
                    ),
                    const SizedBox(height: 32),

                    // Additional Notes Section
                    _buildSectionTitle(
                      l10n.meetingNotes,
                      Icons.notes_rounded,
                    ),
                    const SizedBox(height: 24),
                    _buildTextField(
                      controller: _notesController,
                      label: l10n.meetingNotes,
                      hint: l10n.enterNotes,
                      icon: Icons.description_outlined,
                      maxLines: 4,
                    ),
                    const SizedBox(height: 32),

                    // Submit Button
                    Container(
                      height: 56,
                      decoration: BoxDecoration(
                        gradient: const LinearGradient(
                          colors: [AppColors.primary500, AppColors.primary600],
                        ),
                        borderRadius: BorderRadius.circular(24),
                        boxShadow: [
                          BoxShadow(
                            color: AppColors.primary500.withValues(alpha: 0.3),
                            blurRadius: 12,
                            offset: const Offset(0, 4),
                          ),
                        ],
                      ),
                      child: ElevatedButton.icon(
                        onPressed: _isLoading ? null : _createMeeting,
                        icon: const Icon(Icons.check_circle_rounded),
                        label: Text(
                          l10n.createMeeting,
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.transparent,
                          foregroundColor: Colors.white,
                          shadowColor: Colors.transparent,
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(24),
                          ),
                        ),
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

  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: AppColors.primary50,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(icon, color: AppColors.primary600, size: 20),
        ),
        const SizedBox(width: 12),
        Text(
          title,
          style: const TextStyle(
            fontSize: 18,
            fontWeight: FontWeight.w600,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    String? hint,
    IconData? icon,
    int maxLines = 1,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppColors.textSecondary,
          ),
        ),
        const SizedBox(height: 8),
        TextFormField(
          controller: controller,
          maxLines: maxLines,
          decoration: InputDecoration(
            hintText: hint,
            prefixIcon:
                icon != null ? Icon(icon, color: AppColors.primary500) : null,
            filled: true,
            fillColor: AppColors.surfaceMuted,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14), // radius-lg
              borderSide: BorderSide(color: AppColors.border),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: BorderSide(color: AppColors.border),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: const BorderSide(color: AppColors.primary400, width: 2),
            ),
            errorBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(14),
              borderSide: const BorderSide(color: AppColors.danger),
            ),
            contentPadding: maxLines > 1
                ? const EdgeInsets.all(16)
                : const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
          ),
          validator: validator,
          enabled: !_isLoading,
        ),
      ],
    );
  }

  Widget _buildModalSelectionField({
    required String label,
    required IconData icon,
    required String? value,
    String? hint,
    required VoidCallback onTap,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: AppColors.textSecondary,
          ),
        ),
        const SizedBox(height: 8),
        Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: _isLoading ? null : onTap,
            borderRadius: BorderRadius.circular(14),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
              decoration: BoxDecoration(
                color: AppColors.surfaceMuted,
                borderRadius: BorderRadius.circular(14),
                border: Border.all(color: AppColors.border),
              ),
              child: Row(
                children: [
                  Icon(icon, color: AppColors.primary500, size: 20),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      value ?? hint ?? '',
                      style: TextStyle(
                        fontSize: 16,
                        color: value != null
                            ? AppColors.textPrimary
                            : AppColors.textTertiary,
                      ),
                    ),
                  ),
                  Icon(
                    Icons.keyboard_arrow_down_rounded,
                    color: _isLoading
                        ? AppColors.textTertiary
                        : AppColors.textSecondary,
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildSelectionBottomSheet<T>({
    required String title,
    required List<_SelectionItem<T>> items,
    required T? selectedValue,
  }) {
    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.only(
          topLeft: Radius.circular(24),
          topRight: Radius.circular(24),
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Handle bar
          Container(
            margin: const EdgeInsets.only(top: 12, bottom: 8),
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: AppColors.border,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          // Title
          Padding(
            padding: const EdgeInsets.fromLTRB(24, 16, 24, 8),
            child: Row(
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
          ),
          const Divider(height: 1),
          // Items list
          Flexible(
            child: ListView.builder(
              shrinkWrap: true,
              padding: const EdgeInsets.symmetric(vertical: 8),
              itemCount: items.length,
              itemBuilder: (context, index) {
                final item = items[index];
                final isSelected = item.value == selectedValue;
                return Material(
                  color: Colors.transparent,
                  child: InkWell(
                    onTap: () => Navigator.pop(context, item.value),
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24,
                        vertical: 16,
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              item.label,
                              style: TextStyle(
                                fontSize: 16,
                                fontWeight: isSelected
                                    ? FontWeight.w600
                                    : FontWeight.normal,
                                color: isSelected
                                    ? AppColors.primary600
                                    : AppColors.textPrimary,
                              ),
                            ),
                          ),
                          if (isSelected)
                            const Icon(
                              Icons.check_circle_rounded,
                              color: AppColors.primary600,
                              size: 24,
                            ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          ),
          // Bottom padding for safe area
          SizedBox(height: MediaQuery.of(context).padding.bottom + 8),
        ],
      ),
    );
  }

  Widget _buildDateTimeButton({
    required VoidCallback onTap,
    required IconData icon,
    required String label,
  }) {
    return OutlinedButton.icon(
      onPressed: _isLoading ? null : onTap,
      icon: Icon(icon, size: 18),
      label: Text(label),
      style: OutlinedButton.styleFrom(
        foregroundColor: AppColors.primary600,
        side: BorderSide(color: AppColors.border),
        padding: const EdgeInsets.symmetric(vertical: 18, horizontal: 16),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(14),
        ),
        backgroundColor: AppColors.surfaceMuted,
      ),
    );
  }

  Widget _buildSelectionTile({
    required IconData icon,
    required String title,
    required String subtitle,
    required VoidCallback onTap,
  }) {
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: _isLoading ? null : onTap,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          decoration: BoxDecoration(
            color: AppColors.surfaceMuted,
            borderRadius: BorderRadius.circular(14),
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
                      style: const TextStyle(
                        fontWeight: FontWeight.w600,
                        fontSize: 14,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      subtitle,
                      style: const TextStyle(
                        color: AppColors.textSecondary,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
              Icon(
                Icons.chevron_right_rounded,
                color: _isLoading ? AppColors.textTertiary : AppColors.textSecondary,
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSwitchTile({
    required String title,
    required String subtitle,
    required bool value,
    required ValueChanged<bool>? onChanged,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12),
      decoration: BoxDecoration(
        color: AppColors.surfaceMuted,
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: AppColors.border),
      ),
      child: SwitchListTile.adaptive(
        value: value,
        onChanged: _isLoading ? null : onChanged,
        title: Text(title),
        subtitle: Text(
          subtitle,
          style: const TextStyle(
            color: AppColors.textSecondary,
            fontSize: 12,
          ),
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
      chips.add(_buildChip('+$remaining'));
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
          fontSize: 12,
        ),
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
        style: const TextStyle(
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
                prefixIcon: const Icon(Icons.search_rounded,
                    color: AppColors.primary600),
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
                  borderSide:
                      const BorderSide(color: AppColors.primary500, width: 2),
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
        style: const TextStyle(
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
        style: const TextStyle(
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
                prefixIcon: const Icon(Icons.search_rounded,
                    color: AppColors.primary600),
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
                  borderSide:
                      const BorderSide(color: AppColors.primary500, width: 2),
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

// Helper class for selection items
class _SelectionItem<T> {
  final T value;
  final String label;

  _SelectionItem({required this.value, required this.label});
}
