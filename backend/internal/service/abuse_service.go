package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// AbuseService detects and manages suspicious activity flags.
type AbuseService struct {
	abuseRepo  *postgres.AbuseRepo
	userRepo   *postgres.UserRepo
	rewardRepo *postgres.RewardRepo
}

func NewAbuseService(abuseRepo *postgres.AbuseRepo, userRepo *postgres.UserRepo, rewardRepo *postgres.RewardRepo) *AbuseService {
	return &AbuseService{
		abuseRepo:  abuseRepo,
		userRepo:   userRepo,
		rewardRepo: rewardRepo,
	}
}

// CheckPerfectScore flags a user if they got 100% correct on a new account (created < 7 days ago).
func (s *AbuseService) CheckPerfectScore(ctx context.Context, userID uuid.UUID, correctCount, totalCount int) {
	if totalCount == 0 || correctCount < totalCount {
		return
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return
	}

	accountAge := time.Since(user.CreatedAt)
	if accountAge > 7*24*time.Hour {
		return // account is older than 7 days, not suspicious
	}

	// Check if we already flagged this user for perfect score recently
	existing, err := s.abuseRepo.CountByUserAndType(ctx, userID, domain.AbuseFlagPerfectScore)
	if err != nil {
		log.Error().Err(err).Msg("failed to check existing perfect score flags")
		return
	}
	if existing > 0 {
		return // already flagged
	}

	details, _ := json.Marshal(map[string]interface{}{
		"correct_count": correctCount,
		"total_count":   totalCount,
		"account_age_h": int(accountAge.Hours()),
	})

	flag := &domain.AbuseFlag{
		ID:       uuid.New(),
		UserID:   userID,
		FlagType: domain.AbuseFlagPerfectScore,
		Details:  details,
		Status:   domain.AbuseFlagStatusPending,
	}

	if err := s.abuseRepo.Create(ctx, flag); err != nil {
		log.Error().Err(err).Msg("failed to create perfect score abuse flag")
	}
}

// CheckVelocity flags a user if they claimed more than the daily token cap.
// dailyCap is the maximum token amount allowed per day.
func (s *AbuseService) CheckVelocity(ctx context.Context, userID uuid.UUID, dailyCap float64) {
	rewards, err := s.rewardRepo.FindPendingByUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check velocity rewards")
		return
	}

	var totalTokens float64
	for _, r := range rewards {
		if r.RewardType == domain.RewardToken {
			totalTokens += r.Amount
		}
	}

	if totalTokens <= dailyCap {
		return
	}

	details, _ := json.Marshal(map[string]interface{}{
		"total_tokens": totalTokens,
		"daily_cap":    dailyCap,
	})

	flag := &domain.AbuseFlag{
		ID:       uuid.New(),
		UserID:   userID,
		FlagType: domain.AbuseFlagVelocity,
		Details:  details,
		Status:   domain.AbuseFlagStatusPending,
	}

	if err := s.abuseRepo.Create(ctx, flag); err != nil {
		log.Error().Err(err).Msg("failed to create velocity abuse flag")
	}
}

// CheckMultiAccount flags a user if other accounts share the same device_id or IP.
func (s *AbuseService) CheckMultiAccount(ctx context.Context, userID uuid.UUID, deviceID, ip string) {
	if deviceID == "" && ip == "" {
		return
	}

	matches, err := s.userRepo.FindByDeviceIDOrIP(ctx, userID, deviceID, ip)
	if err != nil {
		log.Error().Err(err).Msg("failed to check multi-account matches")
		return
	}

	if len(matches) == 0 {
		return
	}

	// Check if we already flagged this user for multi-account
	existing, err := s.abuseRepo.CountByUserAndType(ctx, userID, domain.AbuseFlagMultiAccount)
	if err != nil {
		log.Error().Err(err).Msg("failed to check existing multi-account flags")
		return
	}
	if existing > 0 {
		return // already flagged
	}

	matchingIDs := make([]string, 0, len(matches))
	for _, m := range matches {
		matchingIDs = append(matchingIDs, m.ID.String())
	}

	details, _ := json.Marshal(map[string]interface{}{
		"device_id":    deviceID,
		"ip":           ip,
		"matching_ids": matchingIDs,
	})

	flag := &domain.AbuseFlag{
		ID:       uuid.New(),
		UserID:   userID,
		FlagType: domain.AbuseFlagMultiAccount,
		Details:  details,
		Status:   domain.AbuseFlagStatusPending,
	}

	if err := s.abuseRepo.Create(ctx, flag); err != nil {
		log.Error().Err(err).Msg("failed to create multi-account abuse flag")
	}
}

// GetPendingFlags returns all pending abuse flags for admin review.
func (s *AbuseService) GetPendingFlags(ctx context.Context, limit, offset int) ([]domain.AbuseFlag, error) {
	flags, err := s.abuseRepo.FindPending(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get pending flags: %w", err)
	}
	return flags, nil
}

// ReviewFlag updates the status of an abuse flag after admin review.
func (s *AbuseService) ReviewFlag(ctx context.Context, flagID, adminID uuid.UUID, status string) error {
	if status != domain.AbuseFlagStatusReviewed && status != domain.AbuseFlagStatusDismissed {
		return fmt.Errorf("invalid review status: %s", status)
	}

	if err := s.abuseRepo.UpdateStatus(ctx, flagID, adminID, status); err != nil {
		return fmt.Errorf("review flag: %w", err)
	}
	return nil
}

// GetFlagsByUser returns all abuse flags for a specific user.
func (s *AbuseService) GetFlagsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.AbuseFlag, error) {
	flags, err := s.abuseRepo.FindByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get flags by user: %w", err)
	}
	return flags, nil
}
