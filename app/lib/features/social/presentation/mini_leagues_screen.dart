import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/app_button.dart';
import '../../../shared/widgets/error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../data/social_models.dart';
import '../providers/social_provider.dart';

class MiniLeaguesScreen extends ConsumerWidget {
  const MiniLeaguesScreen({super.key});

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

    final leaguesAsync = ref.watch(leaguesProvider);

    return Scaffold(
      appBar: AppBar(title: const Text('Mini-Leagues'), centerTitle: true),
      body: Column(
        children: [
          // --- Action buttons ---
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 8),
            child: Row(
              children: [
                Expanded(
                  child: PrimaryButton(
                    label: 'Create League',
                    onPressed: () => _showCreateDialog(context, ref),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: SecondaryButton(
                    label: 'Join League',
                    onPressed: () => _showJoinDialog(context, ref),
                  ),
                ),
              ],
            ),
          ),

          // --- Leagues list ---
          Expanded(
            child: leaguesAsync.when(
              loading: () => const LoadingWidget(),
              error: (error, _) => AppErrorWidget(
                message: error.toString(),
                onRetry: () => ref.invalidate(leaguesProvider),
              ),
              data: (leagues) => leagues.isEmpty
                  ? Center(
                      child: Text(
                        'No leagues yet.\nCreate or join one!',
                        textAlign: TextAlign.center,
                        style: TextStyle(color: textSecondary),
                      ),
                    )
                  : RefreshIndicator(
                      onRefresh: () async {
                        ref.invalidate(leaguesProvider);
                        await ref.read(leaguesProvider.future);
                      },
                      child: ListView.separated(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 20,
                          vertical: 12,
                        ),
                        itemCount: leagues.length,
                        separatorBuilder: (_, _) => const SizedBox(height: 10),
                        itemBuilder: (context, index) => _LeagueCard(
                          league: leagues[index],
                          surface: surface,
                          textPrimary: textPrimary,
                          textSecondary: textSecondary,
                          primaryColor: primaryColor,
                          onTap: () =>
                              context.push('/leagues/${leagues[index].id}'),
                        ),
                      ),
                    ),
            ),
          ),
        ],
      ),
    );
  }

  void _showCreateDialog(BuildContext context, WidgetRef ref) {
    final nameController = TextEditingController();
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurface : AppColors.lightSurface;

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: surface,
        title: const Text('Create League'),
        content: TextField(
          controller: nameController,
          decoration: const InputDecoration(
            hintText: 'League name',
            border: OutlineInputBorder(),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              final name = nameController.text.trim();
              if (name.isEmpty) return;
              Navigator.pop(ctx);
              try {
                await ref
                    .read(socialRemoteSourceProvider)
                    .createLeague(name: name);
                ref.invalidate(leaguesProvider);
              } catch (e) {
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Failed to create league: $e')),
                  );
                }
              }
            },
            child: const Text('Create'),
          ),
        ],
      ),
    );
  }

  void _showJoinDialog(BuildContext context, WidgetRef ref) {
    final codeController = TextEditingController();
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surface =
        isDark ? AppColors.darkSurface : AppColors.lightSurface;

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: surface,
        title: const Text('Join League'),
        content: TextField(
          controller: codeController,
          decoration: const InputDecoration(
            hintText: 'Invite code',
            border: OutlineInputBorder(),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () async {
              final code = codeController.text.trim();
              if (code.isEmpty) return;
              Navigator.pop(ctx);
              try {
                await ref
                    .read(socialRemoteSourceProvider)
                    .joinLeague(code);
                ref.invalidate(leaguesProvider);
              } catch (e) {
                if (context.mounted) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    SnackBar(content: Text('Failed to join league: $e')),
                  );
                }
              }
            },
            child: const Text('Join'),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// League card
// ---------------------------------------------------------------------------

class _LeagueCard extends StatelessWidget {
  final League league;
  final Color surface;
  final Color textPrimary;
  final Color textSecondary;
  final Color primaryColor;
  final VoidCallback onTap;

  const _LeagueCard({
    required this.league,
    required this.surface,
    required this.textPrimary,
    required this.textSecondary,
    required this.primaryColor,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: surface,
          borderRadius: BorderRadius.circular(14),
        ),
        child: Row(
          children: [
            CircleAvatar(
              radius: 22,
              backgroundColor: primaryColor.withValues(alpha: 0.15),
              child: Icon(Icons.groups, color: primaryColor, size: 22),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    league.name,
                    style: Theme.of(context).textTheme.labelLarge?.copyWith(
                          color: textPrimary,
                          fontWeight: FontWeight.w600,
                        ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${league.memberCount} members  •  by ${league.creatorName}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: textSecondary,
                        ),
                  ),
                ],
              ),
            ),
            Icon(Icons.chevron_right, color: textSecondary, size: 20),
          ],
        ),
      ),
    );
  }
}
