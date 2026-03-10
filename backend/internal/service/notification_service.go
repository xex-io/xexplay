package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// NotificationSender is the interface for sending push notifications.
// Implement this with Firebase Cloud Messaging (FCM) for production use.
type NotificationSender interface {
	Send(ctx context.Context, token string, notification *domain.Notification) error
	SendBatch(ctx context.Context, tokens []string, notification *domain.Notification) error
}

// LogSender is a development implementation that logs notifications instead of sending them.
type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) Send(ctx context.Context, token string, notification *domain.Notification) error {
	log.Info().
		Str("token", token).
		Str("title", notification.Title).
		Str("body", notification.Body).
		Str("target_type", notification.TargetType).
		Msg("push notification sent (log sender)")
	return nil
}

func (s *LogSender) SendBatch(ctx context.Context, tokens []string, notification *domain.Notification) error {
	log.Info().
		Int("token_count", len(tokens)).
		Str("title", notification.Title).
		Str("body", notification.Body).
		Str("target_type", notification.TargetType).
		Msg("push notification batch sent (log sender)")
	return nil
}

// NotificationService manages sending push notifications to users.
type NotificationService struct {
	fcmRepo *postgres.FCMTokenRepo
	sender  NotificationSender
}

func NewNotificationService(fcmRepo *postgres.FCMTokenRepo, sender NotificationSender) *NotificationService {
	return &NotificationService{
		fcmRepo: fcmRepo,
		sender:  sender,
	}
}

// SendToUser sends a notification to all active devices of a specific user.
func (s *NotificationService) SendToUser(ctx context.Context, userID uuid.UUID, notification *domain.Notification) error {
	tokens, err := s.fcmRepo.FindActiveByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user tokens: %w", err)
	}

	if len(tokens) == 0 {
		log.Debug().Str("user_id", userID.String()).Msg("no active tokens for user, skipping notification")
		return nil
	}

	for _, t := range tokens {
		if err := s.sender.Send(ctx, t.Token, notification); err != nil {
			log.Warn().Err(err).
				Str("user_id", userID.String()).
				Str("token", t.Token).
				Msg("failed to send notification to device")
			// Deactivate invalid tokens
			_ = s.fcmRepo.DeactivateByTokenValue(ctx, t.Token)
		}
	}

	return nil
}

// SendToAll sends a notification to all active devices (broadcast).
func (s *NotificationService) SendToAll(ctx context.Context, notification *domain.Notification) error {
	tokens, err := s.fcmRepo.FindAllActive(ctx)
	if err != nil {
		return fmt.Errorf("get all active tokens: %w", err)
	}

	if len(tokens) == 0 {
		log.Debug().Msg("no active tokens, skipping broadcast notification")
		return nil
	}

	// Collect token strings for batch send
	tokenStrings := make([]string, 0, len(tokens))
	for _, t := range tokens {
		tokenStrings = append(tokenStrings, t.Token)
	}

	// Send in batches of 500
	batchSize := 500
	for i := 0; i < len(tokenStrings); i += batchSize {
		end := i + batchSize
		if end > len(tokenStrings) {
			end = len(tokenStrings)
		}
		batch := tokenStrings[i:end]
		if err := s.sender.SendBatch(ctx, batch, notification); err != nil {
			log.Warn().Err(err).
				Int("batch_start", i).
				Int("batch_size", len(batch)).
				Msg("failed to send notification batch")
		}
	}

	return nil
}

// NotifyCardResolved sends a notification when a user's card prediction is resolved.
func (s *NotificationService) NotifyCardResolved(ctx context.Context, userID uuid.UUID, cardQuestion string, isCorrect bool, points int) {
	result := "Wrong"
	if isCorrect {
		result = "Correct"
	}

	notification := &domain.Notification{
		Title:      fmt.Sprintf("Prediction %s!", result),
		Body:       fmt.Sprintf("Your prediction for \"%s\" was %s! You earned %d points.", cardQuestion, result, points),
		TargetType: domain.TargetUser,
		Data: map[string]string{
			"type":   "card_resolved",
			"result": result,
			"points": fmt.Sprintf("%d", points),
		},
	}

	if err := s.SendToUser(ctx, userID, notification); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to send card resolved notification")
	}
}

// NotifyStreakAtRisk sends a notification to a user whose streak is at risk.
func (s *NotificationService) NotifyStreakAtRisk(ctx context.Context, userID uuid.UUID, currentStreak int) {
	notification := &domain.Notification{
		Title:      "Your streak is at risk!",
		Body:       fmt.Sprintf("You have a %d-day streak. Play today to keep it going!", currentStreak),
		TargetType: domain.TargetUser,
		Data: map[string]string{
			"type":           "streak_at_risk",
			"current_streak": fmt.Sprintf("%d", currentStreak),
		},
	}

	if err := s.SendToUser(ctx, userID, notification); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to send streak at risk notification")
	}
}

// NotifyRewardEarned sends a notification when a user earns a reward.
func (s *NotificationService) NotifyRewardEarned(ctx context.Context, userID uuid.UUID, rewardType string, amount float64) {
	notification := &domain.Notification{
		Title:      "Reward earned!",
		Body:       fmt.Sprintf("You earned a %s reward of %.2f! Check your rewards to claim it.", rewardType, amount),
		TargetType: domain.TargetUser,
		Data: map[string]string{
			"type":        "reward_earned",
			"reward_type": rewardType,
			"amount":      fmt.Sprintf("%.2f", amount),
		},
	}

	if err := s.SendToUser(ctx, userID, notification); err != nil {
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to send reward earned notification")
	}
}

// NotifyBasketReady broadcasts a notification that a new basket is available.
func (s *NotificationService) NotifyBasketReady(ctx context.Context) {
	notification := &domain.Notification{
		Title:      "New basket is ready!",
		Body:       "Today's prediction cards are available. Start playing now!",
		TargetType: domain.TargetAll,
		Data: map[string]string{
			"type": "basket_ready",
		},
	}

	if err := s.SendToAll(ctx, notification); err != nil {
		log.Warn().Err(err).Msg("failed to send basket ready notification")
	}
}
