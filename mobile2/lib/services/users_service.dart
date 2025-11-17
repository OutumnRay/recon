import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'api_client.dart';
import 'storage_service.dart';
import '../models/user.dart';
import '../utils/logger.dart';

class UsersService {
  final ApiClient _apiClient;
  final StorageService _storageService = StorageService();

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

  /// Get user profile
  Future<User> getUserProfile(String userId) async {
    try {
      final response = await _apiClient.get('/api/v1/users/$userId');

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        Logger.logSuccess('Loaded user profile');
        return User.fromJson(data);
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to fetch user profile', error: e);
      rethrow;
    }
  }

  /// Update user profile
  Future<User> updateUserProfile(String userId, {
    String? firstName,
    String? lastName,
    String? phone,
    String? bio,
    String? language,
    String? notificationPreferences,
    String? avatar,
  }) async {
    try {
      final body = <String, dynamic>{};
      if (firstName != null) body['first_name'] = firstName;
      if (lastName != null) body['last_name'] = lastName;
      if (phone != null) body['phone'] = phone;
      if (bio != null) body['bio'] = bio;
      if (language != null) body['language'] = language;
      if (notificationPreferences != null) body['notification_preferences'] = notificationPreferences;
      if (avatar != null) body['avatar'] = avatar;

      final response = await _apiClient.put(
        '/api/v1/users/$userId',
        body,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        Logger.logSuccess('Profile updated successfully');
        return User.fromJson(data);
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      Logger.logError('Failed to update profile', error: e);
      rethrow;
    }
  }

  /// Upload avatar
  Future<String> uploadAvatar(String userId, File imageFile) async {
    try {
      final token = await _storageService.getToken();
      if (token == null) {
        throw Exception('No authentication token available');
      }

      final uri = Uri.parse('${_apiClient.baseUrl}/api/v1/users/$userId/avatar');
      final request = http.MultipartRequest('POST', uri);
      request.headers['Authorization'] = 'Bearer $token';

      final fileStream = http.ByteStream(imageFile.openRead());
      final fileLength = await imageFile.length();

      final multipartFile = http.MultipartFile(
        'avatar',
        fileStream,
        fileLength,
        filename: imageFile.path.split('/').last,
      );

      request.files.add(multipartFile);

      final streamedResponse = await request.send();
      final response = await http.Response.fromStream(streamedResponse);

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        final avatarUrl = data['avatar_url'] as String?;
        if (avatarUrl == null) {
          throw Exception('No avatar URL in response');
        }
        Logger.logSuccess('Avatar uploaded successfully');
        return avatarUrl;
      } else {
        final errorMessage = _apiClient.handleError(response);
        throw errorMessage;
      }
    } catch (e) {
      Logger.logError('Failed to upload avatar', error: e);
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
