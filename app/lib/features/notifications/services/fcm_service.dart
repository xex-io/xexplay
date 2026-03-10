import 'dart:developer' as developer;
import 'dart:io' show Platform;

import 'package:dio/dio.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import '../../../core/constants/api_constants.dart';

/// Top-level handler for background/terminated messages.
///
/// Must be a top-level function (not a class method) so the Dart isolate
/// can locate it when the app is not running.
@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
  developer.log(
    '[FcmService] Background message: ${message.messageId}',
  );
}

/// Callback invoked when a notification is tapped (background or terminated).
///
/// The [route] is the deep-link path extracted from the notification data
/// (e.g. `/leaderboard`, `/play/session`). If no route is present the
/// callback is still invoked with `null` so the app can decide a default.
typedef NotificationTapCallback = void Function(String? route, Map<String, dynamic> data);

/// Callback invoked when a foreground notification arrives and should be
/// displayed as an in-app banner.
typedef ForegroundNotificationCallback = void Function(
  String? title,
  String? body,
  Map<String, dynamic> data,
);

/// FCM (Firebase Cloud Messaging) service.
///
/// Handles initialisation, permission requests, token management, and
/// dispatches incoming messages to registered callbacks.
class FcmService {
  FcmService(this._dio);

  final Dio _dio;

  String? _currentToken;

  NotificationTapCallback? onNotificationTap;
  ForegroundNotificationCallback? onForegroundNotification;

  /// Initialise Firebase and wire up all message listeners.
  Future<void> init() async {
    await Firebase.initializeApp();
    developer.log('[FcmService] Firebase initialized');

    // Register the background handler.
    FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);

    // --- Foreground messages ---
    FirebaseMessaging.onMessage.listen(_handleForegroundMessage);

    // --- Notification taps (app was in background) ---
    FirebaseMessaging.onMessageOpenedApp.listen(_handleNotificationTap);

    // --- Notification tap that launched a terminated app ---
    final initialMessage =
        await FirebaseMessaging.instance.getInitialMessage();
    if (initialMessage != null) {
      _handleNotificationTap(initialMessage);
    }

    // Listen for token refresh.
    FirebaseMessaging.instance.onTokenRefresh.listen(onTokenRefresh);

    // Show notification heads-up banners while in foreground on iOS/Android.
    await FirebaseMessaging.instance
        .setForegroundNotificationPresentationOptions(
      alert: true,
      badge: true,
      sound: true,
    );
  }

  /// Request notification permissions from the OS.
  Future<bool> requestPermission() async {
    developer.log('[FcmService] Requesting notification permission');
    final settings = await FirebaseMessaging.instance.requestPermission(
      alert: true,
      badge: true,
      sound: true,
      provisional: false,
    );
    final granted =
        settings.authorizationStatus == AuthorizationStatus.authorized ||
            settings.authorizationStatus == AuthorizationStatus.provisional;
    developer.log('[FcmService] Permission granted: $granted');
    return granted;
  }

  /// Retrieve the current FCM registration token.
  Future<String?> getToken() async {
    _currentToken = await FirebaseMessaging.instance.getToken();
    developer.log('[FcmService] Token: $_currentToken');
    return _currentToken;
  }

  /// Register the device token with the Play backend so push notifications
  /// can be sent from the server.
  Future<void> registerToken(String token) async {
    developer.log('[FcmService] Registering device token with backend');
    try {
      await _dio.post(
        ApiConstants.devicesRegister,
        data: {
          'token': token,
          'platform': _platform,
        },
      );
      _currentToken = token;
      developer.log('[FcmService] Device token registered successfully');
    } on DioException catch (e) {
      developer.log('[FcmService] Failed to register token: $e');
      rethrow;
    }
  }

  /// Called when the FCM token is refreshed. Re-registers with the backend.
  Future<void> onTokenRefresh(String newToken) async {
    developer.log('[FcmService] Token refreshed, re-registering');
    await registerToken(newToken);
  }

  // ---------------------------------------------------------------------------
  // Message handlers
  // ---------------------------------------------------------------------------

  /// Handle a message received while the app is in the foreground.
  void _handleForegroundMessage(RemoteMessage message) {
    developer.log(
      '[FcmService] Foreground message: ${message.messageId}',
    );
    final notification = message.notification;
    if (onForegroundNotification != null) {
      onForegroundNotification!(
        notification?.title,
        notification?.body,
        message.data,
      );
    }
  }

  /// Handle a notification tap that opened the app from background or
  /// terminated state.
  void _handleNotificationTap(RemoteMessage message) {
    developer.log(
      '[FcmService] Notification tapped: ${message.messageId}',
    );
    final route = message.data['route'] as String?;
    if (onNotificationTap != null) {
      onNotificationTap!(route, message.data);
    }
  }

  /// Platform string for the registration payload.
  String get _platform {
    if (kIsWeb) return 'web';
    if (Platform.isIOS) return 'ios';
    if (Platform.isAndroid) return 'android';
    return 'unknown';
  }
}
