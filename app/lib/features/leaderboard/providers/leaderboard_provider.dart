import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../auth/providers/auth_provider.dart';
import '../data/leaderboard_models.dart';
import '../data/leaderboard_remote_source.dart';
import '../data/leaderboard_repository.dart';

// ---------------------------------------------------------------------------
// Infrastructure providers
// ---------------------------------------------------------------------------

final leaderboardRemoteSourceProvider =
    Provider<LeaderboardRemoteSource>((ref) {
  return LeaderboardRemoteSource(ref.watch(apiClientProvider));
});

final leaderboardRepositoryProvider = Provider<LeaderboardRepository>((ref) {
  return LeaderboardRepository(ref.watch(leaderboardRemoteSourceProvider));
});

// ---------------------------------------------------------------------------
// UI state providers
// ---------------------------------------------------------------------------

/// The currently selected leaderboard tab.
final selectedLeaderboardTypeProvider =
    StateProvider<String>((_) => 'daily');

/// Parameters for fetching a leaderboard page.
class LeaderboardParams {
  final String type;
  final String? eventId;
  final int limit;
  final int offset;

  const LeaderboardParams({
    required this.type,
    this.eventId,
    this.limit = 50,
    this.offset = 0,
  });

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is LeaderboardParams &&
          type == other.type &&
          eventId == other.eventId &&
          limit == other.limit &&
          offset == other.offset;

  @override
  int get hashCode => Object.hash(type, eventId, limit, offset);
}

/// FutureProvider.family that fetches leaderboard data based on [LeaderboardParams].
final leaderboardDataProvider =
    FutureProvider.family<LeaderboardData, LeaderboardParams>((ref, params) {
  final repository = ref.watch(leaderboardRepositoryProvider);

  return switch (params.type) {
    'daily' => repository.getDaily(
        limit: params.limit,
        offset: params.offset,
      ),
    'weekly' => repository.getWeekly(
        limit: params.limit,
        offset: params.offset,
      ),
    'tournament' => repository.getTournament(
        params.eventId ?? '',
        limit: params.limit,
        offset: params.offset,
      ),
    'all-time' => repository.getAllTime(
        limit: params.limit,
        offset: params.offset,
      ),
    'friends' => repository.getDaily(
        limit: params.limit,
        offset: params.offset,
        // TODO: add repository.getFriends() when backend supports it
      ),
    _ => repository.getDaily(
        limit: params.limit,
        offset: params.offset,
      ),
  };
});
