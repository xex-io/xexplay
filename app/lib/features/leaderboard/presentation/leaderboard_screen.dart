import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/error_widget.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../providers/leaderboard_provider.dart';
import 'leaderboard_row.dart';

class LeaderboardScreen extends ConsumerWidget {
  const LeaderboardScreen({super.key});

  static const _tabs = [
    ('daily', 'Daily'),
    ('weekly', 'Weekly'),
    ('tournament', 'Tournament'),
    ('all-time', 'All-Time'),
    ('friends', 'Friends'),
  ];

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selectedType = ref.watch(selectedLeaderboardTypeProvider);

    final params = LeaderboardParams(type: selectedType);
    final leaderboardAsync = ref.watch(leaderboardDataProvider(params));

    return Scaffold(
      appBar: AppBar(title: const Text('Leaderboard'), centerTitle: true),
      body: Column(
        children: [
          // Tab bar
          _TabBar(
            selectedType: selectedType,
            onSelected: (type) {
              ref.read(selectedLeaderboardTypeProvider.notifier).state = type;
            },
          ),

          // Content
          Expanded(
            child: leaderboardAsync.when(
              loading: () => const LoadingWidget(),
              error: (error, _) => AppErrorWidget(
                message: error.toString(),
                onRetry: () =>
                    ref.invalidate(leaderboardDataProvider(params)),
              ),
              data: (data) => _LeaderboardList(data: data),
            ),
          ),
        ],
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Tab bar with pill buttons
// ---------------------------------------------------------------------------

class _TabBar extends StatelessWidget {
  const _TabBar({required this.selectedType, required this.onSelected});

  final String selectedType;
  final void Function(String) onSelected;

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;
    final surfaceRaised =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: LeaderboardScreen._tabs.map((tab) {
            final (type, label) = tab;
            final isActive = type == selectedType;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: GestureDetector(
                onTap: () => onSelected(type),
                child: AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 8,
                  ),
                  decoration: BoxDecoration(
                    color: isActive ? primaryColor : surfaceRaised,
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    label,
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: FontWeight.w600,
                      color: isActive ? Colors.white : textSecondary,
                    ),
                  ),
                ),
              ),
            );
          }).toList(),
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Leaderboard list with pull-to-refresh and pinned user rank
// ---------------------------------------------------------------------------

class _LeaderboardList extends ConsumerWidget {
  const _LeaderboardList({required this.data});

  final dynamic data; // LeaderboardData

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final leaderboardData = data;
    final entries = leaderboardData.entries;
    final userRank = leaderboardData.userRank;

    // Determine if the user is already visible in the list.
    final userVisibleInList = userRank != null &&
        entries.any((e) => e.userId == userRank.userId);

    return Column(
      children: [
        Expanded(
          child: RefreshIndicator(
            onRefresh: () async {
              final type = ref.read(selectedLeaderboardTypeProvider);
              final params = LeaderboardParams(type: type);
              ref.invalidate(leaderboardDataProvider(params));
              // Wait for the new data to load.
              await ref.read(leaderboardDataProvider(params).future);
            },
            child: entries.isEmpty
                ? ListView(
                    children: const [
                      SizedBox(height: 120),
                      Center(
                        child: Text('No entries yet.'),
                      ),
                    ],
                  )
                : ListView.separated(
                    padding: const EdgeInsets.only(bottom: 8),
                    itemCount: entries.length,
                    separatorBuilder: (_, _) => const Divider(height: 1),
                    itemBuilder: (context, index) {
                      final entry = entries[index];
                      final isCurrent =
                          userRank != null && entry.userId == userRank.userId;
                      return LeaderboardRow(
                        entry: entry,
                        isCurrentUser: isCurrent,
                      );
                    },
                  ),
          ),
        ),

        // Pinned user rank at bottom
        if (userRank != null && !userVisibleInList) ...[
          const Divider(height: 1),
          LeaderboardRow(entry: userRank, isCurrentUser: true),
          SizedBox(height: MediaQuery.of(context).padding.bottom),
        ],
      ],
    );
  }
}
