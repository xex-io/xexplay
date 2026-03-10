import 'dart:developer' as developer;

import 'package:firebase_crashlytics/firebase_crashlytics.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Wraps Firebase Crashlytics with convenience helpers.
///
/// Call [init] once at app startup (before `runApp`) to wire the global
/// Flutter and platform error handlers.
class CrashlyticsService {
  CrashlyticsService({FirebaseCrashlytics? crashlytics})
      : _crashlytics = crashlytics ?? FirebaseCrashlytics.instance;

  final FirebaseCrashlytics _crashlytics;

  // ---------------------------------------------------------------------------
  // Initialisation
  // ---------------------------------------------------------------------------

  /// Wires Crashlytics into Flutter's global error handlers.
  ///
  /// * Sets [FlutterError.onError] to forward framework-level errors.
  /// * Registers a [PlatformDispatcher] error callback for async errors.
  /// * Disables collection in debug mode so dev crashes stay local.
  Future<void> init() async {
    // Disable crash reporting when running in debug to keep the console clean.
    await _crashlytics.setCrashlyticsCollectionEnabled(!kDebugMode);

    // Capture Flutter framework errors.
    FlutterError.onError = _crashlytics.recordFlutterFatalError;

    // Capture asynchronous errors not caught by the Flutter framework.
    PlatformDispatcher.instance.onError = (error, stack) {
      _crashlytics.recordError(error, stack, fatal: true);
      return true;
    };

    developer.log('[CrashlyticsService] Initialised');
  }

  // ---------------------------------------------------------------------------
  // User identification
  // ---------------------------------------------------------------------------

  /// Sets the user identifier that appears alongside crash reports.
  Future<void> setUserIdentifier(String userId) async {
    try {
      await _crashlytics.setUserIdentifier(userId);
    } catch (e, st) {
      developer.log(
        '[CrashlyticsService] Failed to set user identifier',
        error: e,
        stackTrace: st,
      );
    }
  }

  /// Clears the user identifier (e.g. on logout).
  Future<void> clearUserIdentifier() => setUserIdentifier('');

  // ---------------------------------------------------------------------------
  // Error recording
  // ---------------------------------------------------------------------------

  /// Records a non-fatal error.
  ///
  /// Use this for caught exceptions that should still surface in the
  /// Crashlytics dashboard so you can track their frequency.
  Future<void> recordError(
    dynamic exception,
    StackTrace? stackTrace, {
    String? reason,
    bool fatal = false,
  }) async {
    try {
      await _crashlytics.recordError(
        exception,
        stackTrace,
        reason: reason ?? 'non-fatal',
        fatal: fatal,
      );
    } catch (e, st) {
      developer.log(
        '[CrashlyticsService] Failed to record error',
        error: e,
        stackTrace: st,
      );
    }
  }

  /// Adds a breadcrumb-style log message to the next crash report.
  Future<void> log(String message) async {
    try {
      await _crashlytics.log(message);
    } catch (e, st) {
      developer.log(
        '[CrashlyticsService] Failed to log message',
        error: e,
        stackTrace: st,
      );
    }
  }

  // ---------------------------------------------------------------------------
  // Custom keys
  // ---------------------------------------------------------------------------

  /// Sets a custom key/value pair on the crash report.
  Future<void> setCustomKey(String key, Object value) async {
    try {
      await _crashlytics.setCustomKey(key, value);
    } catch (e, st) {
      developer.log(
        '[CrashlyticsService] Failed to set custom key "$key"',
        error: e,
        stackTrace: st,
      );
    }
  }
}

// ---------------------------------------------------------------------------
// Riverpod provider
// ---------------------------------------------------------------------------

final crashlyticsServiceProvider = Provider<CrashlyticsService>((_) {
  return CrashlyticsService();
});
