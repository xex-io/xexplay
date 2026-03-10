import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../features/auth/providers/auth_provider.dart';
import '../network/websocket_client.dart';

/// Bootstraps the WebSocket connection when an access token is available.
final wsBootstrapProvider = FutureProvider<WebSocketClient?>((ref) async {
  final storage = ref.watch(secureStorageProvider);
  final token = await storage.getAccessToken();
  if (token == null) return null;

  final client = WebSocketClient(token: token);
  client.connect();

  ref.onDispose(() => client.dispose());

  return client;
});

/// Stream of WebSocket events, powered by the bootstrapped client.
final wsEventsProvider = StreamProvider<WsEvent>((ref) async* {
  final clientAsync = ref.watch(wsBootstrapProvider);
  final client = clientAsync.valueOrNull;
  if (client == null) return;

  yield* client.events;
});

/// Stream of WebSocket connection status changes.
final wsStatusProvider = StreamProvider<WsConnectionStatus>((ref) async* {
  final clientAsync = ref.watch(wsBootstrapProvider);
  final client = clientAsync.valueOrNull;
  if (client == null) return;

  yield* client.status;
});
