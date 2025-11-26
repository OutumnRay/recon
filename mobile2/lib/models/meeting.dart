import '../utils/date_utils.dart';

class Meeting {
  final String id;
  final String title;
  final DateTime scheduledAt;
  final int duration;
  final String? recurrence;
  final String type;
  final String? subjectId;
  final String status;
  final bool needsRecord;
  final bool needsTranscription;
  final bool isRecording;
  final bool isTranscribing;
  final String? additionalNotes;
  final bool isPermanent;
  final bool allowAnonymous;
  final String? liveKitRoomId;
  final String? videoPlaylistUrl;
  final String? summaryEn;
  final String? summaryRu;
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
    this.subjectId,
    required this.status,
    required this.needsRecord,
    required this.needsTranscription,
    required this.isRecording,
    required this.isTranscribing,
    this.additionalNotes,
    required this.isPermanent,
    required this.allowAnonymous,
    this.liveKitRoomId,
    this.videoPlaylistUrl,
    this.summaryEn,
    this.summaryRu,
    required this.createdBy,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Meeting.fromJson(Map<String, dynamic> json) {
    return Meeting(
      id: json['id'] as String,
      title: json['title'] as String,
      scheduledAt: AppDateUtils.parseToLocal(json['scheduled_at'] as String),
      duration: json['duration'] as int,
      recurrence: json['recurrence'] as String?,
      type: json['type'] as String,
      subjectId: json['subject_id'] as String?,
      status: json['status'] as String,
      needsRecord: json['needs_record'] as bool? ?? false,
      needsTranscription: json['needs_transcription'] as bool? ?? false,
      isRecording: json['is_recording'] as bool? ?? false,
      isTranscribing: json['is_transcribing'] as bool? ?? false,
      additionalNotes: json['additional_notes'] as String?,
      isPermanent: json['is_permanent'] as bool? ?? false,
      allowAnonymous: json['allow_anonymous'] as bool? ?? false,
      liveKitRoomId: json['livekit_room_id'] as String?,
      videoPlaylistUrl: json['video_playlist_url'] as String?,
      summaryEn: json['summary_en'] as String?,
      summaryRu: json['summary_ru'] as String?,
      createdBy: json['created_by'] as String,
      createdAt: AppDateUtils.parseToLocal(json['created_at'] as String),
      updatedAt: AppDateUtils.parseToLocal(json['updated_at'] as String),
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
      'needs_record': needsRecord,
      'needs_transcription': needsTranscription,
      'is_recording': isRecording,
      'is_transcribing': isTranscribing,
      'additional_notes': additionalNotes,
      'is_permanent': isPermanent,
      'allow_anonymous': allowAnonymous,
      'livekit_room_id': liveKitRoomId,
      'video_playlist_url': videoPlaylistUrl,
      'summary_en': summaryEn,
      'summary_ru': summaryRu,
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
  final int activeParticipantsCount;
  final int anonymousGuestsCount;
  final int recordingsCount;
  final String? summaryStatus; // pending, processing, completed, failed
  final bool hasTranscriptions;
  final bool hasVideo;

  MeetingWithDetails({
    required super.id,
    required super.title,
    required super.scheduledAt,
    required super.duration,
    super.recurrence,
    required super.type,
    super.subjectId,
    required super.status,
    required super.needsRecord,
    required super.needsTranscription,
    required super.isRecording,
    required super.isTranscribing,
    super.additionalNotes,
    required super.isPermanent,
    required super.allowAnonymous,
    super.liveKitRoomId,
    super.videoPlaylistUrl,
    super.summaryEn,
    super.summaryRu,
    required super.createdBy,
    required super.createdAt,
    required super.updatedAt,
    this.subjectName,
    required this.participants,
    required this.departments,
    this.activeParticipantsCount = 0,
    this.anonymousGuestsCount = 0,
    this.recordingsCount = 0,
    this.summaryStatus,
    this.hasTranscriptions = false,
    this.hasVideo = false,
  });

  /// Проверяет, есть ли у встречи какой-либо контент для просмотра
  /// (записи, транскрипции, видео или summary)
  bool get hasAnyContent =>
      recordingsCount > 0 ||
      hasTranscriptions ||
      hasVideo ||
      summaryStatus == 'completed' ||
      (summaryEn != null && summaryEn!.isNotEmpty) ||
      (summaryRu != null && summaryRu!.isNotEmpty);

  factory MeetingWithDetails.fromJson(Map<String, dynamic> json) {
    return MeetingWithDetails(
      id: json['id'] as String,
      title: json['title'] as String,
      scheduledAt: AppDateUtils.parseToLocal(json['scheduled_at'] as String),
      duration: json['duration'] as int,
      recurrence: json['recurrence'] as String?,
      type: json['type'] as String,
      subjectId: json['subject_id'] as String?,
      status: json['status'] as String,
      needsRecord: json['needs_record'] as bool? ?? false,
      needsTranscription: json['needs_transcription'] as bool? ?? false,
      isRecording: json['is_recording'] as bool? ?? false,
      isTranscribing: json['is_transcribing'] as bool? ?? false,
      additionalNotes: json['additional_notes'] as String?,
      isPermanent: json['is_permanent'] as bool? ?? false,
      allowAnonymous: json['allow_anonymous'] as bool? ?? false,
      liveKitRoomId: json['livekit_room_id'] as String?,
      videoPlaylistUrl: json['video_playlist_url'] as String?,
      summaryEn: json['summary_en'] as String?,
      summaryRu: json['summary_ru'] as String?,
      createdBy: json['created_by'] as String,
      createdAt: AppDateUtils.parseToLocal(json['created_at'] as String),
      updatedAt: AppDateUtils.parseToLocal(json['updated_at'] as String),
      subjectName: json['subject_name'] as String?,
      participants: (json['participants'] as List<dynamic>?)
              ?.map((p) => MeetingParticipant.fromJson(p as Map<String, dynamic>))
              .toList() ??
          [],
      departments: (json['departments'] as List<dynamic>?)
              ?.map((d) => d as String)
              .toList() ??
          [],
      activeParticipantsCount: json['active_participants_count'] as int? ?? 0,
      anonymousGuestsCount: json['anonymous_guests_count'] as int? ?? 0,
      recordingsCount: json['recordings_count'] as int? ?? 0,
      summaryStatus: json['summary_status'] as String?,
      hasTranscriptions: json['has_transcriptions'] as bool? ?? false,
      hasVideo: json['has_video'] as bool? ?? false,
    );
  }
}

class MeetingParticipant {
  final String id;
  final String meetingId;
  final String userId;
  final String role;
  final String status;
  final DateTime? joinedAt;
  final DateTime? leftAt;
  final DateTime createdAt;
  final ParticipantUser? user;

  MeetingParticipant({
    required this.id,
    required this.meetingId,
    required this.userId,
    required this.role,
    required this.status,
    this.joinedAt,
    this.leftAt,
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
      joinedAt: json['joined_at'] != null
          ? AppDateUtils.parseToLocal(json['joined_at'] as String)
          : null,
      leftAt: json['left_at'] != null
          ? AppDateUtils.parseToLocal(json['left_at'] as String)
          : null,
      createdAt: AppDateUtils.parseToLocal(json['created_at'] as String),
      user: json['user'] != null
          ? ParticipantUser.fromJson(json['user'] as Map<String, dynamic>)
          : null,
    );
  }

  String get displayName {
    if (user != null) {
      return user!.displayName;
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
  final String? firstName;
  final String? lastName;
  final String? role;

  ParticipantUser({
    required this.id,
    required this.username,
    required this.email,
    this.firstName,
    this.lastName,
    this.role,
  });

  factory ParticipantUser.fromJson(Map<String, dynamic> json) {
    return ParticipantUser(
      id: json['id'] as String,
      username: json['username'] as String? ?? '',
      email: json['email'] as String? ?? '',
      firstName: json['first_name'] as String?,
      lastName: json['last_name'] as String?,
      role: json['role'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      if (firstName != null) 'first_name': firstName,
      if (lastName != null) 'last_name': lastName,
      if (role != null) 'role': role,
    };
  }

  String get displayName {
    if (firstName != null && lastName != null && firstName!.isNotEmpty && lastName!.isNotEmpty) {
      return '$firstName $lastName';
    }
    if (username.isNotEmpty) {
      return username;
    }
    return email;
  }
}

class CreateMeetingRequest {
  final String title;
  final DateTime scheduledAt;
  final int duration;
  final String? recurrence;
  final String type;
  final String? subjectId;
  final bool needsRecord;
  final bool needsTranscription;
  final bool allowAnonymous;
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
    this.subjectId,
    required this.needsRecord,
    required this.needsTranscription,
    required this.allowAnonymous,
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
    final formattedTime = '${utcTime.toIso8601String().split('.').first}Z';

    return {
      'title': title,
      'scheduled_at': formattedTime,
      'duration': duration,
      if (recurrence != null) 'recurrence': recurrence,
      'type': type,
      if (subjectId != null && subjectId!.isNotEmpty) 'subject_id': subjectId,
      'needs_record': needsRecord,
      'needs_transcription': needsTranscription,
      'is_permanent': recurrence == 'permanent',
      'allow_anonymous': allowAnonymous,
      if (additionalNotes != null) 'additional_notes': additionalNotes,
      'participant_ids': participantIds,
      'department_ids': departmentIds,
      if (speakerId != null) 'speaker_id': speakerId,
    };
  }
}
