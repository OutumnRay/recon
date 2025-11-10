import 'dart:convert';
import 'package:http/http.dart' as http;
import 'storage_service.dart';
import '../utils/logger.dart';

class ApiClient {
  final String baseUrl;
  final StorageService _storageService = StorageService();

  ApiClient({required this.baseUrl});

  Future<Map<String, String>> _getHeaders({bool requiresAuth = true}) async {
    final headers = {
      'Content-Type': 'application/json',
    };

    if (requiresAuth) {
      final token = await _storageService.getToken();
      if (token != null) {
        headers['Authorization'] = 'Bearer $token';
      }
    }

    return headers;
  }

  Future<http.Response> get(String endpoint, {bool requiresAuth = true}) async {
    final url = Uri.parse('$baseUrl$endpoint');
    final headers = await _getHeaders(requiresAuth: requiresAuth);

    Logger.logRequest('GET', url.toString());

    final response = await http.get(url, headers: headers);

    Logger.logResponse(url.toString(), response.statusCode, response.body);

    return response;
  }

  Future<http.Response> post(
    String endpoint,
    Map<String, dynamic> body, {
    bool requiresAuth = true,
  }) async {
    final url = Uri.parse('$baseUrl$endpoint');
    final headers = await _getHeaders(requiresAuth: requiresAuth);

    Logger.logRequest('POST', url.toString(), body: body);

    final response = await http.post(
      url,
      headers: headers,
      body: jsonEncode(body),
    );

    Logger.logResponse(url.toString(), response.statusCode, response.body);

    return response;
  }

  Future<http.Response> put(
    String endpoint,
    Map<String, dynamic> body, {
    bool requiresAuth = true,
  }) async {
    final url = Uri.parse('$baseUrl$endpoint');
    final headers = await _getHeaders(requiresAuth: requiresAuth);

    Logger.logRequest('PUT', url.toString(), body: body);

    final response = await http.put(
      url,
      headers: headers,
      body: jsonEncode(body),
    );

    Logger.logResponse(url.toString(), response.statusCode, response.body);

    return response;
  }

  Future<http.Response> delete(String endpoint, {bool requiresAuth = true}) async {
    final url = Uri.parse('$baseUrl$endpoint');
    final headers = await _getHeaders(requiresAuth: requiresAuth);

    Logger.logRequest('DELETE', url.toString());

    final response = await http.delete(url, headers: headers);

    Logger.logResponse(url.toString(), response.statusCode, response.body);

    return response;
  }

  // Helper method to handle API errors
  ApiException handleError(http.Response response) {
    try {
      final body = jsonDecode(response.body);
      return ApiException(
        statusCode: response.statusCode,
        message: body['error'] ?? body['message'] ?? 'Unknown error',
        detail: body['detail'],
      );
    } catch (e) {
      return ApiException(
        statusCode: response.statusCode,
        message: 'Failed to parse error response',
        detail: response.body,
      );
    }
  }
}

class ApiException implements Exception {
  final int statusCode;
  final String message;
  final String? detail;

  ApiException({
    required this.statusCode,
    required this.message,
    this.detail,
  });

  @override
  String toString() {
    if (detail != null) {
      return 'ApiException ($statusCode): $message - $detail';
    }
    return 'ApiException ($statusCode): $message';
  }
}
