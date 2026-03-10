import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/app_colors.dart';
import '../../game/presentation/exchange_prompt_widget.dart';
import '../providers/rewards_provider.dart';
import 'reward_card.dart';
import 'streak_widget.dart';

class RewardsScreen extends ConsumerStatefulWidget {
  const RewardsScreen({super.key});

  @override
  ConsumerState<RewardsScreen> createState() => _RewardsScreenState();
}

class _RewardsScreenState extends ConsumerState<RewardsScreen> {
  String? _claimingId;

  Future<void> _handleClaim(String rewardId) async {
    // Show confirmation dialog before claiming.
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Claim Reward'),
        content: const Text(
          'Are you sure you want to claim this reward? '
          'Tokens will be added to your XEX Exchange balance.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.of(ctx).pop(true),
            child: const Text('Claim'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _claimingId = rewardId);
    try {
      await claimReward(ref, rewardId);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Failed to claim reward: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _claimingId = null);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;

    final rewardsAsync = ref.watch(rewardsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Rewards'), centerTitle: true),
      body: rewardsAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, _) => Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.error_outline, size: 48, color: textSecondary),
                const SizedBox(height: 16),
                Text(
                  'Failed to load rewards',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  error.toString(),
                  style: Theme.of(context)
                      .textTheme
                      .bodyMedium
                      ?.copyWith(color: textSecondary),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 24),
                TextButton(
                  onPressed: () => ref.invalidate(rewardsProvider),
                  child: const Text('Try Again'),
                ),
              ],
            ),
          ),
        ),
        data: (rewards) {
          final streak = rewards.streak;

          return RefreshIndicator(
            onRefresh: () async => ref.invalidate(rewardsProvider),
            child: ListView(
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
              children: [
                // --- Streak section ---
                if (streak != null) ...[
                  Center(
                    child: StreakWidget(currentStreak: streak.currentStreak),
                  ),
                  const SizedBox(height: 8),
                  // Streak stats row.
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _StreakStat(
                          label: 'Longest', value: '${streak.longestStreak}'),
                      const SizedBox(width: 32),
                      _StreakStat(
                          label: 'Bonus Skips', value: '${streak.bonusSkips}'),
                      const SizedBox(width: 32),
                      _StreakStat(
                          label: 'Bonus Answers',
                          value: '${streak.bonusAnswers}'),
                    ],
                  ),
                  const SizedBox(height: 28),
                ],

                // --- Pending rewards section ---
                if (rewards.pending.isNotEmpty) ...[
                  Text(
                    'Pending Rewards',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          color: textPrimary,
                        ),
                  ),
                  const SizedBox(height: 12),
                  ...rewards.pending.map(
                    (r) => RewardCard(
                      reward: r,
                      isClaiming: _claimingId == r.id,
                      onClaim: () => _handleClaim(r.id),
                    ),
                  ),
                  const SizedBox(height: 24),
                ],

                // --- History section ---
                Text(
                  'Claimed History',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        color: textPrimary,
                      ),
                ),
                const SizedBox(height: 12),
                if (rewards.history.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 32),
                    child: Center(
                      child: Text(
                        'No claimed rewards yet',
                        style: Theme.of(context)
                            .textTheme
                            .bodyMedium
                            ?.copyWith(color: textSecondary),
                      ),
                    ),
                  )
                else
                  ...rewards.history.map(
                    (r) => RewardCard(reward: r),
                  ),

                // --- Exchange promotion ---
                const SizedBox(height: 8),
                const ExchangePromptWidget(
                  type: ExchangePromptType.reward,
                ),

                // Bottom padding for safe area.
                const SizedBox(height: 24),
              ],
            ),
          );
        },
      ),
    );
  }
}

/// Small stat label used in the streak stats row.
class _StreakStat extends StatelessWidget {
  final String label;
  final String value;

  const _StreakStat({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;

    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontFamily: 'monospace',
            fontSize: 18,
            fontWeight: FontWeight.w700,
            color: textPrimary,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          label,
          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: textSecondary,
              ),
        ),
      ],
    );
  }
}
