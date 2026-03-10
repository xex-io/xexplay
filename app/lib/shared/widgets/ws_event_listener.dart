import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../core/constants/app_colors.dart';
import '../../core/network/websocket_client.dart';
import '../../core/providers/ws_provider.dart';
import '../../features/leaderboard/providers/leaderboard_provider.dart';
import 'achievement_celebration.dart';

/// Wraps the app shell and listens to [wsEventsProvider] to trigger
/// global UI reactions for real-time WebSocket events.
class WsEventListener extends ConsumerStatefulWidget {
  const WsEventListener({super.key, required this.child});

  final Widget child;

  @override
  ConsumerState<WsEventListener> createState() => _WsEventListenerState();
}

class _WsEventListenerState extends ConsumerState<WsEventListener> {
  OverlayEntry? _achievementOverlay;

  @override
  void dispose() {
    _dismissAchievement();
    super.dispose();
  }

  void _dismissAchievement() {
    _achievementOverlay?.remove();
    _achievementOverlay = null;
  }

  void _handleEvent(WsEvent event) {
    switch (event.type) {
      case 'card_resolved':
        _showCardResolved(event.data);
      case 'leaderboard_update':
        _handleLeaderboardUpdate();
      case 'achievement_unlocked':
        _showAchievementUnlocked(event.data);
      case 'reward_earned':
        _showRewardEarned(event.data);
    }
  }

  void _showCardResolved(Map<String, dynamic> data) {
    final correct = data['correct'] as bool? ?? false;
    final points = data['points'] as int? ?? 0;
    final message = correct
        ? 'Your prediction was correct! +$points points'
        : 'Your prediction was incorrect!';
    final color = correct ? AppColors.darkPositive : AppColors.darkNegative;

    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message, style: const TextStyle(color: Colors.white)),
        backgroundColor: color,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 3),
      ),
    );
  }

  void _handleLeaderboardUpdate() {
    // Invalidate all leaderboard data providers so the screen refreshes
    // the next time it is visible.
    final type = ref.read(selectedLeaderboardTypeProvider);
    ref.invalidate(leaderboardDataProvider(LeaderboardParams(type: type)));
  }

  void _showAchievementUnlocked(Map<String, dynamic> data) {
    final name = data['name'] as String? ?? 'Achievement';
    final description = data['description'] as String? ?? '';

    // Dismiss any existing overlay first.
    _dismissAchievement();

    final overlay = Overlay.of(context, rootOverlay: true);
    _achievementOverlay = OverlayEntry(
      builder: (_) => AchievementCelebration(
        name: name,
        description: description,
        onDismiss: _dismissAchievement,
      ),
    );
    overlay.insert(_achievementOverlay!);
  }

  void _showRewardEarned(Map<String, dynamic> data) {
    final amount = data['amount'] ?? data['tokens'] ?? 0;

    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          'You earned $amount tokens!',
          style: const TextStyle(color: Colors.black87),
        ),
        backgroundColor: AppColors.goldStart,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 3),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    ref.listen<AsyncValue<WsEvent>>(wsEventsProvider, (_, next) {
      next.whenData(_handleEvent);
    });

    return widget.child;
  }
}
