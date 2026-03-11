import 'package:dart_jsonwebtoken/dart_jsonwebtoken.dart';
import 'package:dio/dio.dart';

/// Shared JWT secret for generating Exchange-compatible tokens.
const _jwtSecret = 'B281NlDKzgVTlOIy0avLS6qwMJDtiFjxXdYj7A2VaJY4AZySvzDZjHt1GVM3tP';

/// Base URL for the live API, overridable via --dart-define=API_BASE_URL=...
const String apiBaseUrl = String.fromEnvironment(
  'API_BASE_URL',
  defaultValue: 'http://localhost:8080/v1',
);

/// A fixed UUID used as the test user's Exchange user_id.
/// Using a deterministic UUID so the same Play user is reused across test runs.
const String testXexUserId = 'a1b2c3d4-e5f6-4a7b-8c9d-0e1f2a3b4c5d';
const String testEmail = 'e2e-test@xexplay.io';

/// Generate an HS256 JWT that mimics an Exchange-issued access token.
String generateExchangeJwt({
  String userId = testXexUserId,
  String email = testEmail,
  String role = 'user',
  Duration ttl = const Duration(hours: 1),
}) {
  final now = DateTime.now().toUtc();

  final jwt = JWT(
    {
      'user_id': userId,
      'email': email,
      'role': role,
      'token_type': 'access',
    },
    issuer: 'nyyu',
    subject: userId,
  );

  return jwt.sign(
    SecretKey(_jwtSecret),
    algorithm: JWTAlgorithm.HS256,
    expiresIn: ttl,
    notBefore: Duration.zero,
  );
}

/// Create a Dio instance pointed at the live API with an auth token.
Dio authenticatedDio(String token) {
  return Dio(
    BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': 'Bearer $token',
      },
      // Don't throw on non-2xx so we can assert status codes ourselves.
      validateStatus: (status) => true,
    ),
  );
}

/// Create an unauthenticated Dio instance pointed at the live API.
Dio unauthenticatedDio() {
  return Dio(
    BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
      validateStatus: (status) => true,
    ),
  );
}

/// Login with the test Exchange JWT and return the auth token for further requests.
/// The backend uses the Exchange JWT directly as the Bearer token,
/// but login must be called first to ensure the Play user exists.
Future<String> loginAndGetToken(Dio dio) async {
  final exchangeJwt = generateExchangeJwt();

  final response = await Dio(
    BaseOptions(
      baseUrl: apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
      validateStatus: (status) => true,
    ),
  ).post('/auth/login', data: {'token': exchangeJwt});

  if (response.statusCode != 200) {
    throw Exception(
      'Login failed: ${response.statusCode} ${response.data}',
    );
  }

  // The backend validates the Exchange JWT on every request via the shared
  // secret, so we use the same Exchange JWT as the Bearer token.
  return exchangeJwt;
}
