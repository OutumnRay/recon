import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:permission_handler/permission_handler.dart';
import 'package:provider/provider.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'firebase_options.dart';
import 'screens/splash_screen.dart';
import 'services/locale_service.dart';
import 'services/fcm_service.dart';
import 'l10n/app_localizations.dart';
import 'theme/app_theme.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase with platform-specific options
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  // Set up background message handler
  FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);

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
          theme: AppTheme.lightTheme,
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
