class ApiConstants {
  ApiConstants._();

  static const String baseUrl = 'http://localhost:8080/v1';

  // Auth
  static const String login = '/auth/login';
  static const String logout = '/auth/logout';

  // User
  static const String me = '/me';
  static const String meStats = '/me/stats';
  static const String meHistory = '/me/history';
  static const String meAchievements = '/me/achievements';
  static const String meRewards = '/me/rewards';
  static const String meRewardsClaim = '/me/rewards/claim';

  /// Returns the claim endpoint for a specific reward.
  static String meRewardClaimById(String id) => '/me/rewards/$id/claim';

  // Game
  static const String sessionsStart = '/sessions/start';
  static const String sessionsCurrent = '/sessions/current';
  static const String sessionsCurrentCard = '/sessions/current/card';
  static const String sessionsCurrentAnswer = '/sessions/current/answer';
  static const String sessionsCurrentSkip = '/sessions/current/skip';

  // Events
  static const String events = '/events';

  // Leaderboards
  static const String leaderboardDaily = '/leaderboards/daily';
  static const String leaderboardWeekly = '/leaderboards/weekly';
  static const String leaderboardAllTime = '/leaderboards/all-time';
  static String leaderboardTournament(String eventId) =>
      '/leaderboards/tournament/$eventId';

  // Referral
  static const String referralCode = '/referral/code';
  static const String referralStats = '/referral/stats';

  // Leagues
  static const String leagues = '/leagues';
  static const String leaguesJoin = '/leagues/join';

  /// Returns the detail endpoint for a specific league.
  static String leagueDetail(String id) => '/leagues/$id';

  // Devices
  static const String devicesRegister = '/devices/register';
}
