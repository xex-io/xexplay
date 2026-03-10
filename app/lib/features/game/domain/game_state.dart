import '../data/card_models.dart';

/// Sealed class representing all possible game states.
sealed class GameState {
  const GameState();
}

/// No active session. Optionally holds daily status info.
class NoSession extends GameState {
  final DailyStatusModel? dailyStatus;
  const NoSession({this.dailyStatus});
}

/// Loading state for any async operation.
class Loading extends GameState {
  const Loading();
}

/// An active game session with the current card and timer.
class ActiveSession extends GameState {
  final SessionModel session;
  final SessionCardModel currentCard;
  final int timeRemaining;

  const ActiveSession({
    required this.session,
    required this.currentCard,
    required this.timeRemaining,
  });

  ActiveSession copyWith({
    SessionModel? session,
    SessionCardModel? currentCard,
    int? timeRemaining,
  }) {
    return ActiveSession(
      session: session ?? this.session,
      currentCard: currentCard ?? this.currentCard,
      timeRemaining: timeRemaining ?? this.timeRemaining,
    );
  }
}

/// Session has been completed — show summary.
class SessionComplete extends GameState {
  final SessionModel session;
  const SessionComplete({required this.session});
}

/// An error occurred.
class GameError extends GameState {
  final String message;
  const GameError({required this.message});
}
