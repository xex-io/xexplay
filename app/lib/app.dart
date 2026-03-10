import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/l10n/generated/app_localizations.dart';
import 'core/theme/app_theme.dart';
import 'core/routing/app_router.dart';
import 'features/auth/domain/auth_state.dart';
import 'features/auth/providers/auth_provider.dart';

class XexPlayApp extends ConsumerWidget {
  const XexPlayApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final isLoggedIn = authState is AuthAuthenticated;

    final router = AppRouter.router(isLoggedIn: isLoggedIn);

    return MaterialApp.router(
      title: 'XEX Play',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light,
      darkTheme: AppTheme.dark,
      themeMode: ThemeMode.dark,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: const [
        Locale('en'),
        Locale('fa'),
        Locale('ar'),
        Locale('tr'),
        Locale('es'),
        Locale('fr'),
      ],
      routerConfig: router,
    );
  }
}
