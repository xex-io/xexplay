package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type RewardService struct {
	rewardRepo *postgres.RewardRepo
	userRepo   *postgres.UserRepo
	hub        *ws.Hub
}

func NewRewardService(rewardRepo *postgres.RewardRepo, userRepo *postgres.UserRepo, hub *ws.Hub) *RewardService {
	return &RewardService{rewardRepo: rewardRepo, userRepo: userRepo, hub: hub}
}

// DistributeRewards matches leaderboard entries against active reward configs for the given
// period type and creates reward distribution records for qualifying users.
func (s *RewardService) DistributeRewards(ctx context.Context, periodType, periodKey string, entries []domain.RewardLeaderboardEntry) (int, error) {
	configs, err := s.rewardRepo.FindConfigsByPeriod(ctx, periodType)
	if err != nil {
		return 0, fmt.Errorf("find reward configs: %w", err)
	}

	if len(configs) == 0 {
		return 0, nil
	}

	created := 0
	for _, entry := range entries {
		for _, cfg := range configs {
			if entry.Rank >= cfg.RankFrom && entry.Rank <= cfg.RankTo {
				dist := &domain.RewardDistribution{
					ID:             uuid.New(),
					UserID:         entry.UserID,
					RewardConfigID: &cfg.ID,
					PeriodType:     periodType,
					PeriodKey:      periodKey,
					RewardType:     cfg.RewardType,
					Amount:         cfg.Amount,
					Rank:           entry.Rank,
					Status:         domain.StatusPending,
				}
				if err := s.rewardRepo.CreateDistribution(ctx, dist); err != nil {
					return created, fmt.Errorf("create distribution for user %s: %w", entry.UserID, err)
				}
				created++

				// Send reward_earned notification to the user
				if s.hub != nil {
					s.hub.SendToUser(entry.UserID, ws.Message{
						Type: "reward_earned",
						Data: map[string]interface{}{
							"reward_type": cfg.RewardType,
							"amount":      cfg.Amount,
						},
					})
				}
			}
		}
	}

	return created, nil
}

// GetPendingRewards returns all pending rewards for a user.
func (s *RewardService) GetPendingRewards(ctx context.Context, userID uuid.UUID) ([]domain.RewardDistribution, error) {
	rewards, err := s.rewardRepo.FindPendingByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get pending rewards: %w", err)
	}
	return rewards, nil
}

// GetRewardHistory returns all rewards for a user with pagination.
func (s *RewardService) GetRewardHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.RewardDistribution, error) {
	rewards, err := s.rewardRepo.FindByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get reward history: %w", err)
	}
	return rewards, nil
}

// ClaimReward claims a pending reward for a user.
// Enforces minimum account age (7 days) and reward hold period (24 hours).
func (s *RewardService) ClaimReward(ctx context.Context, distributionID, userID uuid.UUID) error {
	// Check minimum account age
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}
	if time.Since(user.CreatedAt) < 7*24*time.Hour {
		return fmt.Errorf("Account must be at least 7 days old to claim rewards")
	}

	// Check reward hold period (24 hours)
	dist, err := s.rewardRepo.FindDistributionByID(ctx, distributionID)
	if err != nil {
		return fmt.Errorf("find distribution: %w", err)
	}
	if dist == nil {
		return fmt.Errorf("reward not found")
	}
	if time.Since(dist.CreatedAt) < 24*time.Hour {
		return fmt.Errorf("Rewards can be claimed after 24 hours")
	}

	if err := s.rewardRepo.ClaimReward(ctx, distributionID, userID); err != nil {
		return fmt.Errorf("claim reward: %w", err)
	}
	return nil
}

// CreditReward marks a claimed reward as credited (after Exchange API confirms).
func (s *RewardService) CreditReward(ctx context.Context, distributionID, userID uuid.UUID) error {
	if err := s.rewardRepo.CreditReward(ctx, distributionID, userID); err != nil {
		return fmt.Errorf("credit reward: %w", err)
	}
	return nil
}

// CreateConfig creates a new reward config (admin).
func (s *RewardService) CreateConfig(ctx context.Context, cfg *domain.RewardConfig) error {
	return s.rewardRepo.CreateConfig(ctx, cfg)
}

// UpdateConfig updates an existing reward config (admin).
func (s *RewardService) UpdateConfig(ctx context.Context, cfg *domain.RewardConfig) error {
	return s.rewardRepo.UpdateConfig(ctx, cfg)
}

// ListAllConfigs returns all reward configs (admin).
func (s *RewardService) ListAllConfigs(ctx context.Context) ([]domain.RewardConfig, error) {
	return s.rewardRepo.FindAllConfigs(ctx)
}

// GetDistributionHistory returns all distributions with pagination (admin).
func (s *RewardService) GetDistributionHistory(ctx context.Context, limit, offset int) ([]domain.RewardDistribution, error) {
	return s.rewardRepo.FindDistributionHistory(ctx, limit, offset)
}
