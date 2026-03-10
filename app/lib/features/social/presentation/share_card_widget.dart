import 'dart:ui' as ui;

import 'package:flutter/material.dart';
import 'package:flutter/rendering.dart';

import '../../../core/constants/app_colors.dart';

/// The type of content being shared.
enum ShareType { score, streak, rank, achievement }

/// Data bag for the share card content.
class ShareData {
  final ShareType type;
  final String headline;
  final String value;
  final String? subtitle;

  const ShareData({
    required this.type,
    required this.headline,
    required this.value,
    this.subtitle,
  });
}

/// A branded share-card widget wrapped in a [RepaintBoundary] so it can
/// be captured as an image via [ShareCardController.captureImage].
class ShareCardWidget extends StatelessWidget {
  const ShareCardWidget({
    super.key,
    required this.data,
    required this.repaintKey,
  });

  final ShareData data;
  final GlobalKey repaintKey;

  @override
  Widget build(BuildContext context) {
    return RepaintBoundary(
      key: repaintKey,
      child: Container(
        width: 350,
        padding: const EdgeInsets.all(3),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(20),
          gradient: AppColors.goldGradient,
        ),
        child: Container(
          decoration: BoxDecoration(
            color: AppColors.darkBackground,
            borderRadius: BorderRadius.circular(18),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 28, vertical: 32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // XEX Play logo / title
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.play_circle_filled,
                    color: AppColors.darkPrimary,
                    size: 28,
                  ),
                  const SizedBox(width: 8),
                  const Text(
                    'XEX Play',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.w700,
                      color: AppColors.darkTextPrimary,
                      letterSpacing: 1.0,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 28),

              // Icon for share type
              Icon(
                _iconForType(data.type),
                size: 44,
                color: AppColors.goldStart,
              ),
              const SizedBox(height: 16),

              // Headline
              Text(
                data.headline,
                style: const TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: AppColors.darkTextSecondary,
                  letterSpacing: 0.8,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),

              // Main value
              Text(
                data.value,
                style: const TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 48,
                  fontWeight: FontWeight.w700,
                  color: AppColors.darkTextPrimary,
                ),
                textAlign: TextAlign.center,
              ),

              // Optional subtitle
              if (data.subtitle != null) ...[
                const SizedBox(height: 6),
                Text(
                  data.subtitle!,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.darkTextSecondary,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],

              const SizedBox(height: 28),

              // CTA
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 10,
                ),
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(24),
                  gradient: AppColors.goldGradient,
                ),
                child: const Text(
                  'Play on XEX Play',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: AppColors.darkBackground,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  IconData _iconForType(ShareType type) {
    return switch (type) {
      ShareType.score => Icons.stars_rounded,
      ShareType.streak => Icons.local_fire_department,
      ShareType.rank => Icons.leaderboard,
      ShareType.achievement => Icons.emoji_events,
    };
  }
}

/// Utility to capture the [ShareCardWidget] as an image.
class ShareCardController {
  /// Captures the [RepaintBoundary] identified by [key] as a PNG image.
  static Future<ui.Image?> captureImage(GlobalKey key,
      {double pixelRatio = 3.0}) async {
    final boundary =
        key.currentContext?.findRenderObject() as RenderRepaintBoundary?;
    if (boundary == null) return null;
    return boundary.toImage(pixelRatio: pixelRatio);
  }
}
