import 'package:flutter/material.dart';

/// Coinbase-inspired color system for XEX Play.
/// Dark mode is the default.
class AppColors {
  AppColors._();

  // === Dark Mode ===
  static const darkBackground = Color(0xFF0A0B0D);
  static const darkSurface = Color(0xFF141519);
  static const darkSurfaceRaised = Color(0xFF1E1F25);
  static const darkPrimary = Color(0xFF587BFA);
  static const darkPrimaryBold = Color(0xFF0052FF);
  static const darkPositive = Color(0xFF09A85A);
  static const darkNegative = Color(0xFFCF202F);
  static const darkWarning = Color(0xFFED702F);
  static const darkTextPrimary = Color(0xFFFFFFFF);
  static const darkTextSecondary = Color(0xFF8A8F98);
  static const darkTextTertiary = Color(0xFF555961);
  static const darkBorder = Color(0xFF1E1F25);
  static const darkBorderSubtle = Color(0xFF141519);

  // === Light Mode ===
  static const lightBackground = Color(0xFFFFFFFF);
  static const lightSurface = Color(0xFFF5F8FF);
  static const lightSurfaceRaised = Color(0xFFEBEDF2);
  static const lightPrimary = Color(0xFF0052FF);
  static const lightPrimaryBold = Color(0xFF0052FF);
  static const lightPositive = Color(0xFF098551);
  static const lightNegative = Color(0xFFCF202F);
  static const lightWarning = Color(0xFFED702F);
  static const lightTextPrimary = Color(0xFF0A0B0D);
  static const lightTextSecondary = Color(0xFF5B616E);
  static const lightTextTertiary = Color(0xFF9DA3AE);
  static const lightBorder = Color(0xFFE2E4E9);
  static const lightBorderSubtle = Color(0xFFF0F1F3);

  // === Card Tier Colors ===
  static const goldStart = Color(0xFFFFD700);
  static const goldEnd = Color(0xFFFFA500);
  static const silverStart = Color(0xFFC0C0C0);
  static const silverEnd = Color(0xFF8A9BAE);
  static const whiteStart = Color(0xFFE2E4E9);
  static const whiteEnd = Color(0xFFFFFFFF);

  // === Card Tier Gradients ===
  static const goldGradient = LinearGradient(
    colors: [goldStart, goldEnd],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
  static const silverGradient = LinearGradient(
    colors: [silverStart, silverEnd],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
  static const whiteGradient = LinearGradient(
    colors: [whiteStart, whiteEnd],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );
}
