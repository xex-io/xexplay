import '../../../core/storage/secure_storage.dart';
import '../domain/user_model.dart';
import 'auth_remote_source.dart';

class AuthRepository {
  final AuthRemoteSource _remoteSource;
  final SecureStorage _secureStorage;

  AuthRepository({
    required AuthRemoteSource remoteSource,
    required SecureStorage secureStorage,
  })  : _remoteSource = remoteSource,
        _secureStorage = secureStorage;

  /// Logs in with an Exchange JWT, stores Play tokens, returns the user.
  Future<User> login(String exchangeToken) async {
    final result = await _remoteSource.login(exchangeToken);
    await _secureStorage.saveAccessToken(result.accessToken);
    await _secureStorage.saveRefreshToken(result.refreshToken);
    return result.user;
  }

  /// Clears local tokens and notifies the server.
  Future<void> logout() async {
    await _remoteSource.logout();
    await _secureStorage.clearAll();
  }

  /// Checks whether a valid token exists and fetches the user profile.
  /// Returns `null` if no token or the token is invalid/expired.
  Future<User?> checkAuth() async {
    final token = await _secureStorage.getAccessToken();
    if (token == null) return null;

    try {
      return await _remoteSource.me();
    } catch (_) {
      // Token is invalid or expired.
      await _secureStorage.clearAll();
      return null;
    }
  }

  /// Returns the stored access token, if any.
  Future<String?> getStoredToken() => _secureStorage.getAccessToken();
}
