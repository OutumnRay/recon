import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../services/api_client.dart';
import '../services/meetings_service.dart';
import '../services/users_service.dart';
import '../models/meeting.dart';
import '../widgets/error_display.dart';
import '../l10n/app_localizations.dart';

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
        backgroundColor: const Color(0xFFF5F5F5),
        appBar: AppBar(
          title: Text(l10n.createMeeting),
          backgroundColor: Colors.white,
          foregroundColor: const Color(0xFF26C6DA),
          elevation: 1,
          shadowColor: Colors.black.withOpacity(0.1),
        ),
        body: const Center(
          child: CircularProgressIndicator(
            color: Color(0xFF26C6DA),
          ),
        ),
      );
    }

    if (_error != null && _subjects == null) {
      return Scaffold(
        backgroundColor: const Color(0xFFF5F5F5),
        appBar: AppBar(
          title: Text(l10n.createMeeting),
          backgroundColor: Colors.white,
          foregroundColor: const Color(0xFF26C6DA),
          elevation: 1,
          shadowColor: Colors.black.withOpacity(0.1),
        ),
        body: FullScreenError(
          error: _error!,
          onRetry: _loadInitialData,
          title: l10n.failedToLoadFormData,
        ),
      );
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF5F5F5),
      appBar: AppBar(
        title: Text(l10n.createMeeting),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF26C6DA),
        elevation: 1,
        shadowColor: Colors.black.withOpacity(0.1),
        actions: [
          if (_isLoading)
            const Center(
              child: Padding(
                padding: EdgeInsets.all(16.0),
                child: SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    color: Color(0xFF26C6DA),
                  ),
                ),
              ),
            ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            // Title
            TextFormField(
              controller: _titleController,
              decoration: InputDecoration(
                labelText: '${l10n.meetingTitle} ${l10n.required}',
                hintText: l10n.enterTitle,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.title, color: Color(0xFF26C6DA)),
              ),
              validator: (value) {
                if (value == null || value.isEmpty) {
                  return l10n.pleaseEnterTitle;
                }
                return null;
              },
              enabled: !_isLoading,
            ),
            const SizedBox(height: 16),

            // Subject
            DropdownButtonFormField<String>(
              initialValue: _subjectId,
              decoration: InputDecoration(
                labelText: '${l10n.meetingSubject} ${l10n.required}',
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.category, color: Color(0xFF26C6DA)),
              ),
              items: _subjects?.map((subject) {
                return DropdownMenuItem(
                  value: subject.id,
                  child: Text(subject.name),
                );
              }).toList(),
              onChanged: _isLoading
                  ? null
                  : (value) {
                      setState(() {
                        _subjectId = value;
                      });
                    },
              validator: (value) {
                if (value == null) {
                  return l10n.pleaseSelectSubject;
                }
                return null;
              },
            ),
            const SizedBox(height: 16),

            // Type
            DropdownButtonFormField<String>(
              initialValue: _type,
              decoration: InputDecoration(
                labelText: l10n.meetingType,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.event, color: Color(0xFF26C6DA)),
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
                      setState(() {
                        _type = value!;
                      });
                    },
            ),
            const SizedBox(height: 16),

            // Date and Time
            Builder(
              builder: (context) {
                final locale = Localizations.localeOf(context).toString();
                final dateFormat = DateFormat.yMMMd(locale);
                return Row(
                  children: [
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: _isLoading ? null : _selectDate,
                        icon: const Icon(Icons.calendar_today),
                        label: Text(dateFormat.format(_scheduledDate)),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: const Color(0xFF26C6DA),
                          side: BorderSide(color: Colors.grey.shade300, width: 1.5),
                          padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Expanded(
                      child: OutlinedButton.icon(
                        onPressed: _isLoading ? null : _selectTime,
                        icon: const Icon(Icons.access_time),
                        label: Text(_scheduledTime.format(context)),
                        style: OutlinedButton.styleFrom(
                          foregroundColor: const Color(0xFF26C6DA),
                          side: BorderSide(color: Colors.grey.shade300, width: 1.5),
                          padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16),
                          ),
                        ),
                      ),
                    ),
                  ],
                );
              },
            ),
            const SizedBox(height: 8),
            Text(
              _formatDateTime(context, _scheduledDate, _scheduledTime),
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey[600],
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),

            // Duration
            DropdownButtonFormField<int>(
              initialValue: _duration,
              decoration: InputDecoration(
                labelText: l10n.meetingDuration,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.timer, color: Color(0xFF26C6DA)),
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
                      setState(() {
                        _duration = value!;
                      });
                    },
            ),
            const SizedBox(height: 16),

            // Recurrence
            DropdownButtonFormField<String?>(
              initialValue: _recurrence,
              decoration: InputDecoration(
                labelText: l10n.meetingRecurrence,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.repeat, color: Color(0xFF26C6DA)),
              ),
              items: [
                DropdownMenuItem(value: null, child: Text(l10n.recurrenceNone)),
                DropdownMenuItem(value: 'daily', child: Text(l10n.recurrenceDaily)),
                DropdownMenuItem(value: 'weekly', child: Text(l10n.recurrenceWeekly)),
                DropdownMenuItem(value: 'monthly', child: Text(l10n.recurrenceMonthly)),
              ],
              onChanged: _isLoading
                  ? null
                  : (value) {
                      setState(() {
                        _recurrence = value;
                      });
                    },
            ),
            const SizedBox(height: 16),

            // Speaker
            Card(
              elevation: 0,
              color: const Color(0xFF26C6DA).withOpacity(0.08),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: const Color(0xFF26C6DA).withOpacity(0.2), width: 1),
              ),
              child: ListTile(
                title: Text(
                  '${l10n.meetingSpeaker} ${l10n.optional}',
                  style: const TextStyle(fontWeight: FontWeight.w500),
                ),
                subtitle: _selectedSpeakerId != null
                    ? Text(
                        _users?.firstWhere((u) => u.id == _selectedSpeakerId).displayName ??
                            'Unknown')
                    : Text(l10n.noSpeaker),
                leading: const Icon(Icons.person, color: Color(0xFF26C6DA)),
                trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                onTap: _isLoading ? null : _selectSpeaker,
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              ),
            ),
            const SizedBox(height: 8),

            // Participants
            Card(
              elevation: 0,
              color: const Color(0xFF26C6DA).withOpacity(0.08),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: const Color(0xFF26C6DA).withOpacity(0.2), width: 1),
              ),
              child: ListTile(
                title: Text(
                  l10n.meetingParticipants,
                  style: const TextStyle(fontWeight: FontWeight.w500),
                ),
                subtitle: Text(l10n.participantsSelected(_selectedParticipantIds.length)),
                leading: const Icon(Icons.people, color: Color(0xFF26C6DA)),
                trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                onTap: _isLoading ? null : _selectParticipants,
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              ),
            ),
            const SizedBox(height: 8),

            // Departments
            Card(
              elevation: 0,
              color: const Color(0xFF26C6DA).withOpacity(0.08),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(16),
                side: BorderSide(color: const Color(0xFF26C6DA).withOpacity(0.2), width: 1),
              ),
              child: ListTile(
                title: Text(
                  l10n.meetingDepartments,
                  style: const TextStyle(fontWeight: FontWeight.w500),
                ),
                subtitle: Text(l10n.departmentsSelected(_selectedDepartmentIds.length)),
                leading: const Icon(Icons.business, color: Color(0xFF26C6DA)),
                trailing: const Icon(Icons.chevron_right, color: Color(0xFF26C6DA)),
                onTap: _isLoading ? null : _selectDepartments,
                contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              ),
            ),
            const SizedBox(height: 16),

            // Recording options
            Card(
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
              elevation: 2,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        const Icon(Icons.videocam, color: Color(0xFF26C6DA)),
                        const SizedBox(width: 8),
                        Text(
                          l10n.recordingOptions,
                          style: Theme.of(context).textTheme.titleMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    SwitchListTile(
                      title: Text(l10n.meetingVideoRecord),
                      value: _needsVideoRecord,
                      activeColor: const Color(0xFF26C6DA),
                      onChanged: _isLoading
                          ? null
                          : (value) {
                              setState(() {
                                _needsVideoRecord = value;
                              });
                            },
                      contentPadding: EdgeInsets.zero,
                    ),
                    SwitchListTile(
                      title: Text(l10n.meetingAudioRecord),
                      value: _needsAudioRecord,
                      activeColor: const Color(0xFF26C6DA),
                      onChanged: _isLoading
                          ? null
                          : (value) {
                              setState(() {
                                _needsAudioRecord = value;
                              });
                            },
                      contentPadding: EdgeInsets.zero,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Additional notes
            TextFormField(
              controller: _notesController,
              decoration: InputDecoration(
                labelText: l10n.meetingNotes,
                hintText: l10n.enterNotes,
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
                prefixIcon: const Icon(Icons.notes, color: Color(0xFF26C6DA)),
                alignLabelWithHint: true,
              ),
              maxLines: 3,
              enabled: !_isLoading,
            ),
            const SizedBox(height: 24),

            // Create button
            FilledButton.icon(
              onPressed: _isLoading ? null : _createMeeting,
              icon: const Icon(Icons.add),
              label: Text(l10n.createMeeting),
              style: FilledButton.styleFrom(
                padding: const EdgeInsets.all(18),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                backgroundColor: const Color(0xFF26C6DA),
                elevation: 4,
              ),
            ),
          ],
        ),
      ),
    );
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
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      title: Text(
        l10n.selectParticipants,
        style: const TextStyle(color: Color(0xFF26C6DA), fontWeight: FontWeight.bold),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: InputDecoration(
                hintText: l10n.searchUsers,
                prefixIcon: const Icon(Icons.search, color: Color(0xFF26C6DA)),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
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
                    activeColor: const Color(0xFF26C6DA),
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
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selected),
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xFF26C6DA),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
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
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      title: Text(
        l10n.selectDepartments,
        style: const TextStyle(color: Color(0xFF26C6DA), fontWeight: FontWeight.bold),
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
              activeColor: const Color(0xFF26C6DA),
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
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selected),
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xFF26C6DA),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
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
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      title: Text(
        l10n.selectSpeaker,
        style: const TextStyle(color: Color(0xFF26C6DA), fontWeight: FontWeight.bold),
      ),
      content: SizedBox(
        width: double.maxFinite,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              decoration: InputDecoration(
                hintText: l10n.searchUsers,
                prefixIcon: const Icon(Icons.search, color: Color(0xFF26C6DA)),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: const BorderSide(color: Color(0xFF26C6DA), width: 2),
                ),
                filled: true,
                fillColor: Colors.white,
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
                activeColor: const Color(0xFF26C6DA),
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
                    activeColor: const Color(0xFF26C6DA),
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
          child: Text(l10n.cancel),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, _selectedId),
          style: FilledButton.styleFrom(
            backgroundColor: const Color(0xFF26C6DA),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          ),
          child: Text(l10n.select),
        ),
      ],
    );
  }
}
