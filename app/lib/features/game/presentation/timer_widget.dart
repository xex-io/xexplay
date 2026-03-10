import 'dart:math';
import 'package:flutter/material.dart';
import '../../../core/constants/app_colors.dart';

class TimerWidget extends StatefulWidget {
  const TimerWidget({
    super.key,
    this.duration = const Duration(seconds: 40),
    this.onExpired,
  });

  final Duration duration;
  final VoidCallback? onExpired;

  @override
  State<TimerWidget> createState() => _TimerWidgetState();
}

class _TimerWidgetState extends State<TimerWidget>
    with TickerProviderStateMixin {
  late AnimationController _timerController;
  late AnimationController _pulseController;

  @override
  void initState() {
    super.initState();
    _timerController = AnimationController(
      vsync: this,
      duration: widget.duration,
    );
    _pulseController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 600),
    );

    _timerController.addListener(_onTimerTick);
    _timerController.addStatusListener((status) {
      if (status == AnimationStatus.completed) {
        _pulseController.stop();
        widget.onExpired?.call();
      }
    });

    _timerController.forward();
  }

  void _onTimerTick() {
    final remaining = _remainingSeconds;
    if (remaining <= 10 && !_pulseController.isAnimating) {
      _pulseController.repeat(reverse: true);
    }
    // Trigger rebuild for color / text changes.
    setState(() {});
  }

  int get _remainingSeconds {
    final fraction = 1.0 - _timerController.value;
    return (fraction * widget.duration.inSeconds).ceil();
  }

  Color get _ringColor {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final remaining = _remainingSeconds;
    if (remaining <= 5) {
      return isDark ? AppColors.darkNegative : AppColors.lightNegative;
    }
    if (remaining <= 10) {
      return isDark ? AppColors.darkWarning : AppColors.lightWarning;
    }
    return isDark ? AppColors.darkPositive : AppColors.lightPositive;
  }

  @override
  void dispose() {
    _timerController.removeListener(_onTimerTick);
    _timerController.dispose();
    _pulseController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final remaining = _remainingSeconds;
    final color = _ringColor;
    final progress = 1.0 - _timerController.value;

    Widget ring = SizedBox(
      width: 80,
      height: 80,
      child: CustomPaint(
        painter: _CircularTimerPainter(
          progress: progress,
          color: color,
          trackColor: color.withAlpha(40),
        ),
        child: Center(
          child: Text(
            '$remaining',
            style: TextStyle(
              fontFamily: 'monospace',
              fontSize: 28,
              fontWeight: FontWeight.w700,
              color: color,
            ),
          ),
        ),
      ),
    );

    if (remaining <= 10) {
      return AnimatedBuilder(
        animation: _pulseController,
        builder: (context, child) {
          final scale = 1.0 + _pulseController.value * 0.06;
          return Transform.scale(scale: scale, child: child);
        },
        child: ring,
      );
    }

    return ring;
  }
}

class _CircularTimerPainter extends CustomPainter {
  _CircularTimerPainter({
    required this.progress,
    required this.color,
    required this.trackColor,
  });

  final double progress;
  final Color color;
  final Color trackColor;

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2 - 4;
    const strokeWidth = 4.0;

    // Track
    final trackPaint = Paint()
      ..color = trackColor
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;

    canvas.drawCircle(center, radius, trackPaint);

    // Progress arc
    final progressPaint = Paint()
      ..color = color
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;

    final sweepAngle = 2 * pi * progress;
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -pi / 2,
      sweepAngle,
      false,
      progressPaint,
    );
  }

  @override
  bool shouldRepaint(covariant _CircularTimerPainter oldDelegate) =>
      oldDelegate.progress != progress || oldDelegate.color != color;
}
