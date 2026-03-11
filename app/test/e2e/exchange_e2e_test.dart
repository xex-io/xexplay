import 'package:dio/dio.dart';
import 'package:test/test.dart';

import 'test_helper.dart';

void main() {
  group('Exchange E2E', () {
    late String authToken;
    late Dio dio;

    setUpAll(() async {
      authToken = await loginAndGetToken(unauthenticatedDio());
      dio = authenticatedDio(authToken);
    });

    tearDownAll(() {
      dio.close();
    });

    test('GET /me/exchange-prompts returns exchange prompts array', () async {
      final response = await dio.get('/me/exchange-prompts');

      expect(response.statusCode, equals(200));

      final body = response.data as Map<String, dynamic>;
      expect(body['success'], isTrue);

      final data = body['data'] as Map<String, dynamic>;
      expect(data['prompts'], isA<List>());

      final prompts = data['prompts'] as List;
      if (prompts.isNotEmpty) {
        final prompt = prompts.first as Map<String, dynamic>;
        expect(prompt['type'], isA<String>());
        expect(prompt['title'], isA<String>());
        expect(prompt['message'], isA<String>());
        expect(prompt['cta_url'], isA<String>());
      }
    });

    test('GET /me/exchange-prompts requires authentication', () async {
      final noAuth = unauthenticatedDio();
      final resp = await noAuth.get('/me/exchange-prompts');
      expect(resp.statusCode, equals(401));
      noAuth.close();
    });
  });
}
