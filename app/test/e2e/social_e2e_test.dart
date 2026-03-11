import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Social E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    group('Referrals', () {
      test('GET /referral/code returns a referral code', () async {
        final response = await dio.get('/referral/code');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);

        final data = body['data'] as Map<String, dynamic>;
        expect(data['referral_code'], isA<String>());
        expect((data['referral_code'] as String).isNotEmpty, isTrue);
      });

      test('GET /referral/stats returns referral statistics', () async {
        final response = await dio.get('/referral/stats');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);

        final data = body['data'] as Map<String, dynamic>;
        expect(data['total_referrals'], isA<int>());
        expect(data['completed_referrals'], isA<int>());
      });
    });

    group('Achievements', () {
      test('GET /me/achievements returns achievements list', () async {
        final response = await dio.get('/me/achievements');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);

        final data = body['data'] as Map<String, dynamic>;
        expect(data.containsKey('achievements'), isTrue);
        expect(data.containsKey('earned'), isTrue);
      });
    });

    group('Leagues', () {
      test('GET /leagues returns leagues list', () async {
        final response = await dio.get('/leagues');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);
        expect(body['data'], isA<List>());
      });

      test('POST /leagues creates a new league', () async {
        final leagueName = 'E2E Test League ${DateTime.now().millisecondsSinceEpoch}';

        final response = await dio.post(
          '/leagues',
          data: {'name': leagueName},
        );

        // 200/201 on success, 409 if duplicate
        expect(
          response.statusCode,
          anyOf(equals(200), equals(201)),
        );

        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);

        final data = body['data'] as Map<String, dynamic>;
        expect(data['id'], isA<String>());
        expect(data['name'], equals(leagueName));
        expect(data['invite_code'], isA<String>());
        expect(data['creator_id'], isA<String>());
        expect(data['member_count'], isA<int>());
        expect(data['created_at'], isA<String>());
      });

      test('GET /leagues includes the newly created league', () async {
        final response = await dio.get('/leagues');

        expect(response.statusCode, equals(200));

        final body = response.data as Map<String, dynamic>;
        final leagues = body['data'] as List;
        // We just created a league, so there should be at least one
        expect(leagues, isNotEmpty);
      });
    });

    group('Auth required', () {
      test('Social endpoints require authentication', () async {
        final noAuth = unauthenticatedDio();

        final refResp = await noAuth.get('/referral/code');
        expect(refResp.statusCode, equals(401));

        final achResp = await noAuth.get('/me/achievements');
        expect(achResp.statusCode, equals(401));

        noAuth.close();
      });
    });
  });
}
