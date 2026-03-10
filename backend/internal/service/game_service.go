package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

type GameService struct {
	sessionRepo        *postgres.SessionRepo
	answerRepo         *postgres.AnswerRepo
	basketRepo         *postgres.BasketRepo
	cardRepo           *postgres.CardRepo
	cacheRepo          *redis.CacheRepo
	shuffleService     *ShuffleService
	streakService      *StreakService
	achievementService *AchievementService
}

func NewGameService(
	sessionRepo *postgres.SessionRepo,
	answerRepo *postgres.AnswerRepo,
	basketRepo *postgres.BasketRepo,
	cardRepo *postgres.CardRepo,
	cacheRepo *redis.CacheRepo,
	shuffleService *ShuffleService,
	streakService *StreakService,
	achievementService *AchievementService,
) *GameService {
	return &GameService{
		sessionRepo:        sessionRepo,
		answerRepo:         answerRepo,
		basketRepo:         basketRepo,
		cardRepo:           cardRepo,
		cacheRepo:          cacheRepo,
		shuffleService:     shuffleService,
		streakService:      streakService,
		achievementService: achievementService,
	}
}

// StartSession finds today's published basket, checks for an existing session (resume),
// or creates a new one with shuffled card order and initial resource counts.
func (s *GameService) StartSession(ctx context.Context, userID uuid.UUID) (*domain.UserSession, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	// Find today's published baskets
	baskets, err := s.basketRepo.FindPublishedByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("find today's basket: %w", err)
	}
	if len(baskets) == 0 {
		return nil, fmt.Errorf("no published basket available for today")
	}

	// Use the first published basket for today
	basket := baskets[0]

	// Check for existing active session (resume)
	existing, err := s.sessionRepo.FindByUserAndBasket(ctx, userID, basket.ID)
	if err != nil {
		return nil, fmt.Errorf("find existing session: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Create new session with shuffled card order
	shuffleOrder := s.shuffleService.Shuffle(userID, today, domain.TotalCards)

	// Apply streak bonuses to the new session
	var bonusSkips, bonusAnswers int
	if s.streakService != nil {
		bonusSkips, bonusAnswers, _ = s.streakService.ApplyBonuses(ctx, userID)
	}

	session := &domain.UserSession{
		ID:           uuid.New(),
		UserID:       userID,
		BasketID:     basket.ID,
		ShuffleOrder: shuffleOrder,
		CurrentIndex: 0,
		AnswersUsed:  0,
		SkipsUsed:    0,
		BonusAnswers: bonusAnswers,
		BonusSkips:   bonusSkips,
		Status:       domain.SessionStatusActive,
		StartedAt:    time.Now().UTC(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

// GetCurrentCard returns the card at the current index in the session's shuffle order.
// It also sets CardPresentedAt to track when the card was shown (for server-side timer enforcement).
func (s *GameService) GetCurrentCard(ctx context.Context, session *domain.UserSession) (*domain.CardView, error) {
	if session.IsComplete() {
		return nil, fmt.Errorf("session is complete, no more cards")
	}

	// Get all cards for the basket ordered by position
	cards, err := s.basketRepo.GetCardsForBasket(ctx, session.BasketID)
	if err != nil {
		return nil, fmt.Errorf("get basket cards: %w", err)
	}

	// Get the card at the current position in shuffle order
	position := session.ShuffleOrder[session.CurrentIndex]
	if position < 1 || position > len(cards) {
		return nil, fmt.Errorf("invalid card position %d", position)
	}

	// Record when this card was presented (start the timer).
	// Only set if not already set for the current card (avoid resetting on re-fetch).
	if session.CardPresentedAt == nil {
		now := time.Now().UTC()
		session.CardPresentedAt = &now
		if err := s.sessionRepo.UpdateProgress(ctx, session); err != nil {
			return nil, fmt.Errorf("update card presented time: %w", err)
		}
	}

	card := cards[position-1] // positions are 1-based
	return card.ToView(), nil
}

// SubmitAnswer validates resources, creates a user answer, advances the index, and checks completion.
// If the card timer has expired (40s + 2s grace), the card is auto-skipped instead.
func (s *GameService) SubmitAnswer(ctx context.Context, session *domain.UserSession, userID uuid.UUID, answer bool) (*domain.AnswerResult, error) {
	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	if session.Status != domain.SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	if session.IsComplete() {
		return nil, fmt.Errorf("session is complete, no more cards")
	}

	// Check if the card timer has expired — auto-skip instead of accepting the answer
	if session.IsCardExpired() {
		view, err := s.autoSkipCard(ctx, session)
		if err != nil {
			return nil, err
		}
		return &domain.AnswerResult{
			AutoSkipped:     true,
			SessionProgress: view,
		}, nil
	}

	if session.AnswersRemaining() <= 0 {
		return nil, fmt.Errorf("no answers remaining")
	}

	// Get all cards for the basket ordered by position
	cards, err := s.basketRepo.GetCardsForBasket(ctx, session.BasketID)
	if err != nil {
		return nil, fmt.Errorf("get basket cards: %w", err)
	}

	// Get current card
	position := session.ShuffleOrder[session.CurrentIndex]
	if position < 1 || position > len(cards) {
		return nil, fmt.Errorf("invalid card position %d", position)
	}
	card := cards[position-1]

	// Create user answer
	userAnswer := &domain.UserAnswer{
		ID:         uuid.New(),
		SessionID:  session.ID,
		CardID:     card.ID,
		UserID:     userID,
		Answer:     answer,
		AnsweredAt: time.Now().UTC(),
	}

	if err := s.answerRepo.Create(ctx, userAnswer); err != nil {
		return nil, fmt.Errorf("create answer: %w", err)
	}

	// Advance session and reset card timer
	session.AnswersUsed++
	session.CurrentIndex++
	session.CardPresentedAt = nil // Reset timer for next card

	if session.IsComplete() {
		session.Status = domain.SessionStatusCompleted
		now := time.Now().UTC()
		session.CompletedAt = &now
	}

	if err := s.sessionRepo.UpdateProgress(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// Update streak when session is completed
	if session.Status == domain.SessionStatusCompleted && s.streakService != nil {
		_, _ = s.streakService.UpdateStreak(ctx, userID, time.Now().UTC())
	}

	// Check achievements on session completion
	if session.Status == domain.SessionStatusCompleted {
		s.checkSessionAchievements(ctx, session)
	}

	return &domain.AnswerResult{
		CardID:          card.ID,
		Answer:          answer,
		SessionProgress: session.ToView(),
	}, nil
}

// SkipCard validates skips remaining, advances the index, and checks completion.
// If the card timer has expired (40s + 2s grace), the card is auto-skipped without consuming a skip.
func (s *GameService) SkipCard(ctx context.Context, session *domain.UserSession, userID uuid.UUID) (*domain.SkipResult, error) {
	if session.UserID != userID {
		return nil, fmt.Errorf("session does not belong to user")
	}

	if session.Status != domain.SessionStatusActive {
		return nil, fmt.Errorf("session is not active")
	}

	if session.IsComplete() {
		return nil, fmt.Errorf("session is complete, no more cards")
	}

	// Check if the card timer has expired — auto-skip without consuming a skip resource
	if session.IsCardExpired() {
		view, err := s.autoSkipCard(ctx, session)
		if err != nil {
			return nil, err
		}
		return &domain.SkipResult{
			AutoSkipped:     true,
			SessionProgress: view,
		}, nil
	}

	if session.SkipsRemaining() <= 0 {
		return nil, fmt.Errorf("no skips remaining")
	}

	// Advance session and reset card timer
	session.SkipsUsed++
	session.CurrentIndex++
	session.CardPresentedAt = nil // Reset timer for next card

	if session.IsComplete() {
		session.Status = domain.SessionStatusCompleted
		now := time.Now().UTC()
		session.CompletedAt = &now
	}

	if err := s.sessionRepo.UpdateProgress(ctx, session); err != nil {
		return nil, fmt.Errorf("update session: %w", err)
	}

	// Update streak when session is completed
	if session.Status == domain.SessionStatusCompleted && s.streakService != nil {
		_, _ = s.streakService.UpdateStreak(ctx, userID, time.Now().UTC())
	}

	// Check achievements on session completion
	if session.Status == domain.SessionStatusCompleted {
		s.checkSessionAchievements(ctx, session)
	}

	return &domain.SkipResult{
		SessionProgress: session.ToView(),
	}, nil
}

// autoSkipCard advances the session to the next card due to timer expiry.
// It does NOT consume an answer or skip resource.
func (s *GameService) autoSkipCard(ctx context.Context, session *domain.UserSession) (*domain.SessionView, error) {
	session.CurrentIndex++
	session.CardPresentedAt = nil // Reset timer for next card

	if session.IsComplete() {
		session.Status = domain.SessionStatusCompleted
		now := time.Now().UTC()
		session.CompletedAt = &now
	}

	if err := s.sessionRepo.UpdateProgress(ctx, session); err != nil {
		return nil, fmt.Errorf("auto-skip update session: %w", err)
	}

	// Check achievements on session completion
	if session.Status == domain.SessionStatusCompleted {
		s.checkSessionAchievements(ctx, session)
	}

	return session.ToView(), nil
}

// checkSessionAchievements checks for achievements after a session is completed.
// It checks "first_prediction" (first ever completed session) and "perfect_day" (all answers correct).
func (s *GameService) checkSessionAchievements(ctx context.Context, session *domain.UserSession) {
	if s.achievementService == nil {
		return
	}

	userID := session.UserID

	// Check "first_prediction": the user's first ever completed session.
	// We count completed sessions for this user; if this is the only one (count == 1 after completion),
	// it's their first prediction session. We use value=1 so the achievement condition_value=1 matches.
	completedSessions, err := s.sessionRepo.FindByUserID(ctx, userID, 2)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to count sessions for first_prediction check")
	} else {
		completedCount := 0
		for _, sess := range completedSessions {
			if sess.Status == domain.SessionStatusCompleted {
				completedCount++
			}
		}
		if completedCount >= 1 {
			if err := s.achievementService.CheckAndGrant(ctx, userID, "first_prediction", completedCount); err != nil {
				log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to check first_prediction achievement")
			}
		}
	}

	// Check "perfect_day": all answers in this session are correct (resolved and is_correct=true).
	// Since answers may not be resolved yet at session completion time, we check if the user
	// answered all cards (no skips) which is the prerequisite for a perfect day.
	// The actual correctness check uses resolved answers from this session.
	answers, err := s.answerRepo.FindBySession(ctx, session.ID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to fetch answers for perfect_day check")
		return
	}

	if len(answers) > 0 {
		allCorrect := true
		allResolved := true
		for _, a := range answers {
			if a.IsCorrect == nil {
				allResolved = false
				break
			}
			if !*a.IsCorrect {
				allCorrect = false
				break
			}
		}
		// Only grant if all answers are resolved and correct
		if allResolved && allCorrect {
			if err := s.achievementService.CheckAndGrant(ctx, userID, "perfect_day", 1); err != nil {
				log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to check perfect_day achievement")
			}
		}
	}
}
