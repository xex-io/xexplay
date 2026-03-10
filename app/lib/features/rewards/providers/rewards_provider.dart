import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../auth/providers/auth_provider.dart';
import '../data/reward_models.dart';
import '../data/rewards_remote_source.dart';
import '../data/rewards_repository.dart';

// ---------------------------------------------------------------------------
// Infrastructure providers
// ---------------------------------------------------------------------------

final rewardsRemoteSourceProvider = Provider<RewardsRemoteSource>((ref) {
  return RewardsRemoteSource(ref.watch(apiClientProvider));
});

final rewardsRepositoryProvider = Provider<RewardsRepository>((ref) {
  return RewardsRepository(ref.watch(rewardsRemoteSourceProvider));
});

// ---------------------------------------------------------------------------
// Rewards data provider
// ---------------------------------------------------------------------------

/// Fetches rewards data (pending, history, streak) from the API.
final rewardsProvider = FutureProvider.autoDispose<RewardsResponse>((ref) async {
  final repository = ref.watch(rewardsRepositoryProvider);
  return repository.getRewards();
});

// ---------------------------------------------------------------------------
// Claim action
// ---------------------------------------------------------------------------

/// Claims a reward by id and refreshes the rewards list.
Future<void> claimReward(WidgetRef ref, String rewardId) async {
  final repository = ref.read(rewardsRepositoryProvider);
  await repository.claimReward(rewardId);
  ref.invalidate(rewardsProvider);
}
