import 'dart:developer' as developer;

import 'package:dio/dio.dart';

import '../../../core/constants/api_constants.dart';

/// FCM (Firebase Cloud Messaging) service placeholder.
///
/// Actual Firebase initialization requires native platform configuration
/// (google-services.json for Android, GoogleService-Info.plist for iOS).
/// This class defines the interface and registers the device token with the
/// backend. Replace the placeholder logging with real Firebase calls once
/// native config is in place.
class FcmService {
  FcmService(this._dio);

  final Dio _dio;

  String? _currentToken;

  /// Initialise the FCM service.
  ///
  /// In production this would call `Firebase.initializeApp()` and
  /// `FirebaseMessaging.instance`.
  Future<void> init() async {
    developer.log('[FcmService] FCM would be initialized here');
    // TODO: Firebase.initializeApp();
    // TODO: listen for onTokenRefresh and call _onTokenRefresh
  }

  /// Request notification permissions from the OS.
  Future<bool> requestPermission() async {
    developer.log('[FcmService] Requesting notification permission');
    // TODO: FirebaseMessaging.instance.requestPermission()
    return true;
  }

  /// Retrieve the current FCM registration token.
  Future<String?> getToken() async {
    developer.log('[FcmService] Getting FCM token');
    // TODO: _currentToken = await FirebaseMessaging.instance.getToken();
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
  ///
  /// Wire this up as the callback for `FirebaseMessaging.instance.onTokenRefresh`.
  // ignore: unused_element
  Future<void> onTokenRefresh(String newToken) async {
    developer.log('[FcmService] Token refreshed, re-registering');
    await registerToken(newToken);
  }

  /// Simple platform detection string for the registration payload.
  String get _platform {
    // In real code you would use `Platform.isIOS` etc.
    return 'unknown';
  }
}
