// Data models for the rewards feature.

class StreakInfo {
  final int currentStreak;
  final int longestStreak;
  final String? lastPlayedDate;
  final int bonusSkips;
  final int bonusAnswers;

  const StreakInfo({
    required this.currentStreak,
    required this.longestStreak,
    this.lastPlayedDate,
    required this.bonusSkips,
    required this.bonusAnswers,
  });

  factory StreakInfo.fromJson(Map<String, dynamic> json) {
    return StreakInfo(
      currentStreak: json['current_streak'] as int? ?? 0,
      longestStreak: json['longest_streak'] as int? ?? 0,
      lastPlayedDate: json['last_played_date'] as String?,
      bonusSkips: json['bonus_skips'] as int? ?? 0,
      bonusAnswers: json['bonus_answers'] as int? ?? 0,
    );
  }
}

class RewardItem {
  final String id;
  final String periodType;
  final String periodKey;
  final String rewardType;
  final double amount;
  final int? rank;
  final String status;
  final String? claimedAt;
  final String createdAt;

  const RewardItem({
    required this.id,
    required this.periodType,
    required this.periodKey,
    required this.rewardType,
    required this.amount,
    this.rank,
    required this.status,
    this.claimedAt,
    required this.createdAt,
  });

  bool get isPending => status == 'pending';
  bool get isClaimed => status == 'claimed';

  factory RewardItem.fromJson(Map<String, dynamic> json) {
    return RewardItem(
      id: json['id'] as String,
      periodType: json['period_type'] as String,
      periodKey: json['period_key'] as String,
      rewardType: json['reward_type'] as String,
      amount: (json['amount'] as num).toDouble(),
      rank: json['rank'] as int?,
      status: json['status'] as String,
      claimedAt: json['claimed_at'] as String?,
      createdAt: json['created_at'] as String,
    );
  }
}

class RewardsResponse {
  final List<RewardItem> pending;
  final List<RewardItem> history;
  final StreakInfo? streak;

  const RewardsResponse({
    required this.pending,
    required this.history,
    this.streak,
  });

  factory RewardsResponse.fromJson(Map<String, dynamic> json) {
    return RewardsResponse(
      pending: (json['pending'] as List<dynamic>?)
              ?.map((e) => RewardItem.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      history: (json['history'] as List<dynamic>?)
              ?.map((e) => RewardItem.fromJson(e as Map<String, dynamic>))
              .toList() ??
          [],
      streak: json['streak'] != null
          ? StreakInfo.fromJson(json['streak'] as Map<String, dynamic>)
          : null,
    );
  }
}
