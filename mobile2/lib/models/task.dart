import 'package:json_annotation/json_annotation.dart';
import 'user.dart';

part 'task.g.dart';

/// Task status enum
enum TaskStatus {
  @JsonValue('pending')
  pending,
  @JsonValue('in_progress')
  inProgress,
  @JsonValue('completed')
  completed,
  @JsonValue('cancelled')
  cancelled,
}

/// Task priority enum
enum TaskPriority {
  @JsonValue('low')
  low,
  @JsonValue('medium')
  medium,
  @JsonValue('high')
  high,
  @JsonValue('urgent')
  urgent,
}

/// Task model with all details
@JsonSerializable()
class Task {
  final String id;
  @JsonKey(name: 'session_id')
  final String sessionId;
  @JsonKey(name: 'meeting_id')
  final String? meetingId;
  final String title;
  final String? description;
  final String? hint;
  @JsonKey(name: 'assigned_to')
  final String? assignedTo;
  @JsonKey(name: 'assigned_by')
  final String? assignedBy;
  final TaskStatus status;
  final TaskPriority priority;
  @JsonKey(name: 'due_date')
  final DateTime? dueDate;
  @JsonKey(name: 'extracted_by_ai')
  final bool extractedByAi;
  @JsonKey(name: 'ai_confidence')
  final double? aiConfidence;
  @JsonKey(name: 'source_segment')
  final String? sourceSegment;
  @JsonKey(name: 'completed_at')
  final DateTime? completedAt;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;
  @JsonKey(name: 'updated_at')
  final DateTime updatedAt;
  @JsonKey(name: 'assigned_to_user')
  final UserInfo? assignedToUser;
  @JsonKey(name: 'assigned_by_user')
  final UserInfo? assignedByUser;

  Task({
    required this.id,
    required this.sessionId,
    this.meetingId,
    required this.title,
    this.description,
    this.hint,
    this.assignedTo,
    this.assignedBy,
    required this.status,
    required this.priority,
    this.dueDate,
    this.extractedByAi = false,
    this.aiConfidence,
    this.sourceSegment,
    this.completedAt,
    required this.createdAt,
    required this.updatedAt,
    this.assignedToUser,
    this.assignedByUser,
  });

  factory Task.fromJson(Map<String, dynamic> json) => _$TaskFromJson(json);
  Map<String, dynamic> toJson() => _$TaskToJson(this);

  Task copyWith({
    String? id,
    String? sessionId,
    String? meetingId,
    String? title,
    String? description,
    String? hint,
    String? assignedTo,
    String? assignedBy,
    TaskStatus? status,
    TaskPriority? priority,
    DateTime? dueDate,
    bool? extractedByAi,
    double? aiConfidence,
    String? sourceSegment,
    DateTime? completedAt,
    DateTime? createdAt,
    DateTime? updatedAt,
    UserInfo? assignedToUser,
    UserInfo? assignedByUser,
  }) {
    return Task(
      id: id ?? this.id,
      sessionId: sessionId ?? this.sessionId,
      meetingId: meetingId ?? this.meetingId,
      title: title ?? this.title,
      description: description ?? this.description,
      hint: hint ?? this.hint,
      assignedTo: assignedTo ?? this.assignedTo,
      assignedBy: assignedBy ?? this.assignedBy,
      status: status ?? this.status,
      priority: priority ?? this.priority,
      dueDate: dueDate ?? this.dueDate,
      extractedByAi: extractedByAi ?? this.extractedByAi,
      aiConfidence: aiConfidence ?? this.aiConfidence,
      sourceSegment: sourceSegment ?? this.sourceSegment,
      completedAt: completedAt ?? this.completedAt,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
      assignedToUser: assignedToUser ?? this.assignedToUser,
      assignedByUser: assignedByUser ?? this.assignedByUser,
    );
  }
}

/// Request to update task status
@JsonSerializable()
class UpdateTaskStatusRequest {
  final String status;

  UpdateTaskStatusRequest({required this.status});

  factory UpdateTaskStatusRequest.fromJson(Map<String, dynamic> json) =>
      _$UpdateTaskStatusRequestFromJson(json);
  Map<String, dynamic> toJson() => _$UpdateTaskStatusRequestToJson(this);
}

/// Response with list of tasks
@JsonSerializable()
class TasksResponse {
  final List<Task> items;
  final int total;
  @JsonKey(name: 'page_size')
  final int pageSize;
  final int offset;

  TasksResponse({
    required this.items,
    required this.total,
    required this.pageSize,
    required this.offset,
  });

  factory TasksResponse.fromJson(Map<String, dynamic> json) =>
      _$TasksResponseFromJson(json);
  Map<String, dynamic> toJson() => _$TasksResponseToJson(this);
}
