class Achievement {
  final String id;
  final String name;
  final String description;
  final String iconUrl;
  final String category;
  final int requiredValue;

  const Achievement({
    required this.id,
    required this.name,
    required this.description,
    required this.iconUrl,
    required this.category,
    required this.requiredValue,
  });

  factory Achievement.fromJson(Map<String, dynamic> json) {
    return Achievement(
      id: json['id'] as String,
      name: json['name'] as String,
      description: json['description'] as String,
      iconUrl: json['icon_url'] as String,
      category: json['category'] as String,
      requiredValue: json['required_value'] as int,
    );
  }
}

class UserAchievement {
  final Achievement achievement;
  final bool earned;
  final DateTime? earnedAt;
  final int progress;

  const UserAchievement({
    required this.achievement,
    required this.earned,
    this.earnedAt,
    required this.progress,
  });

  factory UserAchievement.fromJson(Map<String, dynamic> json) {
    return UserAchievement(
      achievement:
          Achievement.fromJson(json['achievement'] as Map<String, dynamic>),
      earned: json['earned'] as bool,
      earnedAt: json['earned_at'] != null
          ? DateTime.parse(json['earned_at'] as String)
          : null,
      progress: json['progress'] as int,
    );
  }
}

class LeagueMember {
  final String userId;
  final String displayName;
  final String? avatarUrl;
  final int totalPoints;
  final int rank;

  const LeagueMember({
    required this.userId,
    required this.displayName,
    this.avatarUrl,
    required this.totalPoints,
    required this.rank,
  });

  factory LeagueMember.fromJson(Map<String, dynamic> json) {
    return LeagueMember(
      userId: json['user_id'] as String,
      displayName: json['display_name'] as String,
      avatarUrl: json['avatar_url'] as String?,
      totalPoints: json['total_points'] as int,
      rank: json['rank'] as int,
    );
  }
}

class League {
  final String id;
  final String name;
  final String inviteCode;
  final String creatorId;
  final String creatorName;
  final String? eventId;
  final int memberCount;
  final List<LeagueMember> members;
  final DateTime createdAt;

  const League({
    required this.id,
    required this.name,
    required this.inviteCode,
    required this.creatorId,
    required this.creatorName,
    this.eventId,
    required this.memberCount,
    this.members = const [],
    required this.createdAt,
  });

  factory League.fromJson(Map<String, dynamic> json) {
    return League(
      id: json['id'] as String,
      name: json['name'] as String,
      inviteCode: json['invite_code'] as String,
      creatorId: json['creator_id'] as String,
      creatorName: json['creator_name'] as String,
      eventId: json['event_id'] as String?,
      memberCount: json['member_count'] as int,
      members: json['members'] != null
          ? (json['members'] as List<dynamic>)
              .map((e) => LeagueMember.fromJson(e as Map<String, dynamic>))
              .toList()
          : const [],
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }
}
