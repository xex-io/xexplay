import 'package:flutter/material.dart';
import '../../../core/constants/app_colors.dart';
import '../data/reward_models.dart';

/// A card widget displaying a single reward item.
class RewardCard extends StatelessWidget {
  final RewardItem reward;
  final VoidCallback? onClaim;
  final bool isClaiming;

  const RewardCard({
    super.key,
    required this.reward,
    this.onClaim,
    this.isClaiming = false,
  });

  IconData _rewardTypeIcon() {
    return switch (reward.rewardType) {
      'token' => Icons.monetization_on_outlined,
      'skip' => Icons.fast_forward_outlined,
      'answer' => Icons.check_circle_outline,
      _ => Icons.card_giftcard_outlined,
    };
  }

  String _rewardTypeLabel() {
    return switch (reward.rewardType) {
      'token' => 'XEX Tokens',
      'skip' => 'Bonus Skip',
      'answer' => 'Bonus Answer',
      _ => reward.rewardType,
    };
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final positiveColor =
        isDark ? AppColors.darkPositive : AppColors.lightPositive;
    final warningColor =
        isDark ? AppColors.darkWarning : AppColors.lightWarning;
    final primaryBold =
        isDark ? AppColors.darkPrimaryBold : AppColors.lightPrimaryBold;

    final statusColor = reward.isPending ? warningColor : positiveColor;
    final statusLabel = reward.isPending ? 'Pending' : 'Claimed';

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: surface,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          // Reward type icon.
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(
              color: statusColor.withAlpha(30),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(
              _rewardTypeIcon(),
              color: statusColor,
              size: 24,
            ),
          ),
          const SizedBox(width: 12),

          // Reward info.
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  _rewardTypeLabel(),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                        color: textPrimary,
                      ),
                ),
                const SizedBox(height: 2),
                Text(
                  '${reward.periodType} - ${reward.periodKey}',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: textSecondary,
                      ),
                ),
              ],
            ),
          ),

          // Amount.
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                reward.amount % 1 == 0
                    ? reward.amount.toInt().toString()
                    : reward.amount.toStringAsFixed(2),
                style: TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 20,
                  fontWeight: FontWeight.w700,
                  color: textPrimary,
                ),
              ),
              const SizedBox(height: 4),
              if (reward.isPending)
                GestureDetector(
                  onTap: isClaiming ? null : onClaim,
                  child: Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                    decoration: BoxDecoration(
                      color: isClaiming
                          ? primaryBold.withAlpha(128)
                          : primaryBold,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: isClaiming
                        ? const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Text(
                            'Claim',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                  ),
                )
              else
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withAlpha(30),
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Text(
                    statusLabel,
                    style: TextStyle(
                      color: statusColor,
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
