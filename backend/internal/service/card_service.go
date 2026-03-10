package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type CardService struct {
	cardRepo           *postgres.CardRepo
	answerRepo         *postgres.AnswerRepo
	leaderboardService *LeaderboardService
	hub                *ws.Hub
}

func NewCardService(cardRepo *postgres.CardRepo, answerRepo *postgres.AnswerRepo, leaderboardService *LeaderboardService, hub *ws.Hub) *CardService {
	return &CardService{
		cardRepo:           cardRepo,
		answerRepo:         answerRepo,
		leaderboardService: leaderboardService,
		hub:                hub,
	}
}

// ResolveCard sets the correct answer on a card and bulk-resolves all user answers,
// calculating points earned per tier. It also updates leaderboards for each affected user.
func (s *CardService) ResolveCard(ctx context.Context, cardID uuid.UUID, correctAnswer bool) error {
	// Get the card
	card, err := s.cardRepo.FindByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("find card: %w", err)
	}
	if card == nil {
		return fmt.Errorf("card not found")
	}

	if card.IsResolved {
		return fmt.Errorf("card is already resolved")
	}

	// Fetch unresolved answers before bulk resolve so we can update leaderboards
	answers, err := s.answerRepo.FindByCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("fetch answers for leaderboard: %w", err)
	}

	// Mark the card as resolved with the correct answer
	if err := s.cardRepo.Resolve(ctx, cardID, correctAnswer); err != nil {
		return fmt.Errorf("resolve card: %w", err)
	}

	// Bulk-resolve all user answers for this card (handles scoring and user points internally)
	if err := s.answerRepo.BulkResolve(ctx, cardID, correctAnswer); err != nil {
		return fmt.Errorf("bulk resolve answers: %w", err)
	}

	// Update leaderboards and send WebSocket notifications for each affected user
	for _, a := range answers {
		if a.IsCorrect != nil {
			// Already resolved, skip
			continue
		}
		isCorrect := a.Answer == correctAnswer
		points := 0
		if isCorrect {
			points = card.PointsForAnswer(a.Answer)
		}

		if s.leaderboardService != nil {
			if err := s.leaderboardService.UpdateLeaderboard(ctx, a.UserID, points, isCorrect, nil); err != nil {
				log.Warn().Err(err).
					Str("user_id", a.UserID.String()).
					Str("card_id", cardID.String()).
					Msg("failed to update leaderboard after card resolution")
			}
		}

		// Broadcast card_resolved event to the affected user
		if s.hub != nil {
			s.hub.SendToUser(a.UserID, ws.Message{
				Type: "card_resolved",
				Data: map[string]interface{}{
					"card_id":        cardID.String(),
					"correct_answer": correctAnswer,
					"your_answer":    a.Answer,
					"is_correct":     isCorrect,
					"points_earned":  points,
				},
			})
		}
	}

	return nil
}
