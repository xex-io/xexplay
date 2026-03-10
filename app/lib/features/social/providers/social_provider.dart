import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../auth/providers/auth_provider.dart';
import '../data/referral_models.dart';
import '../data/social_models.dart';
import '../data/social_remote_source.dart';

// ---------------------------------------------------------------------------
// Infrastructure
// ---------------------------------------------------------------------------

final socialRemoteSourceProvider = Provider<SocialRemoteSource>((ref) {
  return SocialRemoteSource(ref.watch(apiClientProvider));
});

// ---------------------------------------------------------------------------
// Referral providers
// ---------------------------------------------------------------------------

final referralCodeProvider = FutureProvider<String>((ref) {
  return ref.watch(socialRemoteSourceProvider).getReferralCode();
});

final referralStatsProvider = FutureProvider<ReferralStats>((ref) {
  return ref.watch(socialRemoteSourceProvider).getReferralStats();
});

// ---------------------------------------------------------------------------
// Achievements provider
// ---------------------------------------------------------------------------

final achievementsProvider = FutureProvider<List<UserAchievement>>((ref) {
  return ref.watch(socialRemoteSourceProvider).getAchievements();
});

// ---------------------------------------------------------------------------
// Leagues providers
// ---------------------------------------------------------------------------

final leaguesProvider = FutureProvider<List<League>>((ref) {
  return ref.watch(socialRemoteSourceProvider).getLeagues();
});

final leagueDetailProvider =
    FutureProvider.family<League, String>((ref, id) {
  return ref.watch(socialRemoteSourceProvider).getLeagueDetail(id);
});
