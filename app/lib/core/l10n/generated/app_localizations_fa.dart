// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Persian (`fa`).
class AppLocalizationsFa extends AppLocalizations {
  AppLocalizationsFa([String locale = 'fa']) : super(locale);

  @override
  String get appName => 'XEX Play';

  @override
  String get loginWithExchange => 'ورود با XEX Exchange';

  @override
  String get noAccount => 'حساب ندارید؟ در XEX Exchange ایجاد کنید';

  @override
  String get play => 'بازی';

  @override
  String get leaderboard => 'جدول رده‌بندی';

  @override
  String get rewards => 'جوایز';

  @override
  String get profile => 'پروفایل';

  @override
  String get startSession => 'شروع بازی امروز';

  @override
  String get resumeSession => 'ادامه بازی';

  @override
  String get sessionComplete => 'بازی تمام شد';

  @override
  String cardOf(int current, int total) {
    return 'کارت $current از $total';
  }

  @override
  String answersRemaining(int count) {
    return '$count پاسخ باقی‌مانده';
  }

  @override
  String skipsRemaining(int count) {
    return '$count رد کردن باقی‌مانده';
  }

  @override
  String get noSkipsRemaining =>
      'رد کردن تمام شد — باید به همه کارت‌های باقی‌مانده پاسخ دهید';

  @override
  String get yes => 'بله';

  @override
  String get no => 'خیر';

  @override
  String get skip => 'رد کردن';

  @override
  String get swipeRightYes => 'برای بله به راست بکشید';

  @override
  String get swipeLeftNo => 'برای خیر به چپ بکشید';

  @override
  String get swipeUpSkip => 'برای رد کردن به بالا بکشید';

  @override
  String get pointsLabel => 'امتیاز';

  @override
  String get goldTier => 'طلایی';

  @override
  String get silverTier => 'نقره‌ای';

  @override
  String get whiteTier => 'سفید';

  @override
  String get daily => 'روزانه';

  @override
  String get weekly => 'هفتگی';

  @override
  String get tournament => 'تورنمنت';

  @override
  String get allTime => 'همه زمان‌ها';

  @override
  String get friends => 'دوستان';

  @override
  String get rank => 'رتبه';

  @override
  String get points => 'امتیاز';

  @override
  String get yourRank => 'رتبه شما';

  @override
  String get streak => 'روزهای متوالی';

  @override
  String streakDays(int count) {
    return '$count روز متوالی';
  }

  @override
  String get correctPredictions => 'پیش‌بینی‌های درست';

  @override
  String get totalAnswered => 'کل پاسخ‌ها';

  @override
  String get score => 'امتیاز';

  @override
  String get settings => 'تنظیمات';

  @override
  String get language => 'زبان';

  @override
  String get darkMode => 'حالت تاریک';

  @override
  String get notifications => 'اعلان‌ها';

  @override
  String get logout => 'خروج';

  @override
  String get error => 'مشکلی پیش آمد';

  @override
  String get retry => 'تلاش مجدد';

  @override
  String get loading => 'در حال بارگذاری...';
}
