import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:web_socket_channel/web_socket_channel.dart';

import '../constants/api_constants.dart';

/// A typed WebSocket event parsed from the server JSON messages.
class WsEvent {
  final String type;
  final Map<String, dynamic> data;

  const WsEvent({required this.type, required this.data});

  factory WsEvent.fromJson(Map<String, dynamic> json) {
    return WsEvent(
      type: json['type'] as String? ?? 'unknown',
      data: json['data'] as Map<String, dynamic>? ?? const {},
    );
  }
}

/// WebSocket client with auto-reconnect and typed event stream.
class WebSocketClient {
  WebSocketClient({required this.token});

  final String token;

  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _subscription;
  Timer? _reconnectTimer;

  final _controller = StreamController<WsEvent>.broadcast();
  final _statusController = StreamController<WsConnectionStatus>.broadcast();

  int _retryCount = 0;
  static const int _maxRetryCount = 10;
  bool _disposed = false;

  /// Stream of parsed WebSocket events.
  Stream<WsEvent> get events => _controller.stream;

  /// Stream of connection status changes.
  Stream<WsConnectionStatus> get status => _statusController.stream;

  /// Build the WebSocket URL from the base API URL.
  String get _wsUrl {
    final base = ApiConstants.baseUrl
        .replaceFirst('http://', 'ws://')
        .replaceFirst('https://', 'wss://');
    return '$base/ws?token=$token';
  }

  /// Connect to the WebSocket server.
  void connect() {
    if (_disposed) return;
    _disconnect();

    _statusController.add(WsConnectionStatus.connecting);

    try {
      _channel = WebSocketChannel.connect(Uri.parse(_wsUrl));
      _retryCount = 0;
      _statusController.add(WsConnectionStatus.connected);

      _subscription = _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
        cancelOnError: false,
      );
    } catch (e) {
      _scheduleReconnect();
    }
  }

  void _onMessage(dynamic raw) {
    try {
      final json = jsonDecode(raw as String) as Map<String, dynamic>;
      _controller.add(WsEvent.fromJson(json));
    } catch (_) {
      // Ignore malformed messages.
    }
  }

  void _onError(Object error) {
    _scheduleReconnect();
  }

  void _onDone() {
    _statusController.add(WsConnectionStatus.disconnected);
    _scheduleReconnect();
  }

  /// Exponential backoff: 1s, 2s, 4s, 8s ... capped at 30s.
  void _scheduleReconnect() {
    if (_disposed || _retryCount >= _maxRetryCount) return;

    final delay = Duration(
      milliseconds: min(1000 * pow(2, _retryCount).toInt(), 30000),
    );
    _retryCount++;

    _statusController.add(WsConnectionStatus.reconnecting);
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(delay, connect);
  }

  void _disconnect() {
    _subscription?.cancel();
    _subscription = null;
    _channel?.sink.close();
    _channel = null;
  }

  /// Send a JSON message to the server.
  void send(Map<String, dynamic> message) {
    _channel?.sink.add(jsonEncode(message));
  }

  /// Permanently dispose the client and close all streams.
  void dispose() {
    _disposed = true;
    _reconnectTimer?.cancel();
    _disconnect();
    _controller.close();
    _statusController.close();
  }
}

enum WsConnectionStatus {
  connecting,
  connected,
  disconnected,
  reconnecting,
}
