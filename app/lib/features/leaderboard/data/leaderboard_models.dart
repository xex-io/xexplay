class LeaderboardEntry {
  final int rank;
  final String userId;
  final String displayName;
  final String? avatarUrl;
  final int totalPoints;
  final int correctAnswers;

  const LeaderboardEntry({
    required this.rank,
    required this.userId,
    required this.displayName,
    this.avatarUrl,
    required this.totalPoints,
    required this.correctAnswers,
  });

  factory LeaderboardEntry.fromJson(Map<String, dynamic> json) {
    return LeaderboardEntry(
      rank: json['rank'] as int,
      userId: json['user_id'] as String,
      displayName: json['display_name'] as String,
      avatarUrl: json['avatar_url'] as String?,
      totalPoints: json['total_points'] as int,
      correctAnswers: json['correct_answers'] as int,
    );
  }
}

class LeaderboardData {
  final String periodType;
  final String periodKey;
  final List<LeaderboardEntry> entries;
  final LeaderboardEntry? userRank;
  final int total;

  const LeaderboardData({
    required this.periodType,
    required this.periodKey,
    required this.entries,
    this.userRank,
    required this.total,
  });

  factory LeaderboardData.fromJson(Map<String, dynamic> json) {
    return LeaderboardData(
      periodType: json['period_type'] as String,
      periodKey: json['period_key'] as String,
      entries: (json['entries'] as List<dynamic>)
          .map((e) => LeaderboardEntry.fromJson(e as Map<String, dynamic>))
          .toList(),
      userRank: json['user_rank'] != null
          ? LeaderboardEntry.fromJson(json['user_rank'] as Map<String, dynamic>)
          : null,
      total: json['total'] as int,
    );
  }
}
