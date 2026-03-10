import 'dart:math';

import 'package:flutter/material.dart';

import '../../core/constants/app_colors.dart';

/// Full-screen overlay that celebrates an unlocked achievement.
///
/// Displays a scale-in animated card with the achievement details,
/// shimmer particles, and auto-dismisses after 3 seconds.
class AchievementCelebration extends StatefulWidget {
  const AchievementCelebration({
    super.key,
    required this.name,
    required this.description,
    this.iconData = Icons.emoji_events,
    this.onDismiss,
  });

  final String name;
  final String description;
  final IconData iconData;
  final VoidCallback? onDismiss;

  @override
  State<AchievementCelebration> createState() =>
      _AchievementCelebrationState();
}

class _AchievementCelebrationState extends State<AchievementCelebration>
    with TickerProviderStateMixin {
  late AnimationController _scaleController;
  late Animation<double> _scaleAnimation;
  late AnimationController _particleController;

  // Random positions for shimmer dots.
  late List<_Particle> _particles;

  @override
  void initState() {
    super.initState();

    _scaleController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 500),
    );
    _scaleAnimation = CurvedAnimation(
      parent: _scaleController,
      curve: Curves.elasticOut,
    );
    _scaleController.forward();

    _particleController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 2500),
    )..repeat();

    final rng = Random();
    _particles = List.generate(
      14,
      (_) => _Particle(
        x: rng.nextDouble(),
        y: rng.nextDouble(),
        size: 3.0 + rng.nextDouble() * 5,
        speed: 0.3 + rng.nextDouble() * 0.7,
      ),
    );

    // Auto-dismiss after 3 seconds.
    Future.delayed(const Duration(seconds: 3), () {
      if (mounted) widget.onDismiss?.call();
    });
  }

  @override
  void dispose() {
    _scaleController.dispose();
    _particleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: widget.onDismiss,
      child: Material(
        color: Colors.black.withAlpha(180),
        child: Stack(
          children: [
            // Shimmer particles
            AnimatedBuilder(
              animation: _particleController,
              builder: (context, _) {
                return CustomPaint(
                  size: MediaQuery.of(context).size,
                  painter: _ParticlePainter(
                    particles: _particles,
                    progress: _particleController.value,
                  ),
                );
              },
            ),

            // Centered achievement card
            Center(
              child: ScaleTransition(
                scale: _scaleAnimation,
                child: Container(
                  width: 300,
                  padding: const EdgeInsets.symmetric(
                    horizontal: 28,
                    vertical: 36,
                  ),
                  decoration: BoxDecoration(
                    color: AppColors.darkSurfaceRaised,
                    borderRadius: BorderRadius.circular(20),
                    border: Border.all(
                      color: AppColors.goldStart.withAlpha(180),
                      width: 2,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: AppColors.goldStart.withAlpha(60),
                        blurRadius: 32,
                        spreadRadius: 4,
                      ),
                    ],
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // Achievement icon
                      Container(
                        width: 72,
                        height: 72,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          gradient: AppColors.goldGradient,
                        ),
                        child: Icon(
                          widget.iconData,
                          size: 40,
                          color: AppColors.darkBackground,
                        ),
                      ),
                      const SizedBox(height: 20),

                      // Title
                      Text(
                        'Achievement Unlocked!',
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: AppColors.goldStart,
                          letterSpacing: 1.2,
                        ),
                      ),
                      const SizedBox(height: 12),

                      // Name
                      Text(
                        widget.name,
                        style: const TextStyle(
                          fontSize: 22,
                          fontWeight: FontWeight.w700,
                          color: AppColors.darkTextPrimary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                      const SizedBox(height: 8),

                      // Description
                      Text(
                        widget.description,
                        style: const TextStyle(
                          fontSize: 14,
                          color: AppColors.darkTextSecondary,
                        ),
                        textAlign: TextAlign.center,
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ---------------------------------------------------------------------------
// Particle data & painter for shimmer effect
// ---------------------------------------------------------------------------

class _Particle {
  final double x;
  final double y;
  final double size;
  final double speed;

  const _Particle({
    required this.x,
    required this.y,
    required this.size,
    required this.speed,
  });
}

class _ParticlePainter extends CustomPainter {
  final List<_Particle> particles;
  final double progress;

  _ParticlePainter({required this.particles, required this.progress});

  @override
  void paint(Canvas canvas, Size size) {
    for (final p in particles) {
      final adjustedProgress = (progress * p.speed) % 1.0;
      final opacity = (1.0 - (adjustedProgress - 0.5).abs() * 2).clamp(0.0, 1.0);
      final paint = Paint()
        ..color = AppColors.goldStart.withAlpha((opacity * 200).toInt());
      final dx = p.x * size.width;
      final dy = ((p.y + adjustedProgress) % 1.0) * size.height;
      canvas.drawCircle(Offset(dx, dy), p.size * opacity, paint);
    }
  }

  @override
  bool shouldRepaint(_ParticlePainter old) => old.progress != progress;
}
