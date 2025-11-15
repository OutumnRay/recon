import 'package:flutter/material.dart';
import '../services/storage_service.dart';
import '../services/config_service.dart';
import '../services/api_client.dart';
import '../services/fcm_service.dart';
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
            await _registerPushNotifications(apiClient);
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

  Future<void> _registerPushNotifications(ApiClient apiClient) async {
    try {
      final fcmService = FCMService(apiClient);
      await fcmService.initialize();
      await fcmService.registerDevice();
    } catch (e) {
      debugPrint('Failed to register push notifications: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final logoSize = screenWidth / 4; // 1/4 ширины экрана

    return Scaffold(
      backgroundColor: Colors.white,
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Логотип - квадратный, 1/4 ширины экрана, без тени
            ClipRRect(
              borderRadius: BorderRadius.circular(16),
              child: Image.asset(
                'assets/icon/app_icon.png',
                width: logoSize,
                height: logoSize,
                fit: BoxFit.cover,
              ),
            ),
            const SizedBox(height: 32),
            const Text(
              'Recontext',
              style: TextStyle(
                fontSize: 38,
                fontWeight: FontWeight.bold,
                color: Color(0xFF26C6DA),
                letterSpacing: 1.2,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Your conversations, reimagined',
              style: TextStyle(
                fontSize: 16,
                color: Colors.grey[600],
                fontWeight: FontWeight.w300,
                letterSpacing: 0.5,
              ),
            ),
            const SizedBox(height: 60),
            // Индикатор загрузки
            const SizedBox(
              width: 32,
              height: 32,
              child: CircularProgressIndicator(
                color: Color(0xFF26C6DA),
                strokeWidth: 3,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
