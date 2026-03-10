import 'package:flutter/material.dart';
import '../../../core/constants/app_colors.dart';

enum CardTier { gold, silver, white }

class PredictionCardWidget extends StatelessWidget {
  const PredictionCardWidget({
    super.key,
    required this.tier,
    required this.question,
    required this.points,
  });

  final CardTier tier;
  final String question;
  final int points;

  LinearGradient get _borderGradient {
    switch (tier) {
      case CardTier.gold:
        return AppColors.goldGradient;
      case CardTier.silver:
        return AppColors.silverGradient;
      case CardTier.white:
        return AppColors.whiteGradient;
    }
  }

  String get _tierLabel {
    switch (tier) {
      case CardTier.gold:
        return 'Gold';
      case CardTier.silver:
        return 'Silver';
      case CardTier.white:
        return 'White';
    }
  }

  Color get _tierColor {
    switch (tier) {
      case CardTier.gold:
        return AppColors.goldStart;
      case CardTier.silver:
        return AppColors.silverStart;
      case CardTier.white:
        return AppColors.whiteEnd;
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surfaceColor =
        isDark ? AppColors.darkSurface : AppColors.lightSurface;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;

    return SizedBox(
      width: 300,
      height: 420,
      child: Container(
        decoration: BoxDecoration(
          gradient: _borderGradient,
          borderRadius: BorderRadius.circular(16),
        ),
        padding: const EdgeInsets.all(2),
        child: Container(
          decoration: BoxDecoration(
            color: surfaceColor,
            borderRadius: BorderRadius.circular(14),
          ),
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Tier badge
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                decoration: BoxDecoration(
                  color: _tierColor.withAlpha(30),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  _tierLabel,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: _tierColor,
                  ),
                ),
              ),

              // Question centered in remaining space
              Expanded(
                child: Center(
                  child: Text(
                    question,
                    style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                          color: textPrimary,
                        ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),

              // Points display
              Center(
                child: Text(
                  '$points pts',
                  style: TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: textSecondary,
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
