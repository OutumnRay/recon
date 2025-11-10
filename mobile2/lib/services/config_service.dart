import 'package:shared_preferences/shared_preferences.dart';

class ConfigService {
  static const String _apiUrlKey = 'api_base_url';
  static const String _defaultApiUrl = 'https://portal.recontext.online';

  Future<void> saveApiUrl(String url) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_apiUrlKey, url);
  }

  Future<String> getApiUrl() async {
    final prefs = await SharedPreferences.getInstance();
    return prefs.getString(_apiUrlKey) ?? _defaultApiUrl;
  }

  Future<void> resetToDefault() async {
    await saveApiUrl(_defaultApiUrl);
  }

  String getDefaultApiUrl() => _defaultApiUrl;
}
