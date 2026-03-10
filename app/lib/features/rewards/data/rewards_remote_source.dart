import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import 'reward_models.dart';

class RewardsRemoteSource {
  final ApiClient _apiClient;

  RewardsRemoteSource(this._apiClient);

  /// GET /me/rewards — fetch pending rewards, history, and streak info.
  Future<RewardsResponse> getRewards() async {
    final response = await _apiClient.dio.get(ApiConstants.meRewards);
    return RewardsResponse.fromJson(response.data as Map<String, dynamic>);
  }

  /// POST /me/rewards/{id}/claim — claim a specific reward.
  Future<void> claimReward(String id) async {
    await _apiClient.dio.post(ApiConstants.meRewardClaimById(id));
  }
}
