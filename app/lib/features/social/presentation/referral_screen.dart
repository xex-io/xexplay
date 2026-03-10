import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../providers/social_provider.dart';

class ReferralScreen extends ConsumerWidget {
  const ReferralScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final surface =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;

    final statsAsync = ref.watch(referralStatsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Referrals'), centerTitle: true),
      body: statsAsync.when(
        loading: () => const LoadingWidget(),
        error: (error, _) => AppErrorWidget(
          message: error.toString(),
          onRetry: () => ref.invalidate(referralStatsProvider),
        ),
        data: (stats) => ListView(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
          children: [
            // --- Referral code card ---
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: surface,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                  color: primaryColor.withValues(alpha: 0.3),
                ),
              ),
              child: Column(
                children: [
                  Text(
                    'Your Referral Code',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          color: textSecondary,
                        ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    stats.code,
                    style: TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 28,
                      fontWeight: FontWeight.w700,
                      color: primaryColor,
                      letterSpacing: 4,
                    ),
                  ),
                  const SizedBox(height: 16),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      _ActionChip(
                        icon: Icons.copy,
                        label: 'Copy',
                        color: primaryColor,
                        surface: surface,
                        onTap: () {
                          Clipboard.setData(ClipboardData(text: stats.code));
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Referral code copied!'),
                              duration: Duration(seconds: 2),
                            ),
                          );
                        },
                      ),
                      const SizedBox(width: 12),
                      _ActionChip(
                        icon: Icons.share,
                        label: 'Share',
                        color: primaryColor,
                        surface: surface,
                        onTap: () {
                          Clipboard.setData(ClipboardData(
                            text:
                                'Join XEX Play with my code: ${stats.code}',
                          ));
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Share text copied!'),
                              duration: Duration(seconds: 2),
                            ),
                          );
                        },
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),

            // --- Stats ---
            Text(
              'Referral Stats',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: textPrimary,
                  ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                _StatCard(
                  label: 'Total Referrals',
                  value: '${stats.totalReferrals}',
                  surface: surface,
                  textPrimary: textPrimary,
                  textSecondary: textSecondary,
                ),
                const SizedBox(width: 10),
                _StatCard(
                  label: 'Completed',
                  value: '${stats.completedReferrals}',
                  surface: surface,
                  textPrimary: textPrimary,
                  textSecondary: textSecondary,
                ),
                const SizedBox(width: 10),
                _StatCard(
                  label: 'Rewards',
                  value: '${stats.rewards}',
                  surface: surface,
                  textPrimary: primaryColor,
                  textSecondary: textSecondary,
                ),
              ],
            ),
            const SizedBox(height: 32),

            // --- How it works ---
            Text(
              'How It Works',
              style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    color: textPrimary,
                  ),
            ),
            const SizedBox(height: 12),
            _StepTile(
              number: '1',
              text: 'Share your referral code with friends',
              surface: surface,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              primaryColor: primaryColor,
            ),
            const SizedBox(height: 8),
            _StepTile(
              number: '2',
              text: 'They sign up and play their first game',
              surface: surface,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              primaryColor: primaryColor,
            ),
            const SizedBox(height: 8),
            _StepTile(
              number: '3',
              text: 'You both earn bonus points!',
              surface: surface,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              primaryColor: primaryColor,
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Action chip (Copy / Share)
// ---------------------------------------------------------------------------

class _ActionChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final Color surface;
  final VoidCallback onTap;

  const _ActionChip({
    required this.icon,
    required this.label,
    required this.color,
    required this.surface,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(20),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 16, color: color),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: color,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Stat card
// ---------------------------------------------------------------------------

class _StatCard extends Expanded {
  _StatCard({
    required String label,
    required String value,
    required Color surface,
    required Color textPrimary,
    required Color textSecondary,
  }) : super(
          child: Container(
            padding: const EdgeInsets.all(14),
            decoration: BoxDecoration(
              color: surface,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              children: [
                Text(
                  value,
                  style: TextStyle(
                    fontFamily: 'monospace',
                    fontSize: 22,
                    fontWeight: FontWeight.w700,
                    color: textPrimary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  label,
                  style: TextStyle(
                    fontSize: 12,
                    color: textSecondary,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
        );
}

// ---------------------------------------------------------------------------
// Step tile
// ---------------------------------------------------------------------------

class _StepTile extends StatelessWidget {
  final String number;
  final String text;
  final Color surface;
  final Color textPrimary;
  final Color textSecondary;
  final Color primaryColor;

  const _StepTile({
    required this.number,
    required this.text,
    required this.surface,
    required this.textPrimary,
    required this.textSecondary,
    required this.primaryColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
      decoration: BoxDecoration(
        color: surface,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          CircleAvatar(
            radius: 14,
            backgroundColor: primaryColor.withValues(alpha: 0.15),
            child: Text(
              number,
              style: TextStyle(
                fontSize: 13,
                fontWeight: FontWeight.w700,
                color: primaryColor,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              text,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: textPrimary,
                  ),
            ),
          ),
        ],
      ),
    );
  }
}
