import 'package:flutter/material.dart';
import '../services/storage_service.dart';
import '../services/config_service.dart';
import '../services/api_client.dart';
import 'login_screen.dart';
import 'main_screen.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  final _storageService = StorageService();
  final _configService = ConfigService();

  @override
  void initState() {
    super.initState();
    _checkAuthStatus();
  }

  Future<void> _checkAuthStatus() async {
    // Небольшая задержка для показа splash screen
    await Future.delayed(const Duration(milliseconds: 500));

    try {
      // Проверяем наличие сохраненного токена
      final token = await _storageService.getToken();

      if (!mounted) return;

      if (token != null && token.isNotEmpty) {
        // Токен есть - проверяем его валидность
        final apiUrl = await _configService.getApiUrl();
        final apiClient = ApiClient(baseUrl: apiUrl);

        try {
          // Пробуем сделать запрос к API для проверки токена
          final response = await apiClient.get('/api/v1/meetings', requiresAuth: true);

          if (!mounted) return;

          if (response.statusCode == 200) {
            // Токен валиден - переходим на главную страницу
            Navigator.of(context).pushReplacement(
              MaterialPageRoute(
                builder: (context) => MainScreen(apiUrl: apiUrl),
              ),
            );
            return;
          }
        } catch (e) {
          // Токен невалиден или истек - переходим на login
          debugPrint('Token validation failed: $e');
        }
      }

      // Токена нет или он невалиден - показываем экран логина
      if (!mounted) return;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (context) => const LoginScreen(),
        ),
      );
    } catch (e) {
      debugPrint('Error checking auth status: $e');

      // В случае ошибки показываем экран логина
      if (!mounted) return;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(
          builder: (context) => const LoginScreen(),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF46afba),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Логотип или название приложения
            const Icon(
              Icons.video_library,
              size: 80,
              color: Colors.white,
            ),
            const SizedBox(height: 24),
            const Text(
              'Recontext',
              style: TextStyle(
                fontSize: 32,
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
            const SizedBox(height: 48),
            // Индикатор загрузки
            const CircularProgressIndicator(
              color: Colors.white,
            ),
          ],
        ),
      ),
    );
  }
}
