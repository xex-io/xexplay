import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Notifications E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('POST /devices/register registers a device token', () async {
      final response = await dio.post(
        '/devices/register',
        data: {
          'token': 'e2e-test-fcm-token-${DateTime.now().millisecondsSinceEpoch}',
          'device_type': 'android',
        },
      );

      // 200 or 201 on success, 409 if token already registered
      expect(
        response.statusCode,
        anyOf(equals(200), equals(201), equals(409)),
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);
      }
    });

    test('DELETE /devices/:token deregisters a device token', () async {
      // Register a token first
      final testToken = 'e2e-delete-test-${DateTime.now().millisecondsSinceEpoch}';

      await dio.post(
        '/devices/register',
        data: {
          'token': testToken,
          'device_type': 'android',
        },
      );

      // Now deregister it
      final response = await dio.delete('/devices/$testToken');

      // 200 on success, 404 if not found (acceptable)
      expect(
        response.statusCode,
        anyOf(equals(200), equals(204), equals(404)),
      );

      if (response.statusCode == 200) {
        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);
      }
    });

    test('POST /devices/register requires authentication', () async {
      final noAuth = unauthenticatedDio();
      final resp = await noAuth.post(
        '/devices/register',
        data: {'token': 'test', 'device_type': 'android'},
      );
      expect(resp.statusCode, equals(401));
      noAuth.close();
    });
  });
}
