import 'package:flutter/material.dart';

/// Shared color palette that mirrors the refreshed web design system.
class AppColors {
  const AppColors._();

  static const Color primary50 = Color(0xFFEFF5FF);
  static const Color primary100 = Color(0xFFDBE9FF);
  static const Color primary200 = Color(0xFFC0D7FF);
  static const Color primary300 = Color(0xFF9AC0FF);
  static const Color primary400 = Color(0xFF6BA3FF);
  static const Color primary500 = Color(0xFF3C82FF);
  static const Color primary600 = Color(0xFF2A63E5);
  static const Color primary700 = Color(0xFF224EC4);

  static const Color secondary50 = Color(0xFFF7F1FF);
  static const Color secondary400 = Color(0xFFB37CFF);
  static const Color secondary500 = Color(0xFFA05BFF);

  static const Color success = Color(0xFF1E9D7A);
  static const Color warning = Color(0xFFF59E0B);
  static const Color danger = Color(0xFFEF4444);
  static const Color info = Color(0xFF38BDF8);

  static const Color surface = Color(0xFFF5F7FB);
  static const Color surfaceCard = Colors.white;
  static const Color surfaceMuted = Color(0xFFF1F4FB);

  static const Color textPrimary = Color(0xFF1C2540);
  static const Color textSecondary = Color(0xFF49566D);
  static const Color textTertiary = Color(0xFF7A869F);

  static const Color border = Color(0xFFE3E8F4);
  static const Color borderStrong = Color(0xFFCBD5F5);

  static LinearGradient heroGradient = const LinearGradient(
    colors: [Color(0xFFEFF5FF), Color(0xFFFDFDFF)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
}
