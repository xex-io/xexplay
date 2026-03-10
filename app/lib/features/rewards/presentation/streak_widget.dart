import 'dart:math';
import 'package:flutter/material.dart';
import '../../../core/constants/app_colors.dart';

/// Milestones for the streak ring progress indicator.
const List<int> _streakMilestones = [3, 7, 10, 14, 21, 30];

/// Returns the next milestone above the current streak,
/// and the previous milestone (or 0) as the base.
({int next, int previous}) _nextMilestone(int currentStreak) {
  for (final m in _streakMilestones) {
    if (currentStreak < m) {
      final idx = _streakMilestones.indexOf(m);
      final prev = idx > 0 ? _streakMilestones[idx - 1] : 0;
      return (next: m, previous: prev);
    }
  }
  // Past all milestones — show full ring.
  return (next: currentStreak, previous: 0);
}

/// A reusable streak display widget with a circular progress ring,
/// milestone markers, and the current streak number.
class StreakWidget extends StatelessWidget {
  final int currentStreak;
  final double size;

  const StreakWidget({
    super.key,
    required this.currentStreak,
    this.size = 140,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor =
        isDark ? AppColors.darkTextPrimary : AppColors.lightTextPrimary;
    final secondaryColor =
        isDark ? AppColors.darkTextSecondary : AppColors.lightTextSecondary;
    final primaryColor = isDark ? AppColors.darkPrimary : AppColors.lightPrimary;
    final surfaceColor =
        isDark ? AppColors.darkSurfaceRaised : AppColors.lightSurfaceRaised;

    final milestone = _nextMilestone(currentStreak);
    final range = milestone.next - milestone.previous;
    final progress =
        range > 0 ? (currentStreak - milestone.previous) / range : 1.0;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          width: size,
          height: size,
          child: CustomPaint(
            painter: _StreakRingPainter(
              progress: progress.clamp(0.0, 1.0),
              milestones: _streakMilestones,
              currentStreak: currentStreak,
              activeColor: primaryColor,
              trackColor: surfaceColor,
              markerColor: secondaryColor,
            ),
            child: Center(
              child: Text(
                '$currentStreak',
                style: TextStyle(
                  fontFamily: 'monospace',
                  fontSize: 48,
                  fontWeight: FontWeight.w700,
                  color: textColor,
                  height: 1.0,
                ),
              ),
            ),
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'day streak',
          style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                color: secondaryColor,
              ),
        ),
        if (currentStreak < _streakMilestones.last) ...[
          const SizedBox(height: 4),
          Text(
            '${milestone.next - currentStreak} days to next milestone',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: secondaryColor,
                ),
          ),
        ],
      ],
    );
  }
}

class _StreakRingPainter extends CustomPainter {
  final double progress;
  final List<int> milestones;
  final int currentStreak;
  final Color activeColor;
  final Color trackColor;
  final Color markerColor;

  _StreakRingPainter({
    required this.progress,
    required this.milestones,
    required this.currentStreak,
    required this.activeColor,
    required this.trackColor,
    required this.markerColor,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = (min(size.width, size.height) / 2) - 8;
    const strokeWidth = 8.0;
    const startAngle = -pi / 2; // 12 o'clock

    // Track (background ring).
    final trackPaint = Paint()
      ..color = trackColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;

    canvas.drawCircle(center, radius, trackPaint);

    // Active arc.
    final activePaint = Paint()
      ..color = activeColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;

    final sweepAngle = 2 * pi * progress;
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      startAngle,
      sweepAngle,
      false,
      activePaint,
    );

    // Milestone markers on the ring.
    final maxMilestone = milestones.last;
    final markerPaint = Paint()
      ..color = markerColor
      ..style = PaintingStyle.fill;

    for (final m in milestones) {
      final angle = startAngle + (2 * pi * m / maxMilestone);
      final markerCenter = Offset(
        center.dx + radius * cos(angle),
        center.dy + radius * sin(angle),
      );

      final dotRadius = currentStreak >= m ? 5.0 : 3.0;
      final dotColor = currentStreak >= m ? activeColor : markerColor;

      canvas.drawCircle(
          markerCenter, dotRadius, markerPaint..color = dotColor);
    }
  }

  @override
  bool shouldRepaint(covariant _StreakRingPainter oldDelegate) {
    return oldDelegate.progress != progress ||
        oldDelegate.currentStreak != currentStreak;
  }
}
