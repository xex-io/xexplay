import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../../leaderboard/data/leaderboard_models.dart';
import '../../leaderboard/presentation/leaderboard_row.dart';
import '../providers/social_provider.dart';

class LeagueDetailScreen extends ConsumerWidget {
  const LeagueDetailScreen({super.key, required this.leagueId});

  final String leagueId;

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

    final leagueAsync = ref.watch(leagueDetailProvider(leagueId));

    return Scaffold(
      appBar: AppBar(title: const Text('League'), centerTitle: true),
      body: leagueAsync.when(
        loading: () => const LoadingWidget(),
        error: (error, _) => AppErrorWidget(
          message: error.toString(),
          onRetry: () => ref.invalidate(leagueDetailProvider(leagueId)),
        ),
        data: (league) => RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(leagueDetailProvider(leagueId));
            await ref.read(leagueDetailProvider(leagueId).future);
          },
          child: ListView(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            children: [
              // --- League header ---
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: surface,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Column(
                  children: [
                    CircleAvatar(
                      radius: 28,
                      backgroundColor: primaryColor.withValues(alpha: 0.15),
                      child:
                          Icon(Icons.groups, color: primaryColor, size: 28),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      league.name,
                      style:
                          Theme.of(context).textTheme.titleLarge?.copyWith(
                                color: textPrimary,
                                fontWeight: FontWeight.w700,
                              ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${league.memberCount} members',
                      style:
                          Theme.of(context).textTheme.bodyMedium?.copyWith(
                                color: textSecondary,
                              ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 16),

              // --- Invite code ---
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
                decoration: BoxDecoration(
                  color: surface,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    Icon(Icons.link, color: textSecondary, size: 20),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Invite Code',
                            style: Theme.of(context)
                                .textTheme
                                .bodySmall
                                ?.copyWith(color: textSecondary),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            league.inviteCode,
                            style: TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 16,
                              fontWeight: FontWeight.w600,
                              color: primaryColor,
                              letterSpacing: 2,
                            ),
                          ),
                        ],
                      ),
                    ),
                    GestureDetector(
                      onTap: () {
                        Clipboard.setData(
                            ClipboardData(text: league.inviteCode));
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(
                            content: Text('Invite code copied!'),
                            duration: Duration(seconds: 2),
                          ),
                        );
                      },
                      child: Container(
                        padding: const EdgeInsets.all(8),
                        decoration: BoxDecoration(
                          color: primaryColor.withValues(alpha: 0.15),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child:
                            Icon(Icons.copy, size: 18, color: primaryColor),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // --- Members leaderboard ---
              Text(
                'Leaderboard',
                style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      color: textPrimary,
                    ),
              ),
              const SizedBox(height: 12),
              if (league.members.isEmpty)
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    color: surface,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Center(
                    child: Text(
                      'No members yet.',
                      style: TextStyle(color: textSecondary),
                    ),
                  ),
                )
              else
                ...league.members.map(
                  (member) => LeaderboardRow(
                    entry: LeaderboardEntry(
                      rank: member.rank,
                      userId: member.userId,
                      displayName: member.displayName,
                      avatarUrl: member.avatarUrl,
                      totalPoints: member.totalPoints,
                      correctAnswers: 0,
                    ),
                  ),
                ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }
}
