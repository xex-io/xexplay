import 'package:dio/dio.dart';
import '../../../core/constants/api_constants.dart';
import '../../../core/network/api_client.dart';
import '../domain/user_model.dart';

class AuthLoginResponse {
  final User user;
  final String accessToken;
  final String refreshToken;

  const AuthLoginResponse({
    required this.user,
    required this.accessToken,
    required this.refreshToken,
  });
}

class AuthRemoteSource {
  final ApiClient _apiClient;

  AuthRemoteSource({required ApiClient apiClient}) : _apiClient = apiClient;

  /// Sends the Exchange JWT to the Play backend.
  /// The backend validates it, creates/finds the Play user,
  /// and returns Play-specific tokens + user data.
  Future<AuthLoginResponse> login(String exchangeToken) async {
    final response = await _apiClient.dio.post(
      ApiConstants.login,
      data: {'exchange_token': exchangeToken},
    );

    final data = response.data as Map<String, dynamic>;
    return AuthLoginResponse(
      user: User.fromJson(data['user'] as Map<String, dynamic>),
      accessToken: data['access_token'] as String,
      refreshToken: data['refresh_token'] as String,
    );
  }

  /// Fetches the current user profile using the stored Play token.
  Future<User> me() async {
    final response = await _apiClient.dio.get(ApiConstants.me);
    final data = response.data as Map<String, dynamic>;
    return User.fromJson(data['user'] as Map<String, dynamic>);
  }

  /// Logs out on the server side.
  Future<void> logout() async {
    try {
      await _apiClient.dio.post(ApiConstants.logout);
    } on DioException {
      // Best-effort; token will be cleared locally regardless.
    }
  }
}
