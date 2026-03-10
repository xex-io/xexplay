import 'package:flutter/material.dart';

import '../../../../core/constants/app_colors.dart';

/// The trader tier assigned to Exchange-active users.
enum TraderTier {
  /// Basic trader — has an Exchange account and some activity.
  trader,

  /// VIP trader — high-volume Exchange user with exclusive perks.
  vip,
}

/// A compact badge that indicates the user is an active Exchange trader.
///
/// When [exclusiveCardsAvailable] is true, a small dot indicator is shown
/// to signal that exclusive cards are unlocked for this user.
class TraderBadgeWidget extends StatelessWidget {
  const TraderBadgeWidget({
    super.key,
    required this.tier,
    this.exclusiveCardsAvailable = false,
  });

  final TraderTier tier;
  final bool exclusiveCardsAvailable;

  String get _label => switch (tier) {
        TraderTier.trader => 'Trader',
        TraderTier.vip => 'VIP',
      };

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final primary =
        isDark ? AppColors.darkPrimaryBold : AppColors.lightPrimaryBold;
    final goldColor = AppColors.goldStart;

    final Color badgeColor;
    final Color textColor;
    final LinearGradient? borderGradient;

    switch (tier) {
      case TraderTier.trader:
        badgeColor = primary.withAlpha(25);
        textColor = primary;
        borderGradient = null;
      case TraderTier.vip:
        badgeColor = goldColor.withAlpha(25);
        textColor = goldColor;
        borderGradient = AppColors.goldGradient;
    }

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
          decoration: BoxDecoration(
            color: badgeColor,
            borderRadius: BorderRadius.circular(8),
            border: borderGradient == null
                ? Border.all(color: textColor.withAlpha(60), width: 1)
                : null,
          ),
          foregroundDecoration: borderGradient != null
              ? BoxDecoration(
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: goldColor.withAlpha(80), width: 1),
                )
              : null,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                tier == TraderTier.vip
                    ? Icons.diamond_outlined
                    : Icons.show_chart_rounded,
                size: 14,
                color: textColor,
              ),
              const SizedBox(width: 4),
              Text(
                _label,
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: textColor,
                ),
              ),
            ],
          ),
        ),
        if (exclusiveCardsAvailable) ...[
          const SizedBox(width: 6),
          _ExclusiveCardIndicator(color: textColor),
        ],
      ],
    );
  }
}

/// A small pulsing dot that signals exclusive cards are available.
class _ExclusiveCardIndicator extends StatefulWidget {
  const _ExclusiveCardIndicator({required this.color});

  final Color color;

  @override
  State<_ExclusiveCardIndicator> createState() =>
      _ExclusiveCardIndicatorState();
}

class _ExclusiveCardIndicatorState extends State<_ExclusiveCardIndicator>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat(reverse: true);
    _animation = Tween<double>(begin: 0.4, end: 1.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOut),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        AnimatedBuilder(
          animation: _animation,
          builder: (context, child) {
            return Opacity(
              opacity: _animation.value,
              child: Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: widget.color,
                  shape: BoxShape.circle,
                ),
              ),
            );
          },
        ),
        const SizedBox(width: 4),
        Text(
          'Exclusive cards',
          style: TextStyle(
            fontSize: 11,
            fontWeight: FontWeight.w500,
            color: textSecondary,
          ),
        ),
      ],
    );
  }
}
