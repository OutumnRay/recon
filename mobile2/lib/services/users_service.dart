import 'dart:convert';
import 'api_client.dart';
import '../utils/logger.dart';

class UsersService {
  final ApiClient _apiClient;

  UsersService(this._apiClient);

  /// Get list of users for participant selection
  Future<List<UserListItem>> getUsers({
    int page = 1,
    int pageSize = 100,
  }) async {
    try {
      final response = await _apiClient.get(
        '/api/v1/users?page=$page&page_size=$pageSize',
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final items = (data['items'] as List<dynamic>?) ?? [];
        Logger.logSuccess('Found ${items.length} users');
        return items
            .map((item) => UserListItem.fromJson(item as Map<String, dynamic>))
            .toList();
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch users', error: e);
      rethrow;
    }
  }

  /// Get list of departments
  Future<List<Department>> getDepartments() async {
    try {
      final response = await _apiClient.get('/api/v1/departments');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final items = (data['items'] as List<dynamic>?) ?? [];
        Logger.logSuccess('Found ${items.length} departments');
        return items
            .map((item) => Department.fromJson(item as Map<String, dynamic>))
            .toList();
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch departments', error: e);
      rethrow;
    }
  }
}

class UserListItem {
  final String id;
  final String username;
  final String email;
  final String? role;
  final bool isActive;

  UserListItem({
    required this.id,
    required this.username,
    required this.email,
    this.role,
    required this.isActive,
  });

  factory UserListItem.fromJson(Map<String, dynamic> json) {
    return UserListItem(
      id: json['id'] as String,
      username: json['username'] as String? ?? '',
      email: json['email'] as String? ?? '',
      role: json['role'] as String?,
      isActive: json['is_active'] as bool? ?? true,
    );
  }

  String get displayName {
    if (username.isNotEmpty) {
      return username;
    }
    return email;
  }
}

class Department {
  final String id;
  final String name;
  final String? description;
  final bool isActive;

  Department({
    required this.id,
    required this.name,
    this.description,
    required this.isActive,
  });

  factory Department.fromJson(Map<String, dynamic> json) {
    return Department(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String?,
      isActive: json['is_active'] as bool? ?? true,
    );
  }
}
