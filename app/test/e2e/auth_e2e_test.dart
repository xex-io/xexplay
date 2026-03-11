import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Auth E2E', () {
    late Dio publicDio;

    setUp(() {
      publicDio = unauthenticatedDio();
    });

    tearDown(() {
      publicDio.close();
    });

    test('POST /auth/login with valid Exchange JWT creates/returns user', () async {
      final exchangeJwt = generateExchangeJwt();

      final response = await publicDio.post(
        '/auth/login',
        data: {'token': exchangeJwt},
      );

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);
      expect(body['data'], isNotNull);

      final data = body['data'] as Map<String, dynamic>;
      // The user object should contain these fields
      expect(data['id'], isA<String>());
      expect(data['xex_user_id'], equals(testXexUserId));
      expect(data['email'], equals(testEmail));
      expect(data['role'], isA<String>());
    });

    test('POST /auth/login with invalid token returns 401', () async {
      final response = await publicDio.post(
        '/auth/login',
        data: {'token': 'not-a-valid-jwt'},
      );

      expect(response.statusCode, equals(401));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isFalse);
      expect(body['error'], isNotNull);
    });

    test('POST /auth/login with missing token returns 400', () async {
      final response = await publicDio.post(
        '/auth/login',
        data: <String, dynamic>{},
      );

      expect(response.statusCode, equals(400));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isFalse);
    });

    test('POST /auth/login with expired token returns 401', () async {
      final expiredJwt = generateExchangeJwt(
        ttl: const Duration(seconds: -10),
      );

      final response = await publicDio.post(
        '/auth/login',
        data: {'token': expiredJwt},
      );

      expect(response.statusCode, equals(401));
    });

    group('Authenticated /me', () {
      late String authToken;
      late Dio authedDio;

      setUpAll(() async {
        authToken = await loginAndGetToken(unauthenticatedDio());
        authedDio = authenticatedDio(authToken);
      });

      tearDownAll(() {
        authedDio.close();
      });

      test('GET /me with valid token returns user profile', () async {
        final response = await authedDio.get('/me');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);

        final data = body['data'] as Map<String, dynamic>;
        expect(data['id'], isA<String>());
        expect(data['email'], equals(testEmail));
      });

      test('GET /me without token returns 401', () async {
        final noAuthDio = unauthenticatedDio();
        final response = await noAuthDio.get('/me');
        noAuthDio.close();

        expect(response.statusCode, equals(401));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isFalse);
        expect(body['error'], isNotNull);
      });

      test('GET /me with invalid token returns 401', () async {
        final badDio = authenticatedDio('invalid-token-value');
        final response = await badDio.get('/me');
        badDio.close();

        expect(response.statusCode, equals(401));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isFalse);
      });
    });
  });
}
