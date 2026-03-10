import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/network/interceptors/auth_interceptor.dart';
import '../../../core/storage/secure_storage.dart';
import '../data/card_models.dart';
import '../data/game_remote_source.dart';
import '../data/game_repository.dart';
import '../domain/game_state.dart';

// ---------------------------------------------------------------------------
// Dependency providers
// ---------------------------------------------------------------------------

final secureStorageProvider = Provider<SecureStorage>((ref) {
  return SecureStorage();
});

final apiClientProvider = Provider<ApiClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return ApiClient(authInterceptor: AuthInterceptor(storage));
});

final gameRemoteSourceProvider = Provider<GameRemoteSource>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return GameRemoteSource(apiClient);
});

final gameRepositoryProvider = Provider<GameRepository>((ref) {
  final remoteSource = ref.watch(gameRemoteSourceProvider);
  return GameRepository(remoteSource);
});

// ---------------------------------------------------------------------------
// Game state provider
// ---------------------------------------------------------------------------

final gameProvider = StateNotifierProvider<GameNotifier, GameState>((ref) {
  final repository = ref.watch(gameRepositoryProvider);
  return GameNotifier(repository);
});

class GameNotifier extends StateNotifier<GameState> {
  final GameRepository _repository;

  GameNotifier(this._repository) : super(const NoSession());

  /// Check whether the user can play today.
  Future<void> checkDailyStatus() async {
    state = const Loading();
    try {
      final status = await _repository.getDailyStatus();

      // If there is an active session, resume it.
      if (status.activeSessionId != null) {
        await _resumeSession();
        return;
      }

      state = NoSession(dailyStatus: status);
    } on GameException catch (e) {
      state = GameError(message: e.message);
    }
  }

  /// Start a new game session.
  Future<void> startSession() async {
    state = const Loading();
    try {
      final session = await _repository.startSession();
      final currentCard = await _repository.getCurrentCard();

      state = ActiveSession(
        session: session,
        currentCard: currentCard,
        timeRemaining: _calculateTimeRemaining(currentCard),
      );
    } on GameException catch (e) {
      state = GameError(message: e.message);
    }
  }

  /// Submit an answer (yes = true, no = false) for the current card.
  Future<void> submitAnswer(bool answer) async {
    final current = state;
    if (current is! ActiveSession) return;

    try {
      final result = await _repository.submitAnswer(
        cardId: current.currentCard.cardId,
        answer: answer,
      );

      await _advanceSession(current.session, result);
    } on GameException catch (e) {
      state = GameError(message: e.message);
    }
  }

  /// Skip the current card.
  Future<void> skipCard() async {
    final current = state;
    if (current is! ActiveSession) return;

    try {
      final result = await _repository.skipCard(
        cardId: current.currentCard.cardId,
      );

      await _advanceSession(current.session, result);
    } on GameException catch (e) {
      state = GameError(message: e.message);
    }
  }

  // ---------------------------------------------------------------------------
  // Private helpers
  // ---------------------------------------------------------------------------

  /// Resume an existing session (e.g. after app restart).
  Future<void> _resumeSession() async {
    try {
      final session = await _repository.getSession();

      if (session.isComplete) {
        state = SessionComplete(session: session);
        return;
      }

      final currentCard = await _repository.getCurrentCard();
      state = ActiveSession(
        session: session,
        currentCard: currentCard,
        timeRemaining: _calculateTimeRemaining(currentCard),
      );
    } on GameException catch (e) {
      state = GameError(message: e.message);
    }
  }

  /// After an answer/skip, either show the next card or complete the session.
  Future<void> _advanceSession(
    SessionModel previousSession,
    AnswerResultModel result,
  ) async {
    // Check if answers and skips are exhausted or we've gone through all cards.
    final allCardsProcessed =
        result.nextCardIndex >= previousSession.totalCards;
    final noActionsLeft =
        result.answersRemaining <= 0 && result.skipsRemaining <= 0;

    if (allCardsProcessed || noActionsLeft) {
      // Fetch the completed session for final scores.
      final session = await _repository.getSession();
      state = SessionComplete(session: session);
      return;
    }

    // Fetch next card.
    final session = await _repository.getSession();
    final nextCard = await _repository.getCurrentCard();

    state = ActiveSession(
      session: session,
      currentCard: nextCard,
      timeRemaining: _calculateTimeRemaining(nextCard),
    );
  }

  /// Placeholder for per-card timer calculation.
  /// Can be refined once backend provides card-level deadlines.
  int _calculateTimeRemaining(SessionCardModel card) {
    // Default: 30 seconds per card.
    return 30;
  }
}
