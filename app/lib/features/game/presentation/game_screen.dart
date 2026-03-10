import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/loading_widget.dart';
import '../../../shared/widgets/error_widget.dart';
import '../data/card_models.dart' as models;
import '../domain/game_state.dart';
import '../providers/game_provider.dart';
import 'card_widget.dart';
import 'swipeable_card.dart';
import 'timer_widget.dart';

/// The active game session screen with card swiping.
class GameScreen extends ConsumerStatefulWidget {
  const GameScreen({super.key});

  @override
  ConsumerState<GameScreen> createState() => _GameScreenState();
}

class _GameScreenState extends ConsumerState<GameScreen>
    with SingleTickerProviderStateMixin {
  /// Key for the timer widget — reset when the card changes.
  Key _timerKey = UniqueKey();

  /// Key for the swipeable card — reset when the card changes.
  Key _cardKey = UniqueKey();

  /// Track current card ID to detect card transitions.
  String? _currentCardId;

  /// Animation controller for the next-card scale transition.
  late AnimationController _nextCardScaleController;
  @override
  void initState() {
    super.initState();
    _nextCardScaleController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 400),
    );
  }

  @override
  void dispose() {
    _nextCardScaleController.dispose();
    super.dispose();
  }

  void _onCardChanged(String newCardId) {
    if (_currentCardId != null && _currentCardId != newCardId) {
      // Card transitioned — animate next card scaling up
      _nextCardScaleController.forward(from: 0.0);
    }
    _currentCardId = newCardId;
    _timerKey = ValueKey(newCardId);
    _cardKey = ValueKey('card_$newCardId');
  }

  CardTier _mapTier(models.CardTier modelTier) {
    switch (modelTier) {
      case models.CardTier.gold:
        return CardTier.gold;
      case models.CardTier.silver:
        return CardTier.silver;
      case models.CardTier.white:
        return CardTier.white;
    }
  }

  void _onTimerExpired() {
    ref.read(gameProvider.notifier).skipCard();
  }

  void _onSwipeRight() {
    ref.read(gameProvider.notifier).submitAnswer(true);
  }

  void _onSwipeLeft() {
    ref.read(gameProvider.notifier).submitAnswer(false);
  }

  void _onSwipeUp() {
    ref.read(gameProvider.notifier).skipCard();
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

    final state = ref.watch(gameProvider);

    // Handle navigation for SessionComplete
    if (state is SessionComplete) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          context.go('/play/summary');
        }
      });
    }

    // If not ActiveSession and not Loading, redirect to home
    if (state is NoSession) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          context.go('/play');
        }
      });
    }

    return Scaffold(
      backgroundColor: bgColor,
      body: SafeArea(
        child: switch (state) {
          Loading() => const LoadingWidget(message: 'Loading cards...'),
          GameError(:final message) => AppErrorWidget(
              message: message,
              onRetry: () =>
                  ref.read(gameProvider.notifier).checkDailyStatus(),
            ),
          ActiveSession() => _buildActiveSession(
              context,
              state,
              textPrimary: textPrimary,
              textSecondary: textSecondary,
              surfaceRaised: surfaceRaised,
            ),
          SessionComplete() =>
            const LoadingWidget(message: 'Loading summary...'),
          NoSession() =>
            const LoadingWidget(message: 'Returning to home...'),
        },
      ),
    );
  }

  Widget _buildActiveSession(
    BuildContext context,
    ActiveSession state, {
    required Color textPrimary,
    required Color textSecondary,
    required Color surfaceRaised,
  }) {
    final session = state.session;
    final card = state.currentCard;
    final answersUsed = session.answersUsed;
    final skipsUsed = session.skipsUsed;
    final maxAnswers = session.maxAnswers;
    final maxSkips = session.maxSkips;
    final cardIndex = session.currentCardIndex;
    final totalCards = session.totalCards;
    final canSkip = skipsUsed < maxSkips;

    // Detect card change
    _onCardChanged(card.cardId);

    // Resolve question text (use 'en' or first available)
    final question =
        card.questionText['en'] ?? card.questionText.values.firstOrNull ?? '';

    // Map tier from data model to widget enum
    final widgetTier = _mapTier(card.tier);

    return Column(
      children: [
        const SizedBox(height: 12),

        // Card progress
        Text(
          'Card ${cardIndex + 1} of $totalCards',
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: textSecondary,
          ),
        ),
        const SizedBox(height: 16),

        // Timer
        TimerWidget(
          key: _timerKey,
          duration: Duration(seconds: state.timeRemaining),
          onExpired: _onTimerExpired,
        ),
        const SizedBox(height: 24),

        // Card stack
        Expanded(
          child: Center(
            child: SizedBox(
              width: 300,
              height: 420,
              child: Stack(
                alignment: Alignment.center,
                children: [
                  // Next card peek (behind current card)
                  if (cardIndex + 1 < totalCards)
                    Positioned(
                      top: 3,
                      child: Transform.scale(
                        scale: 0.95,
                        child: Opacity(
                          opacity: 0.6,
                          child: PredictionCardWidget(
                            tier: widgetTier, // placeholder
                            question: '?',
                            points: 0,
                          ),
                        ),
                      ),
                    ),

                  // Current card (swipeable)
                  SwipeableCard(
                    key: _cardKey,
                    tier: widgetTier,
                    question: question,
                    points: card.pointsForCorrect,
                    onSwipeRight: _onSwipeRight,
                    onSwipeLeft: _onSwipeLeft,
                    onSwipeUp: _onSwipeUp,
                    canSkip: canSkip,
                  ),
                ],
              ),
            ),
          ),
        ),

        const SizedBox(height: 16),

        // Resource counters
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              _ResourceCounter(
                label: 'Answers',
                current: answersUsed,
                max: maxAnswers,
                color: textPrimary,
                labelColor: textSecondary,
                bgColor: surfaceRaised,
              ),
              _ResourceCounter(
                label: 'Skips',
                current: skipsUsed,
                max: maxSkips,
                color: textPrimary,
                labelColor: textSecondary,
                bgColor: surfaceRaised,
              ),
            ],
          ),
        ),

        const SizedBox(height: 24),
      ],
    );
  }
}

class _ResourceCounter extends StatelessWidget {
  const _ResourceCounter({
    required this.label,
    required this.current,
    required this.max,
    required this.color,
    required this.labelColor,
    required this.bgColor,
  });

  final String label;
  final int current;
  final int max;
  final Color color;
  final Color labelColor;
  final Color bgColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(12),
      ),
      child: Column(
        children: [
          Text(
            '$current/$max',
            style: TextStyle(
              fontFamily: 'monospace',
              fontSize: 20,
              fontWeight: FontWeight.w700,
              color: color,
            ),
          ),
          const SizedBox(height: 2),
          Text(
            label,
            style: TextStyle(
              fontSize: 12,
              color: labelColor,
            ),
          ),
        ],
      ),
    );
  }
}
