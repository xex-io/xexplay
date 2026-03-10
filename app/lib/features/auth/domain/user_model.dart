class User {
  final int id;
  final int xexUserId;
  final String email;
  final String displayName;
  final String? avatarUrl;
  final String role;
  final int totalPoints;
  final int currentStreak;

  const User({
    required this.id,
    required this.xexUserId,
    required this.email,
    required this.displayName,
    this.avatarUrl,
    required this.role,
    required this.totalPoints,
    required this.currentStreak,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as int,
      xexUserId: json['xex_user_id'] as int,
      email: json['email'] as String,
      displayName: json['display_name'] as String? ?? '',
      avatarUrl: json['avatar_url'] as String?,
      role: json['role'] as String? ?? 'player',
      totalPoints: json['total_points'] as int? ?? 0,
      currentStreak: json['current_streak'] as int? ?? 0,
    );
  }
}
