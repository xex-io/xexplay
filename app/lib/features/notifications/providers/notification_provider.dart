import 'dart:developer' as developer;

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../auth/providers/auth_provider.dart';
import '../services/fcm_service.dart';

// ---------------------------------------------------------------------------
// In-app notification state
// ---------------------------------------------------------------------------

/// Represents a transient in-app notification banner.
class InAppNotification {
  const InAppNotification({
    required this.title,
    required this.body,
    this.data = const {},
  });

  final String? title;
  final String? body;
  final Map<String, dynamic> data;
}

/// A simple state notifier that holds the latest in-app notification.
///
/// UI widgets can watch this and display a banner / snackbar.  After the
/// notification is shown the UI should call [clear].
class InAppNotificationNotifier extends StateNotifier<InAppNotification?> {
  InAppNotificationNotifier() : super(null);

  void show(InAppNotification notification) => state = notification;
  void clear() => state = null;
}

final inAppNotificationProvider =
    StateNotifierProvider<InAppNotificationNotifier, InAppNotification?>(
  (_) => InAppNotificationNotifier(),
);

// ---------------------------------------------------------------------------
// Pending deep-link from notification tap
// ---------------------------------------------------------------------------

/// Holds a route path that was extracted from a tapped notification so the
/// router can navigate to it once the widget tree is ready.
class PendingNotificationRoute extends StateNotifier<String?> {
  PendingNotificationRoute() : super(null);

  void set(String route) => state = route;
  void clear() => state = null;
}

final pendingNotificationRouteProvider =
    StateNotifierProvider<PendingNotificationRoute, String?>(
  (_) => PendingNotificationRoute(),
);

// ---------------------------------------------------------------------------
// FCM service provider
// ---------------------------------------------------------------------------

final fcmServiceProvider = Provider<FcmService>((ref) {
  final dio = ref.watch(apiClientProvider).dio;
  return FcmService(dio);
});

// ---------------------------------------------------------------------------
// Initialisation helper
// ---------------------------------------------------------------------------

/// Call this once during app startup (after the widget tree is available and
/// the [ProviderContainer] / [WidgetRef] is accessible).
///
/// It wires the [FcmService] callbacks into the Riverpod state so that UI
/// widgets can react to notifications via normal provider watching.
Future<void> initNotifications(Ref ref) async {
  final fcm = ref.read(fcmServiceProvider);

  // Wire foreground notifications into Riverpod state.
  fcm.onForegroundNotification = (title, body, data) {
    ref.read(inAppNotificationProvider.notifier).show(
          InAppNotification(title: title, body: body, data: data),
        );
  };

  // Wire notification taps into a pending-route provider so the router can
  // pick it up.
  fcm.onNotificationTap = (route, data) {
    if (route != null) {
      ref.read(pendingNotificationRouteProvider.notifier).set(route);
    }
  };

  try {
    await fcm.init();
    final granted = await fcm.requestPermission();
    if (granted) {
      final token = await fcm.getToken();
      if (token != null) {
        await fcm.registerToken(token);
      }
    }
  } catch (e) {
    developer.log('[Notifications] Init failed: $e');
  }
}
