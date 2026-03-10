import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/interceptors/auth_interceptor.dart';
import '../../../core/storage/secure_storage.dart';
import '../data/auth_remote_source.dart';
import '../data/auth_repository.dart';
import '../domain/auth_state.dart';

// ---------------------------------------------------------------------------
// Infrastructure providers
// ---------------------------------------------------------------------------

final secureStorageProvider = Provider<SecureStorage>((_) => SecureStorage());

final authInterceptorProvider = Provider<AuthInterceptor>((ref) {
  return AuthInterceptor(ref.watch(secureStorageProvider));
});

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient(authInterceptor: ref.watch(authInterceptorProvider));
});

final authRemoteSourceProvider = Provider<AuthRemoteSource>((ref) {
  return AuthRemoteSource(apiClient: ref.watch(apiClientProvider));
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(
    remoteSource: ref.watch(authRemoteSourceProvider),
    secureStorage: ref.watch(secureStorageProvider),
  );
});

// ---------------------------------------------------------------------------
// Auth state notifier
// ---------------------------------------------------------------------------

class AuthNotifier extends StateNotifier<AuthState> {
  final AuthRepository _repository;

  AuthNotifier(this._repository) : super(const AuthUnauthenticated());

  /// Called once on app start to restore the session.
  Future<void> checkAuth() async {
    state = const AuthLoading();
    try {
      final user = await _repository.checkAuth();
      if (user != null) {
        state = AuthAuthenticated(user);
      } else {
        state = const AuthUnauthenticated();
      }
    } catch (e) {
      state = const AuthUnauthenticated();
    }
  }

  /// Logs in using the Exchange JWT token.
  Future<void> login(String exchangeToken) async {
    state = const AuthLoading();
    try {
      final user = await _repository.login(exchangeToken);
      state = AuthAuthenticated(user);
    } catch (e) {
      state = AuthError(e.toString());
    }
  }

  /// Logs out and clears stored tokens.
  Future<void> logout() async {
    await _repository.logout();
    state = const AuthUnauthenticated();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref.watch(authRepositoryProvider));
});
