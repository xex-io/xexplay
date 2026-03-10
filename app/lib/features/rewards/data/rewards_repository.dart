import 'package:dio/dio.dart';

import 'reward_models.dart';
import 'rewards_remote_source.dart';

class RewardsRepository {
  final RewardsRemoteSource _remoteSource;

  RewardsRepository(this._remoteSource);

  Future<RewardsResponse> getRewards() async {
    return _handleRequest(() => _remoteSource.getRewards());
  }

  Future<void> claimReward(String id) async {
    return _handleRequest(() => _remoteSource.claimReward(id));
  }

  /// Wraps remote calls with consistent error handling.
  Future<T> _handleRequest<T>(Future<T> Function() request) async {
    try {
      return await request();
    } on DioException catch (e) {
      final message = _extractErrorMessage(e);
      throw RewardsException(message);
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

class RewardsException implements Exception {
  final String message;
  const RewardsException(this.message);

  @override
  String toString() => message;
}
