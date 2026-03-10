import '../../../core/network/api_client.dart';
import '../../../core/constants/api_constants.dart';
import 'leaderboard_models.dart';

class LeaderboardRemoteSource {
  final ApiClient _apiClient;

  LeaderboardRemoteSource(this._apiClient);

  /// GET /leaderboards/daily
  Future<LeaderboardData> getDaily({int limit = 50, int offset = 0}) async {
    final response = await _apiClient.dio.get(
      ApiConstants.leaderboardDaily,
      queryParameters: {'limit': limit, 'offset': offset},
    );
    return LeaderboardData.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /leaderboards/weekly
  Future<LeaderboardData> getWeekly({int limit = 50, int offset = 0}) async {
    final response = await _apiClient.dio.get(
      ApiConstants.leaderboardWeekly,
      queryParameters: {'limit': limit, 'offset': offset},
    );
    return LeaderboardData.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /leaderboards/tournament/:eventId
  Future<LeaderboardData> getTournament(
    String eventId, {
    int limit = 50,
    int offset = 0,
  }) async {
    final response = await _apiClient.dio.get(
      ApiConstants.leaderboardTournament(eventId),
      queryParameters: {'limit': limit, 'offset': offset},
    );
    return LeaderboardData.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /leaderboards/all-time
  Future<LeaderboardData> getAllTime({int limit = 50, int offset = 0}) async {
    final response = await _apiClient.dio.get(
      ApiConstants.leaderboardAllTime,
      queryParameters: {'limit': limit, 'offset': offset},
    );
    return LeaderboardData.fromJson(response.data as Map<String, dynamic>);
  }
}
