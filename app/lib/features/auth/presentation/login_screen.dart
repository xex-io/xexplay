import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/constants/app_colors.dart';
import '../domain/auth_state.dart';
import '../providers/auth_provider.dart';

class LoginScreen extends ConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final isLoading = authState is AuthLoading;

    // Show snackbar on error.
    ref.listen<AuthState>(authProvider, (prev, next) {
      if (next is AuthError) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(next.message)),
        );
      }
    });

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Spacer(flex: 2),
              // Logo placeholder
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  color: AppColors.darkPrimaryBold,
                  borderRadius: BorderRadius.circular(20),
                ),
                child: const Center(
                  child: Text(
                    'XP',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 32,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'XEX Play',
                style: Theme.of(context).textTheme.displayLarge,
              ),
              const SizedBox(height: 8),
              Text(
                'Predict. Compete. Earn.',
                style: Theme.of(context).textTheme.bodyMedium,
              ),
              const Spacer(flex: 3),
              ElevatedButton(
                onPressed: isLoading
                    ? null
                    : () => _onLoginTapped(context, ref),
                child: isLoading
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Login with XEX Exchange'),
              ),
              const SizedBox(height: 16),
              Text(
                "Don't have an account? Create one on XEX Exchange",
                style: Theme.of(context).textTheme.bodySmall,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 48),
            ],
          ),
        ),
      ),
    );
  }

  void _onLoginTapped(BuildContext context, WidgetRef ref) {
    // TODO: Replace with actual Exchange login flow (e.g. deep-link or
    // web-view that returns an Exchange JWT). For now we use a placeholder
    // token so the full auth pipeline can be exercised end-to-end.
    const exchangeToken = 'EXCHANGE_JWT_PLACEHOLDER';
    ref.read(authProvider.notifier).login(exchangeToken);
  }
}
