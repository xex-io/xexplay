import '../../../core/network/api_client.dart';
import '../../../core/constants/api_constants.dart';
import 'card_models.dart';

class GameRemoteSource {
  final ApiClient _apiClient;

  GameRemoteSource(this._apiClient);

  /// POST /sessions/start — create a new game session.
  Future<SessionModel> startSession() async {
    final response = await _apiClient.dio.post(ApiConstants.sessionsStart);
    return SessionModel.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /sessions/current — get the current active session.
  Future<SessionModel> getSession() async {
    final response = await _apiClient.dio.get(ApiConstants.sessionsCurrent);
    return SessionModel.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /sessions/current/card — get the current card in the session.
  Future<SessionCardModel> getCurrentCard() async {
    final response =
        await _apiClient.dio.get(ApiConstants.sessionsCurrentCard);
    return SessionCardModel.fromJson(response.data as Map<String, dynamic>);
  }

  /// POST /sessions/current/answer — submit an answer for the current card.
  Future<AnswerResultModel> submitAnswer({
    required String cardId,
    required bool answer,
  }) async {
    final response = await _apiClient.dio.post(
      ApiConstants.sessionsCurrentAnswer,
      data: {
        'card_id': cardId,
        'answer': answer,
      },
    );
    return AnswerResultModel.fromJson(response.data as Map<String, dynamic>);
  }

  /// POST /sessions/current/skip — skip the current card.
  Future<AnswerResultModel> skipCard({required String cardId}) async {
    final response = await _apiClient.dio.post(
      ApiConstants.sessionsCurrentSkip,
      data: {'card_id': cardId},
    );
    return AnswerResultModel.fromJson(response.data as Map<String, dynamic>);
  }

  /// GET /game/daily-status — check today's game availability.
  Future<DailyStatusModel> getDailyStatus() async {
    final response = await _apiClient.dio.get('/game/daily-status');
    return DailyStatusModel.fromJson(response.data as Map<String, dynamic>);
  }
}
