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
    final languageCode = await _storageService.getLocale();
    if (languageCode != null) {
      _locale = Locale(languageCode);
      notifyListeners();
    }
  }

  Future<void> setLocale(Locale locale) async {
    _locale = locale;
    await _storageService.saveLocale(locale.languageCode);
    notifyListeners();
  }

  bool get isEnglish => _locale.languageCode == 'en';
  bool get isRussian => _locale.languageCode == 'ru';
}
