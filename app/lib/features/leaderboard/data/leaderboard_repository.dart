import 'package:dio/dio.dart';

import 'leaderboard_models.dart';
import 'leaderboard_remote_source.dart';

class LeaderboardRepository {
  final LeaderboardRemoteSource _remoteSource;

  LeaderboardRepository(this._remoteSource);

  Future<LeaderboardData> getDaily({int limit = 50, int offset = 0}) async {
    return _handleRequest(
      () => _remoteSource.getDaily(limit: limit, offset: offset),
    );
  }

  Future<LeaderboardData> getWeekly({int limit = 50, int offset = 0}) async {
    return _handleRequest(
      () => _remoteSource.getWeekly(limit: limit, offset: offset),
    );
  }

  Future<LeaderboardData> getTournament(
    String eventId, {
    int limit = 50,
    int offset = 0,
  }) async {
    return _handleRequest(
      () => _remoteSource.getTournament(eventId, limit: limit, offset: offset),
    );
  }

  Future<LeaderboardData> getAllTime({int limit = 50, int offset = 0}) async {
    return _handleRequest(
      () => _remoteSource.getAllTime(limit: limit, offset: offset),
    );
  }

  /// Wraps remote calls with consistent error handling.
  Future<T> _handleRequest<T>(Future<T> Function() request) async {
    try {
      return await request();
    } on DioException catch (e) {
      final message = _extractErrorMessage(e);
      throw LeaderboardException(message);
    }
  }

  String _extractErrorMessage(DioException e) {
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
      DioExceptionType.badResponse =>
        'Server error (${e.response?.statusCode}).',
      _ => 'Something went wrong. Please try again.',
    };
  }
}

class LeaderboardException implements Exception {
  final String message;
  const LeaderboardException(this.message);

  @override
  String toString() => message;
}
