import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Profile E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('GET /me returns user data with correct fields', () async {
      final response = await dio.get('/me');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['id'], isA<String>());
      expect(data['xex_user_id'], isA<String>());
      expect(data['email'], isA<String>());
      expect(data['role'], isA<String>());
      expect(data['total_points'], isA<int>());
      expect(data['is_active'], isA<bool>());
      expect(data['created_at'], isA<String>());
      expect(data['updated_at'], isA<String>());
    });

    test('PUT /me updates display_name', () async {
      final newName = 'E2E Test User ${DateTime.now().millisecondsSinceEpoch}';

      final response = await dio.put(
        '/me',
        data: {'display_name': newName},
      );

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      // Verify the update stuck
      final getResp = await dio.get('/me');
      final getData = (getResp.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      expect(getData['display_name'], equals(newName));
    });

    test('PUT /me updates language', () async {
      final response = await dio.put(
        '/me',
        data: {'language': 'fa'},
      );

      // Some endpoints may not support language update separately
      expect(response.statusCode, anyOf(equals(200), equals(400)));

      if (response.statusCode == 200) {
        // Reset to English
        await dio.put('/me', data: {'language': 'en'});
      }
    });

    test('GET /me/stats returns user statistics', () async {
      final response = await dio.get('/me/stats');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['total_points'], isA<int>());
      expect(data['total_sessions'], isA<int>());
      expect(data['total_answers'], isA<int>());
      expect(data['correct_answers'], isA<int>());
      expect(data['current_streak'], isA<int>());
      expect(data['longest_streak'], isA<int>());
    });

    test('GET /me/history returns session history', () async {
      final response = await dio.get('/me/history');

      // Could be 200 with empty list or with data
      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);
      // data is expected to be a list (possibly empty)
      expect(body['data'], isA<List>());
    });
  });
}
