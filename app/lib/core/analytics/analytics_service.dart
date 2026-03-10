import 'dart:developer' as developer;

import 'package:firebase_analytics/firebase_analytics.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Wraps Firebase Analytics with domain-specific event helpers.
///
/// All public methods are fire-and-forget; errors are logged but never thrown
/// so analytics failures cannot break the UI.
class AnalyticsService {
  AnalyticsService({FirebaseAnalytics? analytics})
      : _analytics = analytics ?? FirebaseAnalytics.instance;

  final FirebaseAnalytics _analytics;

  /// Returns the [NavigatorObserver] that auto-tracks screen transitions.
  FirebaseAnalyticsObserver get observer =>
      FirebaseAnalyticsObserver(analytics: _analytics);

  // ---------------------------------------------------------------------------
  // Session events
  // ---------------------------------------------------------------------------

  /// Logs the start of a new game session.
  Future<void> logSessionStart() => _log('session_start');

  /// Logs completion of a game session with final results.
  Future<void> logSessionComplete({
    required int score,
    required int correctCount,
  }) =>
      _log('session_complete', parameters: {
        'score': score,
        'correct_count': correctCount,
      });

  // ---------------------------------------------------------------------------
  // Gameplay events
  // ---------------------------------------------------------------------------

  /// Logs a single card being answered.
  Future<void> logCardAnswered({
    required String tier,
    required bool isCorrect,
  }) =>
      _log('card_answered', parameters: {
        'tier': tier,
        'is_correct': isCorrect.toString(),
      });

  // ---------------------------------------------------------------------------
  // Reward events
  // ---------------------------------------------------------------------------

  /// Logs the user claiming a reward.
  Future<void> logRewardClaimed({required double amount}) =>
      _log('reward_claimed', parameters: {
        'amount': amount,
      });

  // ---------------------------------------------------------------------------
  // Exchange funnel events
  // ---------------------------------------------------------------------------

  /// Logs the user tapping the "Go to Exchange" prompt.
  Future<void> logExchangePromptTapped() => _log('exchange_prompt_tapped');

  // ---------------------------------------------------------------------------
  // Navigation
  // ---------------------------------------------------------------------------

  /// Manually logs a screen view (for cases the observer cannot cover).
  Future<void> logScreenView({required String screenName}) async {
    try {
      await _analytics.logScreenView(screenName: screenName);
    } catch (e, st) {
      developer.log(
        '[AnalyticsService] Failed to log screen view',
        error: e,
        stackTrace: st,
      );
    }
  }

  // ---------------------------------------------------------------------------
  // Internals
  // ---------------------------------------------------------------------------

  Future<void> _log(
    String name, {
    Map<String, Object>? parameters,
  }) async {
    try {
      await _analytics.logEvent(name: name, parameters: parameters);
    } catch (e, st) {
      developer.log(
        '[AnalyticsService] Failed to log event "$name"',
        error: e,
        stackTrace: st,
      );
    }
  }
}

// ---------------------------------------------------------------------------
// Riverpod provider
// ---------------------------------------------------------------------------

final analyticsServiceProvider = Provider<AnalyticsService>((_) {
  return AnalyticsService();
});
