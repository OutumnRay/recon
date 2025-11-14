class Meeting {
  final String id;
  final String title;
  final DateTime scheduledAt;
  final int duration;
  final String? recurrence;
  final String type;
  final String subjectId;
  final String status;
  final bool needsVideoRecord;
  final bool needsAudioRecord;
  final bool isRecording;
  final bool isTranscribing;
  final String? additionalNotes;
  final String? liveKitRoomId;
  final String createdBy;
  final DateTime createdAt;
  final DateTime updatedAt;

  Meeting({
    required this.id,
    required this.title,
    required this.scheduledAt,
    required this.duration,
    this.recurrence,
    required this.type,
    required this.subjectId,
    required this.status,
    required this.needsVideoRecord,
    required this.needsAudioRecord,
    required this.isRecording,
    required this.isTranscribing,
    this.additionalNotes,
    this.liveKitRoomId,
    required this.createdBy,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Meeting.fromJson(Map<String, dynamic> json) {
    return Meeting(
      id: json['id'] as String,
      title: json['title'] as String,
      scheduledAt: DateTime.parse(json['scheduled_at'] as String),
      duration: json['duration'] as int,
      recurrence: json['recurrence'] as String?,
      type: json['type'] as String,
      subjectId: json['subject_id'] as String,
      status: json['status'] as String,
      needsVideoRecord: json['needs_video_record'] as bool? ?? false,
      needsAudioRecord: json['needs_audio_record'] as bool? ?? false,
      isRecording: json['is_recording'] as bool? ?? false,
      isTranscribing: json['is_transcribing'] as bool? ?? false,
      additionalNotes: json['additional_notes'] as String?,
      liveKitRoomId: json['livekit_room_id'] as String?,
      createdBy: json['created_by'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'scheduled_at': scheduledAt.toIso8601String(),
      'duration': duration,
      'recurrence': recurrence,
      'type': type,
      'subject_id': subjectId,
      'status': status,
      'needs_video_record': needsVideoRecord,
      'needs_audio_record': needsAudioRecord,
      'is_recording': isRecording,
      'is_transcribing': isTranscribing,
      'additional_notes': additionalNotes,
      'livekit_room_id': liveKitRoomId,
      'created_by': createdBy,
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }
}

class MeetingWithDetails extends Meeting {
  final String? subjectName;
  final List<MeetingParticipant> participants;
  final List<String> departments;

  MeetingWithDetails({
    required super.id,
    required super.title,
    required super.scheduledAt,
    required super.duration,
    super.recurrence,
    required super.type,
    required super.subjectId,
    required super.status,
    required super.needsVideoRecord,
    required super.needsAudioRecord,
    required super.isRecording,
    required super.isTranscribing,
    super.additionalNotes,
    super.liveKitRoomId,
    required super.createdBy,
    required super.createdAt,
    required super.updatedAt,
    this.subjectName,
    required this.participants,
    required this.departments,
  });

  factory MeetingWithDetails.fromJson(Map<String, dynamic> json) {
    return MeetingWithDetails(
      id: json['id'] as String,
      title: json['title'] as String,
      scheduledAt: DateTime.parse(json['scheduled_at'] as String),
      duration: json['duration'] as int,
      recurrence: json['recurrence'] as String?,
      type: json['type'] as String,
      subjectId: json['subject_id'] as String,
      status: json['status'] as String,
      needsVideoRecord: json['needs_video_record'] as bool? ?? false,
      needsAudioRecord: json['needs_audio_record'] as bool? ?? false,
      isRecording: json['is_recording'] as bool? ?? false,
      isTranscribing: json['is_transcribing'] as bool? ?? false,
      additionalNotes: json['additional_notes'] as String?,
      liveKitRoomId: json['livekit_room_id'] as String?,
      createdBy: json['created_by'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
      subjectName: json['subject_name'] as String?,
      participants: (json['participants'] as List<dynamic>?)
              ?.map((p) => MeetingParticipant.fromJson(p as Map<String, dynamic>))
              .toList() ??
          [],
      departments: (json['departments'] as List<dynamic>?)
              ?.map((d) => d as String)
              .toList() ??
          [],
    );
  }
}

class MeetingParticipant {
  final String id;
  final String meetingId;
  final String userId;
  final String role;
  final String status;
  final DateTime createdAt;
  final ParticipantUser? user;

  MeetingParticipant({
    required this.id,
    required this.meetingId,
    required this.userId,
    required this.role,
    required this.status,
    required this.createdAt,
    this.user,
  });

  factory MeetingParticipant.fromJson(Map<String, dynamic> json) {
    return MeetingParticipant(
      id: json['id'] as String,
      meetingId: json['meeting_id'] as String,
      userId: json['user_id'] as String,
      role: json['role'] as String,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      user: json['user'] != null
          ? ParticipantUser.fromJson(json['user'] as Map<String, dynamic>)
          : null,
    );
  }

  String get displayName {
    if (user != null) {
      if (user!.username.isNotEmpty) {
        return user!.username;
      }
      if (user!.email.isNotEmpty) {
        return user!.email;
      }
    }
    return userId;
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'meeting_id': meetingId,
      'user_id': userId,
      'role': role,
      'status': status,
      'created_at': createdAt.toIso8601String(),
      if (user != null) 'user': user!.toJson(),
    };
  }
}

class ParticipantUser {
  final String id;
  final String username;
  final String email;
  final String? role;

  ParticipantUser({
    required this.id,
    required this.username,
    required this.email,
    this.role,
  });

  factory ParticipantUser.fromJson(Map<String, dynamic> json) {
    return ParticipantUser(
      id: json['id'] as String,
      username: json['username'] as String? ?? '',
      email: json['email'] as String? ?? '',
      role: json['role'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      if (role != null) 'role': role,
    };
  }
}

class CreateMeetingRequest {
  final String title;
  final DateTime scheduledAt;
  final int duration;
  final String? recurrence;
  final String type;
  final String subjectId;
  final bool needsVideoRecord;
  final bool needsAudioRecord;
  final String? additionalNotes;
  final List<String> participantIds;
  final List<String> departmentIds;
  final String? speakerId;

  CreateMeetingRequest({
    required this.title,
    required this.scheduledAt,
    required this.duration,
    this.recurrence,
    required this.type,
    required this.subjectId,
    required this.needsVideoRecord,
    required this.needsAudioRecord,
    this.additionalNotes,
    required this.participantIds,
    required this.departmentIds,
    this.speakerId,
  });

  Map<String, dynamic> toJson() {
    // Convert to UTC and format as ISO 8601 with Z suffix
    // Backend expects format: "2025-01-15T10:00:00Z"
    final utcTime = scheduledAt.toUtc();
    // Remove milliseconds and ensure Z suffix
    final formattedTime = utcTime.toIso8601String().split('.').first + 'Z';

    return {
      'title': title,
      'scheduled_at': formattedTime,
      'duration': duration,
      if (recurrence != null) 'recurrence': recurrence,
      'type': type,
      'subject_id': subjectId,
      'needs_video_record': needsVideoRecord,
      'needs_audio_record': needsAudioRecord,
      if (additionalNotes != null) 'additional_notes': additionalNotes,
      'participant_ids': participantIds,
      'department_ids': departmentIds,
      if (speakerId != null) 'speaker_id': speakerId,
    };
  }
}
