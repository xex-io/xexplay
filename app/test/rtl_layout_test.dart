import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:xexplay/core/l10n/generated/app_localizations.dart';
import 'package:xexplay/core/l10n/generated/app_localizations_fa.dart';
import 'package:xexplay/core/theme/app_theme.dart';
import 'package:xexplay/shared/widgets/app_button.dart';
import 'package:xexplay/shared/widgets/error_widget.dart';
import 'package:xexplay/shared/widgets/loading_widget.dart';

/// Helper that wraps [child] in a MaterialApp configured for Persian RTL.
Widget buildRtlApp({required Widget child}) {
  return MaterialApp(
    locale: const Locale('fa'),
    supportedLocales: AppLocalizations.supportedLocales,
    localizationsDelegates: AppLocalizations.localizationsDelegates,
    theme: AppTheme.dark,
    home: Scaffold(body: child),
  );
}

/// Retrieves the ambient [TextDirection] from the nearest [Directionality].
TextDirection getTextDirection(WidgetTester tester) {
  final directionality = tester.widget<Directionality>(
    find.byType(Directionality).first,
  );
  return directionality.textDirection;
}

void main() {
  group('RTL layout — Persian (fa)', () {
    testWidgets('Directionality is RTL when locale is fa', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: Builder(
          builder: (context) {
            final direction = Directionality.of(context);
            return Text('direction: $direction');
          },
        ),
      ));
      await tester.pumpAndSettle();

      final direction = getTextDirection(tester);
      expect(direction, TextDirection.rtl);
    });

    testWidgets('AppLocalizations resolves to Persian strings', (tester) async {
      late AppLocalizations l10n;

      await tester.pumpWidget(buildRtlApp(
        child: Builder(
          builder: (context) {
            l10n = AppLocalizations.of(context)!;
            return Text(l10n.play);
          },
        ),
      ));
      await tester.pumpAndSettle();

      expect(l10n, isA<AppLocalizationsFa>());
      expect(l10n.play, 'بازی');
      expect(l10n.leaderboard, 'جدول رده‌بندی');
      expect(l10n.rewards, 'جوایز');
      expect(l10n.profile, 'پروفایل');
      expect(l10n.loginWithExchange, 'ورود با XEX Exchange');
    });

    testWidgets('PrimaryButton renders in RTL without overflow', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: Center(
          child: PrimaryButton(label: 'شروع بازی امروز', onPressed: () {}),
        ),
      ));
      await tester.pumpAndSettle();

      expect(find.byType(PrimaryButton), findsOneWidget);
      expect(find.text('شروع بازی امروز'), findsOneWidget);
      // No overflow — test would fail if a RenderFlex overflow occurred.
    });

    testWidgets('SecondaryButton renders in RTL without overflow', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: Center(
          child: SecondaryButton(label: 'تلاش مجدد', onPressed: () {}),
        ),
      ));
      await tester.pumpAndSettle();

      expect(find.byType(SecondaryButton), findsOneWidget);
      expect(find.text('تلاش مجدد'), findsOneWidget);
    });

    testWidgets('GhostButton renders in RTL without overflow', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: Center(
          child: GhostButton(label: 'خروج', onPressed: () {}),
        ),
      ));
      await tester.pumpAndSettle();

      expect(find.byType(GhostButton), findsOneWidget);
      expect(find.text('خروج'), findsOneWidget);
    });

    testWidgets('LoadingWidget with Persian message renders in RTL', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: const LoadingWidget(message: 'در حال بارگذاری...'),
      ));
      // Don't pumpAndSettle — CircularProgressIndicator animates forever.
      await tester.pump();

      expect(find.byType(LoadingWidget), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('در حال بارگذاری...'), findsOneWidget);
    });

    testWidgets('AppErrorWidget with Persian message renders in RTL', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: AppErrorWidget(
          message: 'مشکلی پیش آمد',
          onRetry: () {},
        ),
      ));
      await tester.pumpAndSettle();

      expect(find.byType(AppErrorWidget), findsOneWidget);
      expect(find.text('مشکلی پیش آمد'), findsOneWidget);
      expect(find.byIcon(Icons.error_outline), findsOneWidget);
      expect(find.byIcon(Icons.refresh), findsOneWidget);
    });

    testWidgets('Row of buttons lays out correctly in RTL', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: Row(
          children: [
            Expanded(
              child: PrimaryButton(label: 'بله', onPressed: () {}),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: SecondaryButton(label: 'خیر', onPressed: () {}),
            ),
          ],
        ),
      ));
      await tester.pumpAndSettle();

      // In RTL, the first child ('بله') should be on the RIGHT side.
      final yesCenter = tester.getCenter(find.text('بله'));
      final noCenter = tester.getCenter(find.text('خیر'));
      expect(
        yesCenter.dx,
        greaterThan(noCenter.dx),
        reason: 'In RTL layout, the first Row child should appear to the right',
      );
    });

    testWidgets('Parameterized localization strings work in Persian', (tester) async {
      late AppLocalizations l10n;

      await tester.pumpWidget(buildRtlApp(
        child: Builder(
          builder: (context) {
            l10n = AppLocalizations.of(context)!;
            return Text(l10n.cardOf(3, 10));
          },
        ),
      ));
      await tester.pumpAndSettle();

      expect(l10n.cardOf(3, 10), 'کارت 3 از 10');
      expect(l10n.streakDays(5), '5 روز متوالی');
      expect(l10n.answersRemaining(7), '7 پاسخ باقی‌مانده');
      expect(l10n.skipsRemaining(2), '2 رد کردن باقی‌مانده');
    });

    testWidgets('Text alignment respects RTL', (tester) async {
      await tester.pumpWidget(buildRtlApp(
        child: const Padding(
          padding: EdgeInsets.all(16),
          child: Align(
            alignment: AlignmentDirectional.centerStart,
            child: Text('تست راست‌چین'),
          ),
        ),
      ));
      await tester.pumpAndSettle();

      // In RTL, AlignmentDirectional.centerStart means right-aligned.
      final textOffset = tester.getTopRight(find.text('تست راست‌چین'));
      final screenWidth = tester.getSize(find.byType(MaterialApp)).width;
      // The text's right edge should be near the right edge of the screen
      // (within padding).
      expect(
        textOffset.dx,
        greaterThan(screenWidth / 2),
        reason: 'AlignmentDirectional.centerStart should align right in RTL',
      );
    });
  });
}
