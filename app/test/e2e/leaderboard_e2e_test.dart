import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Leaderboard E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('GET /leaderboards/daily returns daily leaderboard', () async {
      final response = await dio.get('/leaderboards/daily');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['period_type'], equals('daily'));
      expect(data['period_key'], isA<String>());
      // entries can be null when empty
      expect(data.containsKey('entries'), isTrue);
      expect(data['total'], isA<int>());
    });

    test('GET /leaderboards/weekly returns weekly leaderboard', () async {
      final response = await dio.get('/leaderboards/weekly');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['period_type'], equals('weekly'));
      expect(data['period_key'], isA<String>());
      expect(data.containsKey('entries'), isTrue);
      expect(data['total'], isA<int>());
    });

    test('GET /leaderboards/all-time returns all-time leaderboard', () async {
      final response = await dio.get('/leaderboards/all-time');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['period_type'], equals('all_time'));
      expect(data['period_key'], isA<String>());
      expect(data.containsKey('entries'), isTrue);
      expect(data['total'], isA<int>());
    });

    test('Leaderboard entries have correct structure when non-empty', () async {
      final response = await dio.get('/leaderboards/all-time');

      final body = response.data as Map<String, dynamic>;
      final data = body['data'] as Map<String, dynamic>;
      final entries = data['entries'];

      if (entries != null && (entries as List).isNotEmpty) {
        final entry = entries.first as Map<String, dynamic>;
        expect(entry['rank'], isA<int>());
        expect(entry['user_id'], isA<String>());
        expect(entry['display_name'], isA<String>());
        expect(entry['total_points'], isA<int>());
        expect(entry['correct_answers'], isA<int>());
      }
    });

    test('Leaderboard supports pagination via limit and offset', () async {
      final response = await dio.get(
        '/leaderboards/daily',
        queryParameters: {'limit': 5, 'offset': 0},
      );

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      final entries = data['entries'] as List? ?? [];
      expect(entries.length, lessThanOrEqualTo(5));
    });

    test('Leaderboards require authentication', () async {
      final noAuth = unauthenticatedDio();

      final resp = await noAuth.get('/leaderboards/daily');
      expect(resp.statusCode, equals(401));

      noAuth.close();
    });
  });
}
