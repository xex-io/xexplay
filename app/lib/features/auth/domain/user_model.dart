class User {
  final String id;
  final String xexUserId;
  final String email;
  final String displayName;
  final String? avatarUrl;
  final String role;
  final String referralCode;
  final String language;
  final int totalPoints;
  final int currentStreak;
  final bool isActive;

  const User({
    required this.id,
    required this.xexUserId,
    required this.email,
    required this.displayName,
    this.avatarUrl,
    required this.role,
    this.referralCode = '',
    this.language = 'en',
    required this.totalPoints,
    this.currentStreak = 0,
    this.isActive = true,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      xexUserId: json['xex_user_id'] as String,
      email: json['email'] as String,
      displayName: json['display_name'] as String? ?? '',
      avatarUrl: json['avatar_url'] as String?,
      role: json['role'] as String? ?? 'user',
      referralCode: json['referral_code'] as String? ?? '',
      language: json['language'] as String? ?? 'en',
      totalPoints: json['total_points'] as int? ?? 0,
      currentStreak: json['current_streak'] as int? ?? 0,
      isActive: json['is_active'] as bool? ?? true,
    );
  }
}
