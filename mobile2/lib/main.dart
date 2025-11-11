import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:provider/provider.dart';
import 'screens/splash_screen.dart';
import 'services/locale_service.dart';
import 'l10n/app_localizations.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(
    ChangeNotifierProvider(
      create: (_) => LocaleService(),
      child: const RecontextApp(),
    ),
  );
}

class RecontextApp extends StatefulWidget {
  const RecontextApp({super.key});

  @override
  State<RecontextApp> createState() => _RecontextAppState();
}

class _RecontextAppState extends State<RecontextApp> {
  @override
  void initState() {
    super.initState();
    _requestPermissions();
  }

  Future<void> _requestPermissions() async {
    // Запрашиваем доступ к камере и микрофону
    await [
      Permission.camera,
      Permission.microphone,
    ].request();

    // Проверяем, все ли даны
    if (await Permission.camera.isDenied ||
        await Permission.microphone.isDenied) {
      debugPrint("⚠️ Пользователь не дал разрешения.");
    } else {
      debugPrint("✅ Все разрешения предоставлены.");
    }
  }

  @override
  Widget build(BuildContext context) {
    return Consumer<LocaleService>(
      builder: (context, localeService, child) {
        return MaterialApp(
          title: 'Recontext',
          theme: ThemeData(
            colorScheme: ColorScheme.fromSeed(
              seedColor: const Color(0xFF46afba), // Seafoam color from design
            ),
            useMaterial3: true,
          ),
          locale: localeService.locale,
          localizationsDelegates: [
            ...AppLocalizations.localizationsDelegates,
            GlobalMaterialLocalizations.delegate,
            GlobalWidgetsLocalizations.delegate,
            GlobalCupertinoLocalizations.delegate,
          ],
          supportedLocales: AppLocalizations.supportedLocales,
          home: const SplashScreen(),
        );
      },
    );
  }
}
