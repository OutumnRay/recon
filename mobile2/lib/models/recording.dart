class RoomRecording {
  final String id;
  final String roomSid;
  final String status;
  final DateTime startedAt;
  final DateTime? endedAt;
  final String? playlistUrl;
  final List<TrackRecording> tracks;

  RoomRecording({
    required this.id,
    required this.roomSid,
    required this.status,
    required this.startedAt,
    this.endedAt,
    this.playlistUrl,
    required this.tracks,
  });

  factory RoomRecording.fromJson(Map<String, dynamic> json) {
    return RoomRecording(
      id: json['id'] as String,
      roomSid: json['room_sid'] as String,
      status: json['status'] as String,
      startedAt: DateTime.parse(json['started_at'] as String),
      endedAt: json['ended_at'] != null
          ? DateTime.parse(json['ended_at'] as String)
          : null,
      playlistUrl: json['playlist_url'] as String?,
      tracks: (json['tracks'] as List<dynamic>?)
              ?.map((t) => TrackRecording.fromJson(t as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'room_sid': roomSid,
      'status': status,
      'started_at': startedAt.toIso8601String(),
      if (endedAt != null) 'ended_at': endedAt!.toIso8601String(),
      if (playlistUrl != null) 'playlist_url': playlistUrl,
      'tracks': tracks.map((t) => t.toJson()).toList(),
    };
  }

  Duration? get duration {
    if (endedAt != null) {
      return endedAt!.difference(startedAt);
    }
    return null;
  }

  bool get hasRoomRecording => playlistUrl != null && playlistUrl!.isNotEmpty;
}

class TrackRecording {
  final String id;
  final String status;
  final DateTime startedAt;
  final DateTime? endedAt;
  final String playlistUrl;
  final String participantId;
  final String trackId;
  final ParticipantInfo? participant;

  TrackRecording({
    required this.id,
    required this.status,
    required this.startedAt,
    this.endedAt,
    required this.playlistUrl,
    required this.participantId,
    required this.trackId,
    this.participant,
  });

  factory TrackRecording.fromJson(Map<String, dynamic> json) {
    return TrackRecording(
      id: json['id'] as String,
      status: json['status'] as String,
      startedAt: DateTime.parse(json['started_at'] as String),
      endedAt: json['ended_at'] != null
          ? DateTime.parse(json['ended_at'] as String)
          : null,
      playlistUrl: json['playlist_url'] as String,
      participantId: json['participant_id'] as String,
      trackId: json['track_id'] as String,
      participant: json['participant'] != null
          ? ParticipantInfo.fromJson(json['participant'] as Map<String, dynamic>)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'status': status,
      'started_at': startedAt.toIso8601String(),
      if (endedAt != null) 'ended_at': endedAt!.toIso8601String(),
      'playlist_url': playlistUrl,
      'participant_id': participantId,
      'track_id': trackId,
      if (participant != null) 'participant': participant!.toJson(),
    };
  }

  Duration? get duration {
    if (endedAt != null) {
      return endedAt!.difference(startedAt);
    }
    return null;
  }

  String get participantName {
    if (participant != null) {
      if (participant!.username.isNotEmpty) {
        return participant!.username;
      }
      if (participant!.email.isNotEmpty) {
        return participant!.email;
      }
    }
    return 'Unknown';
  }
}

class ParticipantInfo {
  final String id;
  final String username;
  final String email;
  final String? firstName;
  final String? lastName;
  final String? avatar;

  ParticipantInfo({
    required this.id,
    required this.username,
    required this.email,
    this.firstName,
    this.lastName,
    this.avatar,
  });

  factory ParticipantInfo.fromJson(Map<String, dynamic> json) {
    return ParticipantInfo(
      id: json['id'] as String,
      username: json['username'] as String? ?? '',
      email: json['email'] as String? ?? '',
      firstName: json['first_name'] as String?,
      lastName: json['last_name'] as String?,
      avatar: json['avatar'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'email': email,
      if (firstName != null) 'first_name': firstName,
      if (lastName != null) 'last_name': lastName,
      if (avatar != null) 'avatar': avatar,
    };
  }

  String get displayName {
    if (firstName != null && lastName != null) {
      return '$firstName $lastName';
    }
    if (username.isNotEmpty) {
      return username;
    }
    return email;
  }
}
