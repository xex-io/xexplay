import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Rewards E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('GET /me/rewards returns rewards response with pending, history, streak', () async {
      final response = await dio.get('/me/rewards');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      // pending and history can be null when empty
      expect(data.containsKey('pending'), isTrue);
      expect(data.containsKey('history'), isTrue);
      // streak should always be present
      expect(data.containsKey('streak'), isTrue);
      if (data['streak'] != null) {
        final streak = data['streak'] as Map<String, dynamic>;
        expect(streak['current_streak'], isA<int>());
        expect(streak['longest_streak'], isA<int>());
        expect(streak['bonus_skips'], isA<int>());
        expect(streak['bonus_answers'], isA<int>());
      }
    });

    test('Pending rewards have correct structure when present', () async {
      final response = await dio.get('/me/rewards');
      final data = (response.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
      final pending = data['pending'];

      if (pending != null && (pending as List).isNotEmpty) {
        final reward = pending.first as Map<String, dynamic>;
        expect(reward['id'], isA<String>());
        expect(reward['period_type'], isA<String>());
        expect(reward['period_key'], isA<String>());
        expect(reward['reward_type'], isA<String>());
        expect(reward['amount'], isA<num>());
        expect(reward['status'], equals('pending'));
        expect(reward['created_at'], isA<String>());
      }
    });

    test('GET /me/rewards requires authentication', () async {
      final noAuth = unauthenticatedDio();
      final resp = await noAuth.get('/me/rewards');
      expect(resp.statusCode, equals(401));
      noAuth.close();
    });
  });
}
