class ReferralStats {
  final String code;
  final int totalReferrals;
  final int completedReferrals;
  final int rewards;

  const ReferralStats({
    required this.code,
    required this.totalReferrals,
    required this.completedReferrals,
    required this.rewards,
  });

  factory ReferralStats.fromJson(Map<String, dynamic> json) {
    return ReferralStats(
      code: json['code'] as String,
      totalReferrals: json['total_referrals'] as int,
      completedReferrals: json['completed_referrals'] as int,
      rewards: json['rewards'] as int,
    );
  }
}
