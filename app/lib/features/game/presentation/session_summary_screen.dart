import 'package:flutter/material.dart';
import '../../../core/constants/app_colors.dart';
import '../../../shared/widgets/app_button.dart';
import '../../social/presentation/share_card_widget.dart';
import '../../social/presentation/share_screen.dart';

class SessionSummaryScreen extends StatefulWidget {
  const SessionSummaryScreen({
    super.key,
    this.totalScore = 0,
    this.cardsAnswered = 0,
    this.correctPredictions = 0,
    this.pointsEarned = 0,
    this.skipsUsed = 0,
    this.currentStreak = 0,
  });

  final int totalScore;
  final int cardsAnswered;
  final int correctPredictions;
  final int pointsEarned;
  final int skipsUsed;
  final int currentStreak;

  @override
  State<SessionSummaryScreen> createState() => _SessionSummaryScreenState();
}

class _SessionSummaryScreenState extends State<SessionSummaryScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<int> _scoreAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    );
    _scoreAnimation = IntTween(
      begin: 0,
      end: widget.totalScore,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOut));

    _controller.forward();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
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

    return Scaffold(
      backgroundColor: bgColor,
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: Column(
            children: [
              const Spacer(),

              // Title
              Text(
                'Session Complete',
                style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      color: textSecondary,
                    ),
              ),
              const SizedBox(height: 24),

              // Animated score
              AnimatedBuilder(
                animation: _scoreAnimation,
                builder: (context, child) {
                  return Text(
                    '${_scoreAnimation.value}',
                    style: TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 64,
                      fontWeight: FontWeight.w700,
                      color: textPrimary,
                    ),
                  );
                },
              ),
              Text(
                'Total Points',
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: textSecondary,
                    ),
              ),

              const SizedBox(height: 48),

              // Stats grid
              GridView.count(
                crossAxisCount: 2,
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                mainAxisSpacing: 12,
                crossAxisSpacing: 12,
                childAspectRatio: 1.6,
                children: [
                  _StatTile(
                    label: 'Cards Answered',
                    value: '${widget.cardsAnswered}',
                    bgColor: surfaceRaised,
                    valueColor: textPrimary,
                    labelColor: textSecondary,
                  ),
                  _StatTile(
                    label: 'Correct Predictions',
                    value: '${widget.correctPredictions}',
                    bgColor: surfaceRaised,
                    valueColor: textPrimary,
                    labelColor: textSecondary,
                  ),
                  _StatTile(
                    label: 'Points Earned',
                    value: '${widget.pointsEarned}',
                    bgColor: surfaceRaised,
                    valueColor: textPrimary,
                    labelColor: textSecondary,
                  ),
                  _StatTile(
                    label: 'Skips Used',
                    value: '${widget.skipsUsed}',
                    bgColor: surfaceRaised,
                    valueColor: textPrimary,
                    labelColor: textSecondary,
                  ),
                ],
              ),

              // Streak display
              if (widget.currentStreak > 0) ...[
                const SizedBox(height: 24),
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 14,
                  ),
                  decoration: BoxDecoration(
                    color: surfaceRaised,
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(
                      color: AppColors.goldStart.withAlpha(100),
                    ),
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(
                        Icons.local_fire_department,
                        color: AppColors.goldStart,
                        size: 28,
                      ),
                      const SizedBox(width: 10),
                      Text(
                        '${widget.currentStreak}-day streak!',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w700,
                          color: textPrimary,
                        ),
                      ),
                    ],
                  ),
                ),
              ],

              const Spacer(),

              // Share Results button
              SizedBox(
                width: double.infinity,
                child: SecondaryButton(
                  label: 'Share Results',
                  onPressed: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (_) => ShareScreen(
                          data: ShareData(
                            type: ShareType.score,
                            headline: 'Session Results',
                            value: '${widget.totalScore}',
                            subtitle:
                                '${widget.correctPredictions} correct predictions'
                                '${widget.currentStreak > 0 ? ' | ${widget.currentStreak}-day streak' : ''}',
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
              const SizedBox(height: 12),

              // Done button
              SizedBox(
                width: double.infinity,
                child: PrimaryButton(
                  label: 'Done',
                  onPressed: () {
                    Navigator.of(context).popUntil((route) => route.isFirst);
                  },
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _StatTile extends StatelessWidget {
  const _StatTile({
    required this.label,
    required this.value,
    required this.bgColor,
    required this.valueColor,
    required this.labelColor,
  });

  final String label;
  final String value;
  final Color bgColor;
  final Color valueColor;
  final Color labelColor;

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: bgColor,
        borderRadius: BorderRadius.circular(12),
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(
            value,
            style: TextStyle(
              fontFamily: 'monospace',
              fontSize: 24,
              fontWeight: FontWeight.w700,
              color: valueColor,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 12,
              color: labelColor,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
