// Data models for the game feature with JSON serialization.

enum CardTier {
  gold,
  silver,
  white;

  factory CardTier.fromJson(String value) {
    return CardTier.values.firstWhere(
      (e) => e.name == value,
      orElse: () => CardTier.white,
    );
  }
}

class CardModel {
  final String id;
  final String matchId;
  final CardTier tier;
  final Map<String, String> questionText;
  final bool highAnswerIsYes;
  final bool? correctAnswer;
  final bool isResolved;
  final DateTime availableDate;
  final DateTime expiresAt;
  final int pointsForCorrect;
  final int pointsForIncorrect;

  const CardModel({
    required this.id,
    required this.matchId,
    required this.tier,
    required this.questionText,
    required this.highAnswerIsYes,
    this.correctAnswer,
    required this.isResolved,
    required this.availableDate,
    required this.expiresAt,
    required this.pointsForCorrect,
    required this.pointsForIncorrect,
  });

  factory CardModel.fromJson(Map<String, dynamic> json) {
    return CardModel(
      id: json['id'] as String,
      matchId: json['match_id'] as String,
      tier: CardTier.fromJson(json['tier'] as String),
      questionText: Map<String, String>.from(json['question_text'] as Map),
      highAnswerIsYes: json['high_answer_is_yes'] as bool,
      correctAnswer: json['correct_answer'] as bool?,
      isResolved: json['is_resolved'] as bool,
      availableDate: DateTime.parse(json['available_date'] as String),
      expiresAt: DateTime.parse(json['expires_at'] as String),
      pointsForCorrect: json['points_for_correct'] as int,
      pointsForIncorrect: json['points_for_incorrect'] as int,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'match_id': matchId,
      'tier': tier.name,
      'question_text': questionText,
      'high_answer_is_yes': highAnswerIsYes,
      'correct_answer': correctAnswer,
      'is_resolved': isResolved,
      'available_date': availableDate.toIso8601String(),
      'expires_at': expiresAt.toIso8601String(),
      'points_for_correct': pointsForCorrect,
      'points_for_incorrect': pointsForIncorrect,
    };
  }
}

class SessionCardModel {
  final String cardId;
  final int position;
  final CardTier tier;
  final Map<String, String> questionText;
  final int pointsForCorrect;
  final int pointsForIncorrect;
  final bool? answer;
  final bool isSkipped;
  final int pointsEarned;

  const SessionCardModel({
    required this.cardId,
    required this.position,
    required this.tier,
    required this.questionText,
    required this.pointsForCorrect,
    required this.pointsForIncorrect,
    this.answer,
    required this.isSkipped,
    required this.pointsEarned,
  });

  factory SessionCardModel.fromJson(Map<String, dynamic> json) {
    return SessionCardModel(
      cardId: json['card_id'] as String,
      position: json['position'] as int,
      tier: CardTier.fromJson(json['tier'] as String),
      questionText: Map<String, String>.from(json['question_text'] as Map),
      pointsForCorrect: json['points_for_correct'] as int,
      pointsForIncorrect: json['points_for_incorrect'] as int,
      answer: json['answer'] as bool?,
      isSkipped: json['is_skipped'] as bool? ?? false,
      pointsEarned: json['points_earned'] as int? ?? 0,
    );
  }
}

class SessionModel {
  final String id;
  final String userId;
  final String basketId;
  final DateTime basketDate;
  final int currentCardIndex;
  final int answersUsed;
  final int skipsUsed;
  final int maxAnswers;
  final int maxSkips;
  final int totalCards;
  final int score;
  final bool isComplete;
  final DateTime startedAt;
  final DateTime? completedAt;
  final List<SessionCardModel> cards;

  const SessionModel({
    required this.id,
    required this.userId,
    required this.basketId,
    required this.basketDate,
    required this.currentCardIndex,
    required this.answersUsed,
    required this.skipsUsed,
    this.maxAnswers = 10,
    this.maxSkips = 5,
    this.totalCards = 15,
    required this.score,
    required this.isComplete,
    required this.startedAt,
    this.completedAt,
    required this.cards,
  });

  factory SessionModel.fromJson(Map<String, dynamic> json) {
    return SessionModel(
      id: json['id'] as String,
      userId: json['user_id'] as String,
      basketId: json['basket_id'] as String,
      basketDate: DateTime.parse(json['basket_date'] as String),
      currentCardIndex: json['current_card_index'] as int,
      answersUsed: json['answers_used'] as int,
      skipsUsed: json['skips_used'] as int,
      maxAnswers: json['max_answers'] as int? ?? 10,
      maxSkips: json['max_skips'] as int? ?? 5,
      totalCards: json['total_cards'] as int? ?? 15,
      score: json['score'] as int,
      isComplete: json['is_complete'] as bool,
      startedAt: DateTime.parse(json['started_at'] as String),
      completedAt: json['completed_at'] != null
          ? DateTime.parse(json['completed_at'] as String)
          : null,
      cards: (json['cards'] as List<dynamic>?)
              ?.map((c) =>
                  SessionCardModel.fromJson(c as Map<String, dynamic>))
              .toList() ??
          [],
    );
  }
}

class AnswerResultModel {
  final bool? correct;
  final int pointsEarned;
  final int answersRemaining;
  final int skipsRemaining;
  final int nextCardIndex;

  const AnswerResultModel({
    this.correct,
    required this.pointsEarned,
    required this.answersRemaining,
    required this.skipsRemaining,
    required this.nextCardIndex,
  });

  factory AnswerResultModel.fromJson(Map<String, dynamic> json) {
    return AnswerResultModel(
      correct: json['correct'] as bool?,
      pointsEarned: json['points_earned'] as int,
      answersRemaining: json['answers_remaining'] as int,
      skipsRemaining: json['skips_remaining'] as int,
      nextCardIndex: json['next_card_index'] as int,
    );
  }
}

class DailyStatusModel {
  final bool hasPlayedToday;
  final bool sessionAvailable;
  final String? activeSessionId;
  final int? score;
  final DateTime? nextSessionAt;

  const DailyStatusModel({
    required this.hasPlayedToday,
    required this.sessionAvailable,
    this.activeSessionId,
    this.score,
    this.nextSessionAt,
  });

  factory DailyStatusModel.fromJson(Map<String, dynamic> json) {
    return DailyStatusModel(
      hasPlayedToday: json['has_played_today'] as bool,
      sessionAvailable: json['session_available'] as bool,
      activeSessionId: json['active_session_id'] as String?,
      score: json['score'] as int?,
      nextSessionAt: json['next_session_at'] != null
          ? DateTime.parse(json['next_session_at'] as String)
          : null,
    );
  }
}
