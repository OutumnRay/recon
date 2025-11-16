import 'dart:convert';
import 'api_client.dart';
import '../models/meeting.dart';
import '../models/recording.dart';
import '../models/transcript.dart';
import '../utils/logger.dart';

class MeetingsService {
  final ApiClient _apiClient;

  MeetingsService(this._apiClient);

  /// Get list of meetings with pagination and optional filters
  Future<List<MeetingWithDetails>> getMeetings({
    int page = 1,
    int pageSize = 20,
    String? status,
    String? type,
    String? subjectId,
    DateTime? startDate,
    DateTime? endDate,
  }) async {
    try {
      // Build query parameters
      final queryParams = <String, String>{
        'page': page.toString(),
        'page_size': pageSize.toString(),
      };

      if (status != null) queryParams['status'] = status;
      if (type != null) queryParams['type'] = type;
      if (subjectId != null) queryParams['subject_id'] = subjectId;
      if (startDate != null) queryParams['start_date'] = startDate.toIso8601String();
      if (endDate != null) queryParams['end_date'] = endDate.toIso8601String();

      final queryString = queryParams.entries
          .map((e) => '${e.key}=${Uri.encodeComponent(e.value)}')
          .join('&');

      Logger.logInfo('Fetching meetings with filters', data: queryParams);

      final response = await _apiClient.get('/api/v1/meetings?$queryString');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final items = (data['items'] as List<dynamic>?) ?? [];
        Logger.logSuccess('Found ${items.length} meetings');
        return items
            .map((item) => MeetingWithDetails.fromJson(item as Map<String, dynamic>))
            .toList();
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch meetings', error: e);
      rethrow;
    }
  }

  /// Get a single meeting by ID
  Future<MeetingWithDetails> getMeeting(String meetingId) async {
    try {
      final response = await _apiClient.get('/api/v1/meetings/$meetingId');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return MeetingWithDetails.fromJson(data);
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Create a new meeting
  Future<MeetingWithDetails> createMeeting(CreateMeetingRequest request) async {
    try {
      final response = await _apiClient.post(
        '/api/v1/meetings',
        request.toJson(),
      );

      if (response.statusCode == 201) {
        final data = jsonDecode(response.body);
        return MeetingWithDetails.fromJson(data);
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Update an existing meeting
  Future<MeetingWithDetails> updateMeeting(
    String meetingId,
    Map<String, dynamic> updates,
  ) async {
    try {
      final response = await _apiClient.put(
        '/api/v1/meetings/$meetingId',
        updates,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return MeetingWithDetails.fromJson(data);
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Delete a meeting
  Future<void> deleteMeeting(String meetingId) async {
    try {
      final response = await _apiClient.delete('/api/v1/meetings/$meetingId');

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Get LiveKit token for joining a meeting
  Future<LiveKitToken> getLiveKitToken(String meetingId) async {
    try {
      final response = await _apiClient.get(
        '/api/v1/meetings/$meetingId/token',
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        return LiveKitToken(
          token: data['token'] as String,
          url: 'wss://video.recontext.online',
        );
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Get meeting subjects (categories)
  Future<List<MeetingSubject>> getMeetingSubjects() async {
    try {
      final response = await _apiClient.get('/api/v1/meeting-subjects');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final items = data['items'] as List<dynamic>;
        return items
            .map((item) => MeetingSubject.fromJson(item as Map<String, dynamic>))
            .toList();
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Start recording for a meeting
  Future<void> startRecording(String meetingId) async {
    try {
      Logger.logInfo('Starting recording for meeting', data: {'meetingId': meetingId});

      final response = await _apiClient.post(
        '/api/v1/meetings/$meetingId/recording/start',
        {},
      );

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw _apiClient.handleError(response);
      }

      Logger.logSuccess('Recording started successfully');
    } catch (e) {
      Logger.logError('Failed to start recording', error: e);
      rethrow;
    }
  }

  /// Stop recording for a meeting
  Future<void> stopRecording(String meetingId) async {
    try {
      Logger.logInfo('Stopping recording for meeting', data: {'meetingId': meetingId});

      final response = await _apiClient.post(
        '/api/v1/meetings/$meetingId/recording/stop',
        {},
      );

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw _apiClient.handleError(response);
      }

      Logger.logSuccess('Recording stopped successfully');
    } catch (e) {
      Logger.logError('Failed to stop recording', error: e);
      rethrow;
    }
  }

  /// Start transcription for a meeting
  Future<void> startTranscription(String meetingId) async {
    try {
      Logger.logInfo('Starting transcription for meeting', data: {'meetingId': meetingId});

      final response = await _apiClient.post(
        '/api/v1/meetings/$meetingId/transcription/start',
        {},
      );

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw _apiClient.handleError(response);
      }

      Logger.logSuccess('Transcription started successfully');
    } catch (e) {
      Logger.logError('Failed to start transcription', error: e);
      rethrow;
    }
  }

  /// Stop transcription for a meeting
  Future<void> stopTranscription(String meetingId) async {
    try {
      Logger.logInfo('Stopping transcription for meeting', data: {'meetingId': meetingId});

      final response = await _apiClient.post(
        '/api/v1/meetings/$meetingId/transcription/stop',
        {},
      );

      if (response.statusCode != 200 && response.statusCode != 204) {
        throw _apiClient.handleError(response);
      }

      Logger.logSuccess('Transcription stopped successfully');
    } catch (e) {
      Logger.logError('Failed to stop transcription', error: e);
      rethrow;
    }
  }

  /// Get recordings for a meeting
  Future<List<RoomRecording>> getMeetingRecordings(String meetingId) async {
    try {
      Logger.logInfo('Fetching recordings for meeting', data: {'meetingId': meetingId});

      final response = await _apiClient.get(
        '/api/v1/meetings/$meetingId/recordings',
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final recordings = (data as List<dynamic>)
            .map((item) => RoomRecording.fromJson(item as Map<String, dynamic>))
            .toList();

        Logger.logSuccess('Found ${recordings.length} recording(s)');
        return recordings;
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch recordings', error: e);
      rethrow;
    }
  }

  /// Get transcripts and memo for a room
  Future<RoomTranscripts> getRoomTranscripts(String roomSid) async {
    try {
      Logger.logInfo('Fetching transcripts for room', data: {'roomSid': roomSid});

      final response = await _apiClient.get(
        '/api/v1/rooms/$roomSid/transcripts',
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final transcripts = RoomTranscripts.fromJson(data as Map<String, dynamic>);

        Logger.logSuccess('Found ${transcripts.totalPhrases} phrase(s) across ${transcripts.tracks.length} track(s)');
        return transcripts;
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch transcripts', error: e);
      rethrow;
    }
  }
}

class LiveKitToken {
  final String token;
  final String url;

  LiveKitToken({
    required this.token,
    required this.url,
  });
}

class MeetingSubject {
  final String id;
  final String name;
  final String? description;
  final bool isActive;

  MeetingSubject({
    required this.id,
    required this.name,
    this.description,
    required this.isActive,
  });

  factory MeetingSubject.fromJson(Map<String, dynamic> json) {
    return MeetingSubject(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      isActive: json['is_active'] as bool? ?? true,
    );
  }
}
