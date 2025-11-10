import 'dart:convert';
import 'api_client.dart';
import 'storage_service.dart';
import '../models/user.dart';

class AuthService {
  final ApiClient _apiClient;
  final StorageService _storageService = StorageService();

  AuthService(this._apiClient);

  /// Login with username and password
  /// Returns User object if successful
  Future<User> login(String username, String password) async {
    try {
      final response = await _apiClient.post(
        '/api/v1/auth/login',
        {
          'username': username,
          'password': password,
        },
        requiresAuth: false,
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);

        // Save token
        await _storageService.saveToken(data['token']);

        // Parse user data
        final user = User.fromJson(data['user']);

        // Save user data
        await _storageService.saveUserData(
          userId: user.id,
          username: user.username,
          email: user.email,
          role: user.role,
        );

        return user;
      } else {
        throw _apiClient.handleError(response);
      }
    } catch (e) {
      rethrow;
    }
  }

  /// Logout - clear all stored data
  Future<void> logout() async {
    await _storageService.clearAll();
  }

  /// Check if user is logged in
  Future<bool> isLoggedIn() async {
    return await _storageService.isLoggedIn();
  }

  /// Get current user data from storage
  Future<Map<String, String?>> getCurrentUserData() async {
    return await _storageService.getUserData();
  }

  /// Get current auth token
  Future<String?> getToken() async {
    return await _storageService.getToken();
  }
}
