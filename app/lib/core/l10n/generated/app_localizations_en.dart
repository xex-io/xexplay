// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appName => 'XEX Play';

  @override
  String get loginWithExchange => 'Login with XEX Exchange';

  @override
  String get noAccount => 'Don\'t have an account? Create one on XEX Exchange';

  @override
  String get play => 'Play';

  @override
  String get leaderboard => 'Leaderboard';

  @override
  String get rewards => 'Rewards';

  @override
  String get profile => 'Profile';

  @override
  String get startSession => 'Start Today\'s Session';

  @override
  String get resumeSession => 'Resume Session';

  @override
  String get sessionComplete => 'Session Complete';

  @override
  String cardOf(int current, int total) {
    return 'Card $current of $total';
  }

  @override
  String answersRemaining(int count) {
    return '$count answers left';
  }

  @override
  String skipsRemaining(int count) {
    return '$count skips left';
  }

  @override
  String get noSkipsRemaining =>
      'No skips remaining — you must answer all remaining cards';

  @override
  String get yes => 'Yes';

  @override
  String get no => 'No';

  @override
  String get skip => 'Skip';

  @override
  String get swipeRightYes => 'Swipe right for Yes';

  @override
  String get swipeLeftNo => 'Swipe left for No';

  @override
  String get swipeUpSkip => 'Swipe up to Skip';

  @override
  String get pointsLabel => 'pts';

  @override
  String get goldTier => 'Gold';

  @override
  String get silverTier => 'Silver';

  @override
  String get whiteTier => 'White';

  @override
  String get daily => 'Daily';

  @override
  String get weekly => 'Weekly';

  @override
  String get tournament => 'Tournament';

  @override
  String get allTime => 'All Time';

  @override
  String get friends => 'Friends';

  @override
  String get rank => 'Rank';

  @override
  String get points => 'Points';

  @override
  String get yourRank => 'Your Rank';

  @override
  String get streak => 'Streak';

  @override
  String streakDays(int count) {
    return '$count day streak';
  }

  @override
  String get correctPredictions => 'Correct Predictions';

  @override
  String get totalAnswered => 'Total Answered';

  @override
  String get score => 'Score';

  @override
  String get settings => 'Settings';

  @override
  String get language => 'Language';

  @override
  String get darkMode => 'Dark Mode';

  @override
  String get notifications => 'Notifications';

  @override
  String get logout => 'Logout';

  @override
  String get error => 'Something went wrong';

  @override
  String get retry => 'Retry';

  @override
  String get loading => 'Loading...';
}
