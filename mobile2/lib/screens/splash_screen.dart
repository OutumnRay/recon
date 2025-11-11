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
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              Color(0xFF4DD0E1), // Светло-бирюзовый
              Color(0xFF26C6DA), // Бирюзовый
              Color(0xFF00ACC1), // Глубокий бирюзовый
            ],
          ),
        ),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Логотип приложения с тенью и скруглением
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(32),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.15),
                      blurRadius: 20,
                      offset: const Offset(0, 10),
                    ),
                  ],
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: Image.asset(
                    'assets/icon/app_icon.png',
                    width: 100,
                    height: 100,
                    fit: BoxFit.cover,
                  ),
                ),
              ),
              const SizedBox(height: 32),
              const Text(
                'Recontext',
                style: TextStyle(
                  fontSize: 38,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                  letterSpacing: 1.2,
                  shadows: [
                    Shadow(
                      color: Colors.black26,
                      offset: Offset(0, 2),
                      blurRadius: 4,
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Your conversations, reimagined',
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.white.withOpacity(0.9),
                  fontWeight: FontWeight.w300,
                  letterSpacing: 0.5,
                ),
              ),
              const SizedBox(height: 60),
              // Индикатор загрузки с контейнером
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(24),
                ),
                child: const SizedBox(
                  width: 32,
                  height: 32,
                  child: CircularProgressIndicator(
                    color: Colors.white,
                    strokeWidth: 3,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
