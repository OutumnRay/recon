import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'storage_service.dart';

class LocaleService extends ChangeNotifier {
  final StorageService _storageService = StorageService();

  Locale _locale = const Locale('en');

  Locale get locale => _locale;

  LocaleService() {
    _loadLocale();
  }

  Future<void> _loadLocale() async {
    // Check if user has a saved preference
    final languageCode = await _storageService.getLocale();

    if (languageCode != null) {
      // User has a saved preference, use it
      _locale = Locale(languageCode);
      notifyListeners();
    } else {
      // No saved preference, detect system language
      final systemLocale = _getSystemLocale();
      _locale = systemLocale;
      // Save the detected locale as the default
      await _storageService.saveLocale(systemLocale.languageCode);
      notifyListeners();
    }
  }

  /// Get system locale, fallback to English if not supported
  Locale _getSystemLocale() {
    final systemLocales = ui.PlatformDispatcher.instance.locales;

    // Check if any of the system locales are supported
    for (final systemLocale in systemLocales) {
      if (systemLocale.languageCode == 'ru') {
        return const Locale('ru');
      } else if (systemLocale.languageCode == 'en') {
        return const Locale('en');
      }
    }

    // Default to English if no supported locale found
    return const Locale('en');
  }

  Future<void> setLocale(Locale locale) async {
    _locale = locale;
    await _storageService.saveLocale(locale.languageCode);
    notifyListeners();
  }

  /// Load locale from user profile after login
  Future<void> loadUserLocale() async {
    final languageCode = await _storageService.getLocale();
    if (languageCode != null && languageCode != _locale.languageCode) {
      _locale = Locale(languageCode);
      notifyListeners();
    }
  }

  bool get isEnglish => _locale.languageCode == 'en';
  bool get isRussian => _locale.languageCode == 'ru';
}
