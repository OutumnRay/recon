// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'task.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

Task _$TaskFromJson(Map<String, dynamic> json) => Task(
      id: json['id'] as String,
      sessionId: json['session_id'] as String,
      meetingId: json['meeting_id'] as String?,
      title: json['title'] as String,
      description: json['description'] as String?,
      hint: json['hint'] as String?,
      assignedTo: json['assigned_to'] as String?,
      assignedBy: json['assigned_by'] as String?,
      status: $enumDecode(_$TaskStatusEnumMap, json['status']),
      priority: $enumDecode(_$TaskPriorityEnumMap, json['priority']),
      dueDate: json['due_date'] == null
          ? null
          : DateTime.parse(json['due_date'] as String),
      extractedByAi: json['extracted_by_ai'] as bool? ?? false,
      aiConfidence: (json['ai_confidence'] as num?)?.toDouble(),
      sourceSegment: json['source_segment'] as String?,
      completedAt: json['completed_at'] == null
          ? null
          : DateTime.parse(json['completed_at'] as String),
      createdAt: DateTime.parse(json['created_at'] as String),
      updatedAt: DateTime.parse(json['updated_at'] as String),
      assignedToUser: json['assigned_to_user'] == null
          ? null
          : UserInfo.fromJson(json['assigned_to_user'] as Map<String, dynamic>),
      assignedByUser: json['assigned_by_user'] == null
          ? null
          : UserInfo.fromJson(json['assigned_by_user'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$TaskToJson(Task instance) {
  final val = <String, dynamic>{
    'id': instance.id,
    'session_id': instance.sessionId,
  };

  void writeNotNull(String key, dynamic value) {
    if (value != null) {
      val[key] = value;
    }
  }

  writeNotNull('meeting_id', instance.meetingId);
  val['title'] = instance.title;
  writeNotNull('description', instance.description);
  writeNotNull('hint', instance.hint);
  writeNotNull('assigned_to', instance.assignedTo);
  writeNotNull('assigned_by', instance.assignedBy);
  val['status'] = _$TaskStatusEnumMap[instance.status]!;
  val['priority'] = _$TaskPriorityEnumMap[instance.priority]!;
  writeNotNull('due_date', instance.dueDate?.toIso8601String());
  val['extracted_by_ai'] = instance.extractedByAi;
  writeNotNull('ai_confidence', instance.aiConfidence);
  writeNotNull('source_segment', instance.sourceSegment);
  writeNotNull('completed_at', instance.completedAt?.toIso8601String());
  val['created_at'] = instance.createdAt.toIso8601String();
  val['updated_at'] = instance.updatedAt.toIso8601String();
  writeNotNull('assigned_to_user', instance.assignedToUser?.toJson());
  writeNotNull('assigned_by_user', instance.assignedByUser?.toJson());
  return val;
}

const _$TaskStatusEnumMap = {
  TaskStatus.pending: 'pending',
  TaskStatus.inProgress: 'in_progress',
  TaskStatus.completed: 'completed',
  TaskStatus.cancelled: 'cancelled',
};

const _$TaskPriorityEnumMap = {
  TaskPriority.low: 'low',
  TaskPriority.medium: 'medium',
  TaskPriority.high: 'high',
  TaskPriority.urgent: 'urgent',
};

UpdateTaskStatusRequest _$UpdateTaskStatusRequestFromJson(
        Map<String, dynamic> json) =>
    UpdateTaskStatusRequest(
      status: json['status'] as String,
    );

Map<String, dynamic> _$UpdateTaskStatusRequestToJson(
        UpdateTaskStatusRequest instance) =>
    <String, dynamic>{
      'status': instance.status,
    };

TasksResponse _$TasksResponseFromJson(Map<String, dynamic> json) =>
    TasksResponse(
      items: (json['items'] as List<dynamic>)
          .map((e) => Task.fromJson(e as Map<String, dynamic>))
          .toList(),
      total: json['total'] as int,
      pageSize: json['page_size'] as int,
      offset: json['offset'] as int,
    );

Map<String, dynamic> _$TasksResponseToJson(TasksResponse instance) =>
    <String, dynamic>{
      'items': instance.items.map((e) => e.toJson()).toList(),
      'total': instance.total,
      'page_size': instance.pageSize,
      'offset': instance.offset,
    };
