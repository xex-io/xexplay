import 'package:flutter/material.dart';

import '../../../core/constants/app_colors.dart';
import '../data/leaderboard_models.dart';

class LeaderboardRow extends StatelessWidget {
  const LeaderboardRow({
    super.key,
    required this.entry,
    this.isCurrentUser = false,
  });

  final LeaderboardEntry entry;
  final bool isCurrentUser;

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final surfaceRaised =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;

    final rankColor = _rankColor(entry.rank);

    return Container(
      decoration: BoxDecoration(
        color: isCurrentUser ? surfaceRaised : null,
        border: isCurrentUser
            ? Border(left: BorderSide(color: primaryColor, width: 3))
            : null,
      ),
      padding: EdgeInsets.symmetric(
        horizontal: isCurrentUser ? 13 : 16,
        vertical: 12,
      ),
      child: Row(
        children: [
          // Rank number
          SizedBox(
            width: 32,
            child: Text(
              '${entry.rank}',
              style: TextStyle(
                fontFamily: 'monospace',
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: rankColor ?? Theme.of(context).textTheme.bodyLarge?.color,
              ),
            ),
          ),
          const SizedBox(width: 12),

          // Avatar
          _buildAvatar(context, primaryColor, textSecondary),
          const SizedBox(width: 12),

          // Username
          Expanded(
            child: Text(
              entry.displayName,
              style: Theme.of(context).textTheme.labelLarge,
              overflow: TextOverflow.ellipsis,
              maxLines: 1,
            ),
          ),
          const SizedBox(width: 8),

          // Points
          Text(
            '${entry.totalPoints}',
            style: TextStyle(
              fontFamily: 'monospace',
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: Theme.of(context).textTheme.bodyLarge?.color,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAvatar(
    BuildContext context,
    Color primaryColor,
    Color textSecondary,
  ) {
    if (entry.avatarUrl != null && entry.avatarUrl!.isNotEmpty) {
      return CircleAvatar(
        radius: 16,
        backgroundImage: NetworkImage(entry.avatarUrl!),
        backgroundColor: primaryColor.withValues(alpha: 0.15),
      );
    }

    return CircleAvatar(
      radius: 16,
      backgroundColor: primaryColor.withValues(alpha: 0.15),
      child: Icon(Icons.person, size: 18, color: textSecondary),
    );
  }

  /// Returns a highlight color for top-3 ranks, or null for standard style.
  Color? _rankColor(int rank) {
    return switch (rank) {
      1 => AppColors.goldStart,
      2 => AppColors.silverStart,
      3 => const Color(0xFFCD7F32), // bronze
      _ => null,
    };
  }
}
