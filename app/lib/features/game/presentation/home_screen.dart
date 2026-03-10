import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:intl/intl.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/app_button.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../../../shared/widgets/error_widget.dart';
import '../domain/game_state.dart';
import '../providers/game_provider.dart';

/// Home / Play tab content.
///
/// Shows daily status and lets the user start or resume a game session.
class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    // Check daily status on first load.
    Future.microtask(() {
      ref.read(gameProvider.notifier).checkDailyStatus();
    });
  }

  Future<void> _startAndNavigate(BuildContext ctx) async {
    final router = GoRouter.of(ctx);
    await ref.read(gameProvider.notifier).startSession();
    if (!mounted) return;
    final newState = ref.read(gameProvider);
    if (newState is ActiveSession) {
      router.go('/play/session');
    }
  }

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final bgColor =
        isDark ? AppColors.darkBackground : AppColors.lightBackground;
    final textPrimary =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final textSecondary =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final surfaceRaised =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;
    final surface = isDark ? AppColors.darkSurface : AppColors.lightSurface;

    final state = ref.watch(gameProvider);

    return Scaffold(
      backgroundColor: bgColor,
      body: SafeArea(
        child: switch (state) {
          Loading() => const LoadingWidget(message: 'Loading...'),
          GameError(:final message) => AppErrorWidget(
              message: message,
              onRetry: () =>
                  ref.read(gameProvider.notifier).checkDailyStatus(),
            ),
          ActiveSession() => _buildActiveSessionContent(
              context,
              state,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              surfaceRaised: surfaceRaised,
              surface: surface,
            ),
          SessionComplete(:final session) => _buildCompletedContent(
              context,
              score: session.score,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              surfaceRaised: surfaceRaised,
              surface: surface,
            ),
          NoSession(:final dailyStatus) => _buildHomeContent(
              context,
              dailyStatus: dailyStatus,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              surfaceRaised: surfaceRaised,
              surface: surface,
            ),
        },
      ),
    );
  }

  Widget _buildHomeContent(
    BuildContext context, {
    required dynamic dailyStatus,
    required Color textPrimary,
    required Color textSecondary,
    required Color surfaceRaised,
    required Color surface,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final primaryColor =
        isDark ? AppColors.darkPrimary : AppColors.lightPrimary;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SizedBox(height: 16),

          // Title
          Text(
            'XEX Play',
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  color: textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 24),

          // Event info card
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: surface,
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: isDark
                    ? AppColors.darkBorder
                    : AppColors.lightBorder,
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(Icons.sports_soccer, color: primaryColor, size: 20),
                    const SizedBox(width: 8),
                    Text(
                      'Daily Prediction Challenge',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: textPrimary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Text(
                  DateFormat('EEEE, MMM d, yyyy').format(DateTime.now()),
                  style: TextStyle(
                    fontSize: 13,
                    color: textSecondary,
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Daily status card
          _buildDailyStatusCard(
            context,
            dailyStatus: dailyStatus,
            textPrimary: textPrimary,
            textSecondary: textSecondary,
            surfaceRaised: surfaceRaised,
            surface: surface,
          ),

          const Spacer(),

          // Swipe instructions
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: surfaceRaised,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Column(
              children: [
                Text(
                  'How to Play',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: textPrimary,
                  ),
                ),
                const SizedBox(height: 12),
                _buildInstruction(
                    Icons.arrow_forward, 'Swipe right for YES', textSecondary),
                const SizedBox(height: 6),
                _buildInstruction(
                    Icons.arrow_back, 'Swipe left for NO', textSecondary),
                const SizedBox(height: 6),
                _buildInstruction(
                    Icons.arrow_upward, 'Swipe up to SKIP', textSecondary),
              ],
            ),
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildInstruction(IconData icon, String text, Color color) {
    return Row(
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 8),
        Text(text, style: TextStyle(fontSize: 13, color: color)),
      ],
    );
  }

  Widget _buildDailyStatusCard(
    BuildContext context, {
    required dynamic dailyStatus,
    required Color textPrimary,
    required Color textSecondary,
    required Color surfaceRaised,
    required Color surface,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final positiveColor =
        isDark ? AppColors.darkPositive : AppColors.lightPositive;

    if (dailyStatus == null) {
      // No status yet — show ready state
      return _statusCard(
        context,
        title: "Today's Basket",
        subtitle: '15 Cards Ready',
        icon: Icons.style,
        iconColor: isDark ? AppColors.darkPrimary : AppColors.lightPrimary,
        textPrimary: textPrimary,
        textSecondary: textSecondary,
        surface: surface,
        action: SizedBox(
          width: double.infinity,
          child: PrimaryButton(
            label: 'Start Playing',
            onPressed: () => _startAndNavigate(context),
          ),
        ),
      );
    }

    if (dailyStatus.hasPlayedToday && dailyStatus.score != null) {
      // Completed
      return _statusCard(
        context,
        title: "Today's Basket",
        subtitle: 'Completed! Score: ${dailyStatus.score}',
        icon: Icons.check_circle,
        iconColor: positiveColor,
        textPrimary: textPrimary,
        textSecondary: textSecondary,
        surface: surface,
      );
    }

    if (dailyStatus.activeSessionId != null) {
      // In progress
      return _statusCard(
        context,
        title: "Today's Basket",
        subtitle: 'Session in progress',
        icon: Icons.play_circle,
        iconColor: isDark ? AppColors.darkWarning : AppColors.lightWarning,
        textPrimary: textPrimary,
        textSecondary: textSecondary,
        surface: surface,
        action: SizedBox(
          width: double.infinity,
          child: PrimaryButton(
            label: 'Resume Session',
            onPressed: () {
              context.go('/play/session');
            },
          ),
        ),
      );
    }

    if (dailyStatus.sessionAvailable) {
      // Ready
      return _statusCard(
        context,
        title: "Today's Basket",
        subtitle: '15 Cards Ready',
        icon: Icons.style,
        iconColor: isDark ? AppColors.darkPrimary : AppColors.lightPrimary,
        textPrimary: textPrimary,
        textSecondary: textSecondary,
        surface: surface,
        action: SizedBox(
          width: double.infinity,
          child: PrimaryButton(
            label: 'Start Playing',
            onPressed: () => _startAndNavigate(context),
          ),
        ),
      );
    }

    // Not available
    return _statusCard(
      context,
      title: "Today's Basket",
      subtitle: 'No basket available today',
      icon: Icons.hourglass_empty,
      iconColor: textSecondary,
      textPrimary: textPrimary,
      textSecondary: textSecondary,
      surface: surface,
    );
  }

  Widget _statusCard(
    BuildContext context, {
    required String title,
    required String subtitle,
    required IconData icon,
    required Color iconColor,
    required Color textPrimary,
    required Color textSecondary,
    required Color surface,
    Widget? action,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: surface,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: isDark ? AppColors.darkBorder : AppColors.lightBorder,
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: iconColor, size: 24),
              const SizedBox(width: 10),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                      color: textPrimary,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 13,
                      color: textSecondary,
                    ),
                  ),
                ],
              ),
            ],
          ),
          if (action != null) ...[
            const SizedBox(height: 16),
            action,
          ],
        ],
      ),
    );
  }

  Widget _buildActiveSessionContent(
    BuildContext context,
    ActiveSession state, {
    required Color textPrimary,
    required Color textSecondary,
    required Color surfaceRaised,
    required Color surface,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
      child: Column(
        children: [
          const SizedBox(height: 16),
          Text(
            'Session in Progress',
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  color: textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 16),
          Text(
            'Card ${state.session.currentCardIndex + 1} of ${state.session.totalCards}',
            style: TextStyle(fontSize: 14, color: textSecondary),
          ),
          const Spacer(),
          SizedBox(
            width: double.infinity,
            child: PrimaryButton(
              label: 'Resume Playing',
              onPressed: () => context.go('/play/session'),
            ),
          ),
          const SizedBox(height: 16),
        ],
      ),
    );
  }

  Widget _buildCompletedContent(
    BuildContext context, {
    required int score,
    required Color textPrimary,
    required Color textSecondary,
    required Color surfaceRaised,
    required Color surface,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final positiveColor =
        isDark ? AppColors.darkPositive : AppColors.lightPositive;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.check_circle, size: 64, color: positiveColor),
          const SizedBox(height: 16),
          Text(
            'Session Complete!',
            style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                  color: textPrimary,
                  fontWeight: FontWeight.w700,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            'Score: $score',
            style: TextStyle(
              fontSize: 24,
              fontWeight: FontWeight.w600,
              fontFamily: 'monospace',
              color: textSecondary,
            ),
          ),
          const SizedBox(height: 32),
          SizedBox(
            width: double.infinity,
            child: PrimaryButton(
              label: 'View Summary',
              onPressed: () => context.go('/play/summary'),
            ),
          ),
        ],
      ),
    );
  }
}
