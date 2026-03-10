import 'package:dio/dio.dart';

import 'card_models.dart';
import 'game_remote_source.dart';

class GameRepository {
  final GameRemoteSource _remoteSource;

  GameRepository(this._remoteSource);

  Future<SessionModel> startSession() async {
    return _handleRequest(() => _remoteSource.startSession());
  }

  Future<SessionModel> getSession() async {
    return _handleRequest(() => _remoteSource.getSession());
  }

  Future<SessionCardModel> getCurrentCard() async {
    return _handleRequest(() => _remoteSource.getCurrentCard());
  }

  Future<AnswerResultModel> submitAnswer({
    required String cardId,
    required bool answer,
  }) async {
    return _handleRequest(
      () => _remoteSource.submitAnswer(cardId: cardId, answer: answer),
    );
  }

  Future<AnswerResultModel> skipCard({required String cardId}) async {
    return _handleRequest(
      () => _remoteSource.skipCard(cardId: cardId),
    );
  }

  Future<DailyStatusModel> getDailyStatus() async {
    return _handleRequest(() => _remoteSource.getDailyStatus());
  }

  /// Wraps remote calls with consistent error handling.
  Future<T> _handleRequest<T>(Future<T> Function() request) async {
    try {
      return await request();
    } on DioException catch (e) {
      final message = _extractErrorMessage(e);
      throw GameException(message);
    }
  }

  String _extractErrorMessage(DioException e) {
    // Try to extract server error message from response body.
    final data = e.response?.data;
    if (data is Map<String, dynamic> && data.containsKey('message')) {
      return data['message'] as String;
    }

    return switch (e.type) {
      DioExceptionType.connectionTimeout ||
      DioExceptionType.sendTimeout ||
      DioExceptionType.receiveTimeout =>
        'Connection timed out. Please try again.',
      DioExceptionType.connectionError =>
        'No internet connection. Please check your network.',
      DioExceptionType.badResponse => 'Server error (${e.response?.statusCode}).',
      _ => 'Something went wrong. Please try again.',
    };
  }
}

class GameException implements Exception {
  final String message;
  const GameException(this.message);

  @override
  String toString() => message;
}
