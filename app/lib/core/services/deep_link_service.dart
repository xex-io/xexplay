import 'dart:developer' as developer;

import 'package:flutter/material.dart';

/// Deep link handling service for XEX Play.
///
/// Supports:
/// - Referral deep links: `xexplay://referral?code=ABC123`
/// - Shared content links: `https://play.xex.exchange/share/...`
///
/// Integration notes:
/// - Add `app_links` package for production link handling.
/// - On Android, configure intent-filters in AndroidManifest.xml.
/// - On iOS, configure Associated Domains in the entitlements file.
class DeepLinkService {
  DeepLinkService({this.navigatorKey});

  final GlobalKey<NavigatorState>? navigatorKey;

  /// Initialise the deep link listener.
  ///
  /// Call once from the app's top-level widget (e.g. `initState` of the App).
  Future<void> init() async {
    developer.log('[DeepLinkService] Initialising deep link listener');

    // TODO: With app_links package:
    // final appLinks = AppLinks();
    // appLinks.uriLinkStream.listen(_handleUri);
    //
    // // Handle the link that launched the app (cold start).
    // final initialUri = await appLinks.getInitialLink();
    // if (initialUri != null) _handleUri(initialUri);
  }

  /// Process an incoming URI and route to the appropriate screen.
  void handleUri(Uri uri) {
    developer.log('[DeepLinkService] Received deep link: $uri');

    final scheme = uri.scheme; // xexplay or https
    final host = uri.host; // e.g. referral, share, play.xex.exchange

    // --- Custom scheme: xexplay://referral?code=ABC123 ---
    if (scheme == 'xexplay' && host == 'referral') {
      final code = uri.queryParameters['code'];
      if (code != null && code.isNotEmpty) {
        _handleReferral(code);
        return;
      }
    }

    // --- Universal link: https://play.xex.exchange/share/{id} ---
    if (host == 'play.xex.exchange') {
      final segments = uri.pathSegments;
      if (segments.isNotEmpty && segments.first == 'share') {
        final shareId = segments.length > 1 ? segments[1] : null;
        if (shareId != null) {
          _handleSharedContent(shareId);
          return;
        }
      }
    }

    developer.log('[DeepLinkService] Unhandled deep link: $uri');
  }

  /// Apply a referral code received via deep link.
  void _handleReferral(String code) {
    developer.log('[DeepLinkService] Referral code received: $code');
    // TODO: store referral code and apply after auth
    // navigatorKey?.currentState?.pushNamed('/referral', arguments: code);
  }

  /// Navigate to shared content.
  void _handleSharedContent(String shareId) {
    developer.log('[DeepLinkService] Shared content id: $shareId');
    // TODO: navigate to the shared content screen
  }
}
