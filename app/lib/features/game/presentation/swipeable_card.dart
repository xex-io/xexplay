import 'dart:math';
import 'package:flutter/material.dart';
import 'card_widget.dart';

/// A widget that wraps [PredictionCardWidget] with swipe gesture handling.
///
/// Supports three swipe directions:
/// - Right → Yes
/// - Left  → No
/// - Up    → Skip (can be disabled via [canSkip])
class SwipeableCard extends StatefulWidget {
  const SwipeableCard({
    super.key,
    required this.tier,
    required this.question,
    required this.points,
    required this.onSwipeRight,
    required this.onSwipeLeft,
    required this.onSwipeUp,
    this.canSkip = true,
  });

  final CardTier tier;
  final String question;
  final int points;
  final VoidCallback onSwipeRight;
  final VoidCallback onSwipeLeft;
  final VoidCallback onSwipeUp;
  final bool canSkip;

  @override
  State<SwipeableCard> createState() => _SwipeableCardState();
}

class _SwipeableCardState extends State<SwipeableCard>
    with SingleTickerProviderStateMixin {
  Offset _dragOffset = Offset.zero;
  bool _isFlyingOff = false;

  late AnimationController _animationController;
  Animation<Offset>? _animation;

  static const double _horizontalThreshold = 100.0;
  static const double _verticalThreshold = 80.0;
  static const double _maxRotation = 0.3; // radians (~17 degrees)

  @override
  void initState() {
    super.initState();
    _animationController = AnimationController(vsync: this);
    _animationController.addListener(() {
      if (_animation != null) {
        setState(() {
          _dragOffset = _animation!.value;
        });
      }
    });
    _animationController.addStatusListener((status) {
      if (status == AnimationStatus.completed) {
        if (_isFlyingOff) {
          _completeFlyOff();
        }
      }
    });
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  void _onPanStart(DragStartDetails details) {
    if (_isFlyingOff) return;
    _animationController.stop();
  }

  void _onPanUpdate(DragUpdateDetails details) {
    if (_isFlyingOff) return;
    setState(() {
      double dy = _dragOffset.dy + details.delta.dy;
      // If skip is disabled, prevent upward drag
      if (!widget.canSkip && dy < 0) {
        dy = 0;
      }
      // Only allow upward swipe (negative dy), clamp downward
      if (dy > 0) dy = 0;
      _dragOffset = Offset(
        _dragOffset.dx + details.delta.dx,
        dy,
      );
    });
  }

  void _onPanEnd(DragEndDetails details) {
    if (_isFlyingOff) return;

    final dx = _dragOffset.dx;
    final dy = _dragOffset.dy;
    final velocity = details.velocity.pixelsPerSecond;

    // Check if swipe exceeds threshold
    if (dx.abs() > _horizontalThreshold) {
      _flyOff(dx > 0 ? _SwipeDirection.right : _SwipeDirection.left, velocity);
    } else if (dy < -_verticalThreshold && widget.canSkip) {
      _flyOff(_SwipeDirection.up, velocity);
    } else {
      _springBack();
    }
  }

  void _flyOff(_SwipeDirection direction, Offset velocity) {
    _isFlyingOff = true;

    final screenSize = MediaQuery.of(context).size;
    late Offset target;

    switch (direction) {
      case _SwipeDirection.right:
        target = Offset(screenSize.width * 1.5, _dragOffset.dy);
      case _SwipeDirection.left:
        target = Offset(-screenSize.width * 1.5, _dragOffset.dy);
      case _SwipeDirection.up:
        target = Offset(_dragOffset.dx, -screenSize.height);
    }

    _animation = Tween<Offset>(
      begin: _dragOffset,
      end: target,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.easeOut,
    ));

    // Use velocity to determine duration — faster swipe = shorter duration
    final speed = max(velocity.distance, 800.0);
    final distance = (target - _dragOffset).distance;
    final duration = Duration(
      milliseconds: (distance / speed * 1000).clamp(200, 500).toInt(),
    );

    _animationController.duration = duration;
    _animationController.forward(from: 0.0);

    // Store direction for callback
    _pendingDirection = direction;
  }

  _SwipeDirection? _pendingDirection;

  void _completeFlyOff() {
    final direction = _pendingDirection;
    _pendingDirection = null;
    _isFlyingOff = false;
    _dragOffset = Offset.zero;

    switch (direction) {
      case _SwipeDirection.right:
        widget.onSwipeRight();
      case _SwipeDirection.left:
        widget.onSwipeLeft();
      case _SwipeDirection.up:
        widget.onSwipeUp();
      case null:
        break;
    }
  }

  void _springBack() {
    _animation = Tween<Offset>(
      begin: _dragOffset,
      end: Offset.zero,
    ).animate(CurvedAnimation(
      parent: _animationController,
      curve: Curves.elasticOut,
    ));

    _animationController.duration = const Duration(milliseconds: 600);
    _animationController.forward(from: 0.0);
  }

  double get _rotation {
    // Rotate proportionally to horizontal drag
    final screenWidth = MediaQuery.of(context).size.width;
    return (_dragOffset.dx / screenWidth) * _maxRotation;
  }

  double get _yesOpacity {
    return (_dragOffset.dx / _horizontalThreshold).clamp(0.0, 1.0);
  }

  double get _noOpacity {
    return (-_dragOffset.dx / _horizontalThreshold).clamp(0.0, 1.0);
  }

  double get _skipOpacity {
    if (!widget.canSkip) return 0.0;
    return (-_dragOffset.dy / _verticalThreshold).clamp(0.0, 1.0);
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onPanStart: _onPanStart,
      onPanUpdate: _onPanUpdate,
      onPanEnd: _onPanEnd,
      child: Transform.translate(
        offset: _dragOffset,
        child: Transform.rotate(
          angle: _rotation,
          child: Stack(
            alignment: Alignment.center,
            children: [
              // The card itself
              PredictionCardWidget(
                tier: widget.tier,
                question: widget.question,
                points: widget.points,
              ),

              // YES overlay (right swipe)
              if (_yesOpacity > 0)
                Positioned.fill(
                  child: _SwipeOverlay(
                    label: 'YES',
                    color: Colors.green,
                    opacity: _yesOpacity,
                    alignment: Alignment.centerLeft,
                  ),
                ),

              // NO overlay (left swipe)
              if (_noOpacity > 0)
                Positioned.fill(
                  child: _SwipeOverlay(
                    label: 'NO',
                    color: Colors.red,
                    opacity: _noOpacity,
                    alignment: Alignment.centerRight,
                  ),
                ),

              // SKIP overlay (up swipe)
              if (_skipOpacity > 0)
                Positioned.fill(
                  child: _SwipeOverlay(
                    label: 'SKIP',
                    color: Colors.grey,
                    opacity: _skipOpacity,
                    alignment: Alignment.bottomCenter,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }
}

enum _SwipeDirection { left, right, up }

class _SwipeOverlay extends StatelessWidget {
  const _SwipeOverlay({
    required this.label,
    required this.color,
    required this.opacity,
    required this.alignment,
  });

  final String label;
  final Color color;
  final double opacity;
  final Alignment alignment;

  @override
  Widget build(BuildContext context) {
    return Align(
      alignment: alignment,
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Opacity(
          opacity: opacity,
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            decoration: BoxDecoration(
              border: Border.all(color: color, width: 3),
              borderRadius: BorderRadius.circular(12),
              color: color.withAlpha(30),
            ),
            child: Text(
              label,
              style: TextStyle(
                fontSize: 28,
                fontWeight: FontWeight.w900,
                color: color,
                letterSpacing: 2,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
