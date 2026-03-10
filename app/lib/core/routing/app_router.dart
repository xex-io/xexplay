import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../features/auth/presentation/login_screen.dart';
import '../../features/auth/presentation/splash_screen.dart';
import '../../features/game/presentation/game_screen.dart';
import '../../features/game/presentation/home_screen.dart';
import '../../features/game/presentation/session_summary_screen.dart';
import '../../features/leaderboard/presentation/leaderboard_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/rewards/presentation/rewards_screen.dart';
import '../../features/social/presentation/achievements_screen.dart';
import '../../features/social/presentation/league_detail_screen.dart';
import '../../features/social/presentation/mini_leagues_screen.dart';
import '../../features/social/presentation/referral_screen.dart';
import '../../shared/widgets/main_shell.dart';

class AppRouter {
  static final _rootNavigatorKey = GlobalKey<NavigatorState>();
  static final _shellNavigatorKey = GlobalKey<NavigatorState>();

  static GoRouter router({required bool isLoggedIn}) => GoRouter(
        navigatorKey: _rootNavigatorKey,
        initialLocation: isLoggedIn ? '/play' : '/splash',
        redirect: (context, state) {
          final loc = state.matchedLocation;

          // Allow splash to always render.
          if (loc == '/splash') return null;

          if (!isLoggedIn && loc != '/login') {
            return '/login';
          }
          if (isLoggedIn && loc == '/login') {
            return '/play';
          }
          return null;
        },
        routes: [
          GoRoute(
            path: '/splash',
            builder: (context, state) => const SplashScreen(),
          ),
          GoRoute(
            path: '/login',
            builder: (context, state) => const LoginScreen(),
          ),
          ShellRoute(
            navigatorKey: _shellNavigatorKey,
            builder: (context, state, child) => MainShell(child: child),
            routes: [
              GoRoute(
                path: '/play',
                builder: (context, state) => const HomeScreen(),
              ),
              GoRoute(
                path: '/leaderboard',
                builder: (context, state) => const LeaderboardScreen(),
              ),
              GoRoute(
                path: '/rewards',
                builder: (context, state) => const RewardsScreen(),
              ),
              GoRoute(
                path: '/profile',
                builder: (context, state) => const ProfileScreen(),
              ),
            ],
          ),
          GoRoute(
            path: '/play/session',
            builder: (context, state) => const GameScreen(),
          ),
          GoRoute(
            path: '/play/summary',
            builder: (context, state) => const SessionSummaryScreen(),
          ),
          GoRoute(
            path: '/referral',
            builder: (context, state) => const ReferralScreen(),
          ),
          GoRoute(
            path: '/achievements',
            builder: (context, state) => const AchievementsScreen(),
          ),
          GoRoute(
            path: '/leagues',
            builder: (context, state) => const MiniLeaguesScreen(),
          ),
          GoRoute(
            path: '/leagues/:id',
            builder: (context, state) => LeagueDetailScreen(
              leagueId: state.pathParameters['id']!,
            ),
          ),
        ],
      );
}
