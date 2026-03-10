package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type AchievementService struct {
	achievementRepo *postgres.AchievementRepo
	hub             *ws.Hub
}

func NewAchievementService(achievementRepo *postgres.AchievementRepo, hub *ws.Hub) *AchievementService {
	return &AchievementService{achievementRepo: achievementRepo, hub: hub}
}

// CheckAndGrant checks if a user qualifies for any achievements of the given condition type
// and grants them if the condition value is met and the achievement is not already earned.
func (s *AchievementService) CheckAndGrant(ctx context.Context, userID uuid.UUID, conditionType string, value int) error {
	achievements, err := s.achievementRepo.FindByConditionType(ctx, conditionType)
	if err != nil {
		return fmt.Errorf("find achievements for condition %s: %w", conditionType, err)
	}

	for _, a := range achievements {
		if value < a.ConditionValue {
			continue
		}

		has, err := s.achievementRepo.HasAchievement(ctx, userID, a.ID)
		if err != nil {
			log.Warn().Err(err).
				Str("user_id", userID.String()).
				Str("achievement", a.Key).
				Msg("failed to check existing achievement")
			continue
		}
		if has {
			continue
		}

		if err := s.achievementRepo.Grant(ctx, userID, a.ID); err != nil {
			log.Warn().Err(err).
				Str("user_id", userID.String()).
				Str("achievement", a.Key).
				Msg("failed to grant achievement")
			continue
		}

		log.Info().
			Str("user_id", userID.String()).
			Str("achievement", a.Key).
			Msg("achievement granted")

		// Send achievement_unlocked notification to the user
		if s.hub != nil {
			// Extract English name from i18n JSON, fallback to raw JSON string
			name := extractEnglishName(a.Name)
			s.hub.SendToUser(userID, ws.Message{
				Type: "achievement_unlocked",
				Data: map[string]interface{}{
					"achievement_key": a.Key,
					"name":            name,
				},
			})
		}
	}

	return nil
}

// GetUserAchievements returns all achievements earned by a user.
func (s *AchievementService) GetUserAchievements(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievement, error) {
	achievements, err := s.achievementRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user achievements: %w", err)
	}
	if achievements == nil {
		achievements = []domain.UserAchievement{}
	}
	return achievements, nil
}

// GetAllAchievements returns all defined achievements.
func (s *AchievementService) GetAllAchievements(ctx context.Context) ([]domain.Achievement, error) {
	achievements, err := s.achievementRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all achievements: %w", err)
	}
	if achievements == nil {
		achievements = []domain.Achievement{}
	}
	return achievements, nil
}

// extractEnglishName extracts the "en" value from an i18n JSON field like {"en": "...", "fa": "..."}.
// Falls back to the raw JSON string if parsing fails.
func extractEnglishName(raw json.RawMessage) string {
	var names map[string]string
	if err := json.Unmarshal(raw, &names); err == nil {
		if en, ok := names["en"]; ok {
			return en
		}
	}
	return string(raw)
}
