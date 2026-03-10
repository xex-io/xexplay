import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/app_colors.dart';
import '../../auth/domain/auth_state.dart';
import '../../auth/providers/auth_provider.dart';
import '../../rewards/presentation/streak_widget.dart';
import '../../rewards/providers/rewards_provider.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final surface =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final negativeColor =
        isDark ? AppColors.darkNegative : AppColors.lightNegative;

    final authState = ref.watch(authProvider);
    final rewardsAsync = ref.watch(rewardsProvider);

    // Extract user from auth state.
    final user = authState is AuthAuthenticated ? authState.user : null;

    return Scaffold(
      appBar: AppBar(title: const Text('Profile'), centerTitle: true),
      body: ListView(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        children: [
          // --- User info section ---
          Center(
            child: Column(
              children: [
                CircleAvatar(
                  radius: 40,
                  backgroundColor: surface,
                  backgroundImage: user?.avatarUrl != null
                      ? NetworkImage(user!.avatarUrl!)
                      : null,
                  child: user?.avatarUrl == null
                      ? Icon(Icons.person, size: 40, color: textSecondary)
                      : null,
                ),
                const SizedBox(height: 12),
                Text(
                  user?.displayName ?? 'Player',
                  style: Theme.of(context).textTheme.titleLarge?.copyWith(
                        color: textPrimary,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  user?.email ?? '',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: textSecondary,
                      ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 28),

          // --- Streak widget ---
          Center(
            child: rewardsAsync.when(
              loading: () => const SizedBox(
                height: 140,
                child: Center(child: CircularProgressIndicator()),
              ),
              error: (_, _) => StreakWidget(
                currentStreak: user?.currentStreak ?? 0,
                size: 120,
              ),
              data: (rewards) => StreakWidget(
                currentStreak:
                    rewards.streak?.currentStreak ?? user?.currentStreak ?? 0,
                size: 120,
              ),
            ),
          ),
          const SizedBox(height: 28),

          // --- Stats grid ---
          Text(
            'Stats',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: textPrimary,
                ),
          ),
          const SizedBox(height: 12),
          _StatsGrid(
            stats: [
              _StatItem(
                  label: 'Total Points',
                  value: '${user?.totalPoints ?? 0}'),
              _StatItem(label: 'Games Played', value: '--'),
              _StatItem(label: 'Correct Predictions', value: '--'),
              _StatItem(label: 'Win Rate', value: '--'),
            ],
            surface: surface,
            textPrimary: textPrimary,
            textSecondary: textSecondary,
          ),
          const SizedBox(height: 28),

          // --- Social section ---
          Text(
            'Social',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: textPrimary,
                ),
          ),
          const SizedBox(height: 12),
          _SettingsTile(
            icon: Icons.card_giftcard_outlined,
            label: 'Referrals',
            trailing: null,
            surface: surface,
            textPrimary: textPrimary,
            onTap: () => context.push('/referral'),
          ),
          const SizedBox(height: 8),
          _SettingsTile(
            icon: Icons.emoji_events_outlined,
            label: 'Achievements',
            trailing: null,
            surface: surface,
            textPrimary: textPrimary,
            onTap: () => context.push('/achievements'),
          ),
          const SizedBox(height: 8),
          _SettingsTile(
            icon: Icons.groups_outlined,
            label: 'Mini-Leagues',
            trailing: null,
            surface: surface,
            textPrimary: textPrimary,
            onTap: () => context.push('/leagues'),
          ),
          const SizedBox(height: 28),

          // --- Settings section ---
          Text(
            'Settings',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: textPrimary,
                ),
          ),
          const SizedBox(height: 12),
          _SettingsTile(
            icon: Icons.language_outlined,
            label: 'Language',
            trailing: Text(
              'English',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: textSecondary,
                  ),
            ),
            surface: surface,
            textPrimary: textPrimary,
            onTap: () {
              // TODO: Navigate to language settings.
            },
          ),
          const SizedBox(height: 8),
          _SettingsTile(
            icon: Icons.dark_mode_outlined,
            label: 'Theme',
            trailing: Text(
              'Dark',
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: textSecondary,
                  ),
            ),
            surface: surface,
            textPrimary: textPrimary,
            onTap: () {
              // TODO: Navigate to theme settings.
            },
          ),
          const SizedBox(height: 8),
          _SettingsTile(
            icon: Icons.logout,
            label: 'Log Out',
            trailing: null,
            surface: surface,
            textPrimary: negativeColor,
            onTap: () {
              ref.read(authProvider.notifier).logout();
            },
          ),
          const SizedBox(height: 32),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Stats grid
// ---------------------------------------------------------------------------

class _StatItem {
  final String label;
  final String value;
  const _StatItem({required this.label, required this.value});
}

class _StatsGrid extends StatelessWidget {
  final List<_StatItem> stats;
  final Color surface;
  final Color textPrimary;
  final Color textSecondary;

  const _StatsGrid({
    required this.stats,
    required this.surface,
    required this.textPrimary,
    required this.textSecondary,
  });

  @override
  Widget build(BuildContext context) {
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 10,
      crossAxisSpacing: 10,
      childAspectRatio: 1.8,
      children: stats
          .map(
            (s) => Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: surface,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    s.value,
                    style: TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 22,
                      fontWeight: FontWeight.w700,
                      color: textPrimary,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    s.label,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: textSecondary,
                        ),
                  ),
                ],
              ),
            ),
          )
          .toList(),
    );
  }
}

// ---------------------------------------------------------------------------
// Settings tile
// ---------------------------------------------------------------------------

class _SettingsTile extends StatelessWidget {
  final IconData icon;
  final String label;
  final Widget? trailing;
  final Color surface;
  final Color textPrimary;
  final VoidCallback? onTap;

  const _SettingsTile({
    required this.icon,
    required this.label,
    required this.trailing,
    required this.surface,
    required this.textPrimary,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Icon(icon, color: textPrimary, size: 22),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                label,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      color: textPrimary,
                    ),
              ),
            ),
            ?trailing,
            if (trailing != null) const SizedBox(width: 4),
            Icon(Icons.chevron_right, color: textPrimary, size: 20),
          ],
        ),
      ),
    );
  }
}
