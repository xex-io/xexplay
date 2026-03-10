package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type StreakService struct {
	streakRepo *postgres.StreakRepo
}

func NewStreakService(streakRepo *postgres.StreakRepo) *StreakService {
	return &StreakService{streakRepo: streakRepo}
}

// UpdateStreak updates the user's streak based on the played date.
// Returns a *StreakMilestone if the new streak exactly hits a milestone, otherwise nil.
func (s *StreakService) UpdateStreak(ctx context.Context, userID uuid.UUID, playedDate time.Time) (*domain.StreakMilestone, error) {
	played := playedDate.UTC().Truncate(24 * time.Hour)

	streak, err := s.streakRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get streak: %w", err)
	}

	if streak == nil {
		streak = &domain.Streak{
			ID:     uuid.New(),
			UserID: userID,
		}
	}

	if streak.LastPlayedDate != nil {
		lastPlayed := streak.LastPlayedDate.UTC().Truncate(24 * time.Hour)
		diff := played.Sub(lastPlayed)

		switch {
		case diff == 0:
			// Already played today — no change
			return nil, nil
		case diff == 24*time.Hour:
			// Consecutive day — increment
			streak.CurrentStreak++
		default:
			// Gap — reset to 1
			streak.CurrentStreak = 1
		}
	} else {
		// First time playing
		streak.CurrentStreak = 1
	}

	// Update longest streak
	if streak.CurrentStreak > streak.LongestStreak {
		streak.LongestStreak = streak.CurrentStreak
	}

	streak.LastPlayedDate = &played

	// Check for milestone
	milestone := domain.CheckMilestone(streak.CurrentStreak)
	if milestone != nil {
		streak.BonusSkips += milestone.BonusSkips
		streak.BonusAnswers += milestone.BonusAnswers
	}

	if err := s.streakRepo.Upsert(ctx, streak); err != nil {
		return nil, fmt.Errorf("update streak: %w", err)
	}

	return milestone, nil
}

// GetStreak returns the current streak info for a user.
func (s *StreakService) GetStreak(ctx context.Context, userID uuid.UUID) (*domain.Streak, error) {
	streak, err := s.streakRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get streak: %w", err)
	}
	if streak == nil {
		return &domain.Streak{
			UserID: userID,
		}, nil
	}
	return streak, nil
}

// ApplyBonuses returns the bonus skips and bonus answers currently available for the user.
// These should be applied to a new session.
func (s *StreakService) ApplyBonuses(ctx context.Context, userID uuid.UUID) (bonusSkips, bonusAnswers int, err error) {
	streak, err := s.streakRepo.FindByUserID(ctx, userID)
	if err != nil {
		return 0, 0, fmt.Errorf("get streak for bonuses: %w", err)
	}
	if streak == nil {
		return 0, 0, nil
	}
	return streak.BonusSkips, streak.BonusAnswers, nil
}
