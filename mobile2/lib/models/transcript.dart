import 'recording.dart';

/// Represents a single phrase in a transcription with timing information
class TranscriptionPhrase {
  final String text;
  final double startTime;
  final double endTime;
  final int phraseIndex;
  final String trackId;

  TranscriptionPhrase({
    required this.text,
    required this.startTime,
    required this.endTime,
    required this.phraseIndex,
    required this.trackId,
  });

  factory TranscriptionPhrase.fromJson(Map<String, dynamic> json) {
    return TranscriptionPhrase(
      text: json['text'] as String? ?? '',
      startTime: (json['start_time'] as num?)?.toDouble() ?? 0.0,
      endTime: (json['end_time'] as num?)?.toDouble() ?? 0.0,
      phraseIndex: json['phrase_index'] as int? ?? 0,
      trackId: json['track_id'] as String? ?? '',
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'text': text,
      'start_time': startTime,
      'end_time': endTime,
      'phrase_index': phraseIndex,
      'track_id': trackId,
    };
  }
}

/// Represents all transcription phrases for a single track/participant
class TrackTranscript {
  final String trackId;
  final String participantId;
  final ParticipantInfo? participant;
  final DateTime startedAt;
  final List<TranscriptionPhrase> transcriptionPhrases;

  TrackTranscript({
    required this.trackId,
    required this.participantId,
    this.participant,
    required this.startedAt,
    required this.transcriptionPhrases,
  });

  factory TrackTranscript.fromJson(Map<String, dynamic> json) {
    return TrackTranscript(
      trackId: json['track_id'] as String? ?? '',
      participantId: json['participant_id'] as String? ?? '',
      participant: json['participant'] != null
          ? ParticipantInfo.fromJson(json['participant'] as Map<String, dynamic>)
          : null,
      startedAt: json['started_at'] != null
          ? DateTime.parse(json['started_at'] as String)
          : DateTime.now(),
      transcriptionPhrases: (json['transcription_phrases'] as List<dynamic>?)
              ?.map((e) =>
                  TranscriptionPhrase.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'track_id': trackId,
      'participant_id': participantId,
      'participant': participant?.toJson(),
      'started_at': startedAt.toIso8601String(),
      'transcription_phrases':
          transcriptionPhrases.map((p) => p.toJson()).toList(),
    };
  }

  /// Get display name for the participant
  String getParticipantName() {
    if (participant == null) return 'Unknown';
    return participant!.displayName;
  }

  /// Get initials for the participant avatar
  String getParticipantInitials() {
    final name = getParticipantName();
    if (name.isEmpty || name == 'Unknown') return '??';

    final parts = name.split(' ');
    if (parts.length == 1) {
      return parts[0].substring(0, parts[0].length < 2 ? 1 : 2).toUpperCase();
    }

    return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
  }
}

/// Represents all transcripts for a room, including memo
class RoomTranscripts {
  final String roomSid;
  final List<TrackTranscript> tracks;
  final String? memo;       // English memo
  final String? memoRu;     // Russian memo
  final String? summaryStatus;  // Summary generation status: pending, processing, completed, failed
  final String? summaryError;   // Error message if summary generation failed

  RoomTranscripts({
    required this.roomSid,
    required this.tracks,
    this.memo,
    this.memoRu,
    this.summaryStatus,
    this.summaryError,
  });

  factory RoomTranscripts.fromJson(Map<String, dynamic> json) {
    return RoomTranscripts(
      roomSid: json['room_sid'] as String? ?? '',
      tracks: (json['tracks'] as List<dynamic>?)
              ?.map((e) => TrackTranscript.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      memo: json['memo'] as String?,
      memoRu: json['memo_ru'] as String?,
      summaryStatus: json['summary_status'] as String?,
      summaryError: json['summary_error'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'room_sid': roomSid,
      'tracks': tracks.map((t) => t.toJson()).toList(),
      'memo': memo,
      'memo_ru': memoRu,
      'summary_status': summaryStatus,
      'summary_error': summaryError,
    };
  }

  /// Get memo in the specified language
  String? getMemo(String languageCode) {
    if (languageCode.startsWith('ru')) {
      return memoRu ?? memo;
    }
    return memo;
  }

  /// Check if any transcripts are available
  bool get hasTranscripts => tracks.any((t) => t.transcriptionPhrases.isNotEmpty);

  /// Check if memo is available
  bool get hasMemo => memo != null || memoRu != null;

  /// Get total number of phrases across all tracks
  int get totalPhrases => tracks.fold(
      0, (sum, track) => sum + track.transcriptionPhrases.length);
}

/// Helper class for merged/sorted transcript phrases with speaker info
class MergedTranscriptPhrase {
  final TranscriptionPhrase phrase;
  final TrackTranscript track;
  final DateTime absoluteTimestamp;

  MergedTranscriptPhrase({
    required this.phrase,
    required this.track,
    required this.absoluteTimestamp,
  });

  String get speakerName => track.getParticipantName();
  String get speakerInitials => track.getParticipantInitials();
  String get text => phrase.text;
  double get startTime => phrase.startTime;
  double get endTime => phrase.endTime;
}

/// Extension methods for RoomTranscripts
extension RoomTranscriptsHelpers on RoomTranscripts {
  /// Get all phrases merged and sorted chronologically
  List<MergedTranscriptPhrase> getMergedPhrases() {
    final List<MergedTranscriptPhrase> merged = [];

    for (final track in tracks) {
      final trackStartTime = track.startedAt;

      for (final phrase in track.transcriptionPhrases) {
        final absoluteTimestamp = trackStartTime.add(
          Duration(milliseconds: (phrase.startTime * 1000).toInt()),
        );

        merged.add(MergedTranscriptPhrase(
          phrase: phrase,
          track: track,
          absoluteTimestamp: absoluteTimestamp,
        ));
      }
    }

    // Sort by absolute timestamp
    merged.sort((a, b) => a.absoluteTimestamp.compareTo(b.absoluteTimestamp));

    return merged;
  }
}
