import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/task.dart';
import '../models/user.dart';
import 'auth_service.dart';

class TaskService {
  final String baseUrl;
  final AuthService authService;

  TaskService({
    required this.baseUrl,
    required this.authService,
  });

  /// Get my tasks (assigned to current user)
  Future<List<Task>> getMyTasks({String? status}) async {
    final token = await authService.getToken();
    if (token == null) {
      throw Exception('Not authenticated');
    }

    var url = '$baseUrl/api/v1/my-tasks';
    if (status != null) {
      url += '?status=$status';
    }

    final response = await http.get(
      Uri.parse(url),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      final List<dynamic> jsonList = json.decode(response.body);
      return jsonList.map((json) => Task.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load tasks: ${response.body}');
    }
  }

  /// Get tasks for a specific meeting
  Future<List<Task>> getMeetingTasks(String meetingId) async {
    final token = await authService.getToken();
    if (token == null) {
      throw Exception('Not authenticated');
    }

    final response = await http.get(
      Uri.parse('$baseUrl/api/v1/meetings/$meetingId/tasks'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
    );

    if (response.statusCode == 200) {
      final List<dynamic> jsonList = json.decode(response.body);
      return jsonList.map((json) => Task.fromJson(json)).toList();
    } else {
      throw Exception('Failed to load meeting tasks: ${response.body}');
    }
  }

  /// Update task status
  Future<Task> updateTaskStatus(String taskId, TaskStatus newStatus) async {
    final token = await authService.getToken();
    if (token == null) {
      throw Exception('Not authenticated');
    }

    final statusString = _taskStatusToString(newStatus);
    final request = UpdateTaskStatusRequest(status: statusString);

    final response = await http.put(
      Uri.parse('$baseUrl/api/v1/tasks/$taskId/status'),
      headers: {
        'Authorization': 'Bearer $token',
        'Content-Type': 'application/json',
      },
      body: json.encode(request.toJson()),
    );

    if (response.statusCode == 200) {
      return Task.fromJson(json.decode(response.body));
    } else {
      throw Exception('Failed to update task status: ${response.body}');
    }
  }

  String _taskStatusToString(TaskStatus status) {
    switch (status) {
      case TaskStatus.pending:
        return 'pending';
      case TaskStatus.inProgress:
        return 'in_progress';
      case TaskStatus.completed:
        return 'completed';
      case TaskStatus.cancelled:
        return 'cancelled';
    }
  }
}
