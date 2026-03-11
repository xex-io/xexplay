import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Game E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('GET /events returns list of events', () async {
      final response = await dio.get('/events');

      // Events might be empty if none are configured, but the endpoint should respond.
      expect(response.statusCode, anyOf(equals(200), equals(404)));

      if (response.statusCode == 200) {
        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);
        // data could be a list or wrapped object
        expect(body['data'], isNotNull);
      }
    });

    test('GET /sessions/current returns session or appropriate error', () async {
      final response = await dio.get('/sessions/current');

      // 200 if active session exists, 404 if no session
      expect(
        response.statusCode,
        anyOf(equals(200), equals(404)),
      );

      final body = response.data as Map<String, dynamic>;
      if (response.statusCode == 200) {
        expect(body['success'], isTrue);
        final data = body['data'] as Map<String, dynamic>;
        expect(data['id'], isA<String>());
        expect(data['user_id'], isA<String>());
        expect(data['score'], isA<int>());
        expect(data['is_complete'], isA<bool>());
      } else {
        expect(body['success'], isFalse);
      }
    });

    test('POST /sessions/start attempts to create a session', () async {
      final response = await dio.post('/sessions/start');

      // 200/201 if basket available, 400/404/409 if no basket or session already active
      expect(
        response.statusCode,
        anyOf(equals(200), equals(201), equals(400), equals(404), equals(409)),
      );

      final body = response.data as Map<String, dynamic>;
      if (response.statusCode == 200 || response.statusCode == 201) {
        expect(body['success'], isTrue);
        final data = body['data'] as Map<String, dynamic>;
        expect(data['id'], isA<String>());
        expect(data['basket_id'], isA<String>());
      }
    });

    test('GET /sessions/current/card returns card or error', () async {
      final response = await dio.get('/sessions/current/card');

      expect(
        response.statusCode,
        anyOf(equals(200), equals(404)),
      );

      if (response.statusCode == 200) {
        final body = response.data as Map<String, dynamic>;
        expect(body['success'], isTrue);
        final data = body['data'] as Map<String, dynamic>;
        expect(data['card_id'], isA<String>());
        expect(data['position'], isA<int>());
        expect(data['tier'], isA<String>());
        expect(data['question_text'], isA<Map>());
      }
    });

    test('POST /sessions/current/answer submits answer or returns error', () async {
      // First try to get the current card so we have a card_id
      final cardResponse = await dio.get('/sessions/current/card');

      if (cardResponse.statusCode == 200) {
        final cardData =
            (cardResponse.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
        final cardId = cardData['card_id'] as String;

        final response = await dio.post(
          '/sessions/current/answer',
          data: {'card_id': cardId, 'answer': true},
        );

        // 200 if answered, 409 if already answered, 404 if no session
        expect(
          response.statusCode,
          anyOf(equals(200), equals(409), equals(404), equals(400)),
        );

        if (response.statusCode == 200) {
          final body = response.data as Map<String, dynamic>;
          expect(body['success'], isTrue);
          final data = body['data'] as Map<String, dynamic>;
          expect(data['points_earned'], isA<int>());
          expect(data['answers_remaining'], isA<int>());
          expect(data['skips_remaining'], isA<int>());
        }
      } else {
        // No active session -- just verify the answer endpoint also rejects
        final response = await dio.post(
          '/sessions/current/answer',
          data: {'card_id': 'nonexistent', 'answer': true},
        );

        expect(
          response.statusCode,
          anyOf(equals(404), equals(400)),
        );
      }
    });

    test('POST /sessions/current/skip skips card or returns error', () async {
      final cardResponse = await dio.get('/sessions/current/card');

      if (cardResponse.statusCode == 200) {
        final cardData =
            (cardResponse.data as Map<String, dynamic>)['data'] as Map<String, dynamic>;
        final cardId = cardData['card_id'] as String;

        final response = await dio.post(
          '/sessions/current/skip',
          data: {'card_id': cardId},
        );

        expect(
          response.statusCode,
          anyOf(equals(200), equals(409), equals(404), equals(400)),
        );

        if (response.statusCode == 200) {
          final body = response.data as Map<String, dynamic>;
          expect(body['success'], isTrue);
          final data = body['data'] as Map<String, dynamic>;
          expect(data['skips_remaining'], isA<int>());
        }
      } else {
        // No active session
        final response = await dio.post(
          '/sessions/current/skip',
          data: {'card_id': 'nonexistent'},
        );

        expect(
          response.statusCode,
          anyOf(equals(404), equals(400)),
        );
      }
    });

    test('Game endpoints require authentication', () async {
      final noAuth = unauthenticatedDio();

      final eventsResp = await noAuth.get('/events');
      expect(eventsResp.statusCode, equals(401));

      final sessionResp = await noAuth.get('/sessions/current');
      expect(sessionResp.statusCode, equals(401));

      noAuth.close();
    });
  });
}
