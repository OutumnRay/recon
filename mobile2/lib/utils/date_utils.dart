import 'package:intl/intl.dart';

/// Утилиты для работы с датами и временем
/// Все даты преобразуются в локальный часовой пояс пользователя
class AppDateUtils {
  /// Парсит дату из JSON и конвертирует в локальный часовой пояс
  /// Если дата приходит как UTC без указания зоны, добавляем 'Z' для корректного парсинга
  static DateTime parseToLocal(String dateString) {
    // Если строка уже содержит информацию о часовом поясе, парсим как есть
    if (dateString.endsWith('Z') || dateString.contains('+') || dateString.contains('T')) {
      return DateTime.parse(dateString).toLocal();
    }

    // Если это ISO 8601 без указания зоны, считаем что это UTC
    // Добавляем 'Z' для корректного парсинга как UTC
    return DateTime.parse('${dateString}Z').toLocal();
  }

  /// Форматирует DateTime для отображения пользователю
  /// Всегда использует локальный часовой пояс
  static String formatDateTime(DateTime dateTime, {String? locale}) {
    final localDateTime = dateTime.toLocal();
    final format = DateFormat('dd MMM yyyy, HH:mm', locale);
    return format.format(localDateTime);
  }

  /// Форматирует только дату
  static String formatDate(DateTime dateTime, {String? locale}) {
    final localDateTime = dateTime.toLocal();
    final format = DateFormat.yMMMd(locale);
    return format.format(localDateTime);
  }

  /// Форматирует только время
  static String formatTime(DateTime dateTime, {String? locale}) {
    final localDateTime = dateTime.toLocal();
    final format = DateFormat.Hm(locale);
    return format.format(localDateTime);
  }

  /// Форматирует дату в удобочитаемом формате (Сегодня, Завтра, и т.д.)
  /// Требует передачи локализованных строк для "Сегодня" и "Завтра"
  static String formatRelativeDateTime(
    DateTime dateTime, {
    String? locale,
    String? todayLabel,
    String? tomorrowLabel,
  }) {
    final localDateTime = dateTime.toLocal();
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final meetingDate = DateTime(localDateTime.year, localDateTime.month, localDateTime.day);

    if (meetingDate == today) {
      final label = todayLabel ?? 'Today';
      return '$label, ${formatTime(localDateTime, locale: locale)}';
    } else if (meetingDate == today.add(const Duration(days: 1))) {
      final label = tomorrowLabel ?? 'Tomorrow';
      return '$label, ${formatTime(localDateTime, locale: locale)}';
    } else if (meetingDate.isAfter(today) &&
        meetingDate.isBefore(today.add(const Duration(days: 7)))) {
      final weekdayFormat = DateFormat.E(locale);
      return '${weekdayFormat.format(localDateTime)}, ${formatTime(localDateTime, locale: locale)}';
    } else {
      return formatDateTime(localDateTime, locale: locale);
    }
  }

  /// Конвертирует локальную дату в UTC для отправки на сервер
  static String toUtcString(DateTime localDateTime) {
    return localDateTime.toUtc().toIso8601String();
  }

  /// Проверяет, является ли дата сегодняшней
  static bool isToday(DateTime dateTime) {
    final localDateTime = dateTime.toLocal();
    final now = DateTime.now();
    return localDateTime.year == now.year &&
        localDateTime.month == now.month &&
        localDateTime.day == now.day;
  }

  /// Проверяет, является ли дата прошедшей
  static bool isPast(DateTime dateTime) {
    return dateTime.toLocal().isBefore(DateTime.now());
  }

  /// Проверяет, является ли дата будущей
  static bool isFuture(DateTime dateTime) {
    return dateTime.toLocal().isAfter(DateTime.now());
  }
}
