import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/social_models.dart';
import '../providers/social_provider.dart';

class AchievementsScreen extends ConsumerWidget {
  const AchievementsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final achievementsAsync = ref.watch(achievementsProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Achievements'), centerTitle: true),
      body: achievementsAsync.when(
        loading: () => const LoadingWidget(),
        error: (error, _) => AppErrorWidget(
          message: error.toString(),
          onRetry: () => ref.invalidate(achievementsProvider),
        ),
        data: (achievements) => achievements.isEmpty
            ? const Center(child: Text('No achievements yet.'))
            : GridView.builder(
                padding: const EdgeInsets.all(16),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 3,
                  mainAxisSpacing: 12,
                  crossAxisSpacing: 12,
                  childAspectRatio: 0.8,
                ),
                itemCount: achievements.length,
                itemBuilder: (context, index) {
                  final ua = achievements[index];
                  return _AchievementBadge(
                    userAchievement: ua,
                    onTap: () => _showDetail(context, ua),
                  );
                },
              ),
      ),
    );
  }

  void _showDetail(BuildContext context, UserAchievement ua) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;

    showModalBottomSheet(
      context: context,
      backgroundColor: surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              width: 40,
              height: 4,
              decoration: BoxDecoration(
                color: textSecondary.withValues(alpha: 0.3),
                borderRadius: BorderRadius.circular(2),
              ),
            ),
            const SizedBox(height: 20),
            Icon(
              ua.earned ? Icons.emoji_events : Icons.lock_outline,
              size: 48,
              color: ua.earned ? primaryColor : textSecondary,
            ),
            const SizedBox(height: 16),
            Text(
              ua.achievement.name,
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                    color: textPrimary,
                    fontWeight: FontWeight.w700,
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              ua.achievement.description,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: textSecondary,
                  ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            if (ua.earned && ua.earnedAt != null)
              Text(
                'Earned on ${DateFormat.yMMMd().format(ua.earnedAt!)}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: primaryColor,
                    ),
              )
            else
              Text(
                'Progress: ${ua.progress} / ${ua.achievement.requiredValue}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: textSecondary,
                    ),
              ),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Achievement badge tile
// ---------------------------------------------------------------------------

class _AchievementBadge extends StatelessWidget {
  final UserAchievement userAchievement;
  final VoidCallback onTap;

  const _AchievementBadge({
    required this.userAchievement,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;
    final earned = userAchievement.earned;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(14),
          boxShadow: earned
              ? [
                  BoxShadow(
                    color: primaryColor.withValues(alpha: 0.25),
                    blurRadius: 12,
                    spreadRadius: 0,
                  ),
                ]
              : null,
        ),
        padding: const EdgeInsets.all(10),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Stack(
              alignment: Alignment.center,
              children: [
                Icon(
                  Icons.emoji_events,
                  size: 36,
                  color: earned ? primaryColor : textSecondary.withValues(alpha: 0.4),
                ),
                if (!earned)
                  Icon(
                    Icons.lock,
                    size: 18,
                    color: textSecondary,
                  ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              userAchievement.achievement.name,
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: earned ? textPrimary : textSecondary,
              ),
              textAlign: TextAlign.center,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ],
        ),
      ),
    );
  }
}
