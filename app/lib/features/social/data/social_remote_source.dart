import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import 'referral_models.dart';
import 'social_models.dart';

class SocialRemoteSource {
  final ApiClient _apiClient;

  SocialRemoteSource(this._apiClient);

  /// GET /referral/code
  Future<String> getReferralCode() async {
    final response = await _apiClient.dio.get(ApiConstants.referralCode);
    return response.data['code'] as String;
  }

  /// GET /referral/stats
  Future<ReferralStats> getReferralStats() async {
    final response = await _apiClient.dio.get(ApiConstants.referralStats);
    return ReferralStats.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /me/achievements
  Future<List<UserAchievement>> getAchievements() async {
    final response = await _apiClient.dio.get(ApiConstants.meAchievements);
    final list = response.data as List<dynamic>;
    return list
        .map((e) => UserAchievement.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// GET /leagues
  Future<List<League>> getLeagues() async {
    final response = await _apiClient.dio.get(ApiConstants.leagues);
    final list = response.data as List<dynamic>;
    return list
        .map((e) => League.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  /// POST /leagues
  Future<League> createLeague({
    required String name,
    String? eventId,
  }) async {
    final response = await _apiClient.dio.post(
      ApiConstants.leagues,
      data: {
        'name': name,
        // ignore: use_null_aware_elements
        if (eventId != null) 'event_id': eventId,
      },
    );
    return League.fromJson(response.data as Map<String, dynamic>);
  }

  /// POST /leagues/join
  Future<League> joinLeague(String inviteCode) async {
    final response = await _apiClient.dio.post(
      ApiConstants.leaguesJoin,
      data: {'invite_code': inviteCode},
    );
    return League.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /leagues/:id
  Future<League> getLeagueDetail(String id) async {
    final response = await _apiClient.dio.get(ApiConstants.leagueDetail(id));
    return League.fromJson(response.data as Map<String, dynamic>);
  }
}
