package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type ReferralService struct {
	referralRepo *postgres.ReferralRepo
	userRepo     *postgres.UserRepo
}

func NewReferralService(referralRepo *postgres.ReferralRepo, userRepo *postgres.UserRepo) *ReferralService {
	return &ReferralService{
		referralRepo: referralRepo,
		userRepo:     userRepo,
	}
}

// ApplyReferral creates a referral record linking the referrer (by referral code) to the referred user.
func (s *ReferralService) ApplyReferral(ctx context.Context, referrerCode string, referredUserID uuid.UUID) error {
	// Check if the referred user already has a referral
	existing, err := s.referralRepo.FindByReferred(ctx, referredUserID)
	if err != nil {
		return fmt.Errorf("check existing referral: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("user already has a referral")
	}

	// Find the referrer by their referral code
	referrer, err := s.userRepo.FindByReferralCode(ctx, referrerCode)
	if err != nil {
		return fmt.Errorf("find referrer: %w", err)
	}
	if referrer == nil {
		return fmt.Errorf("invalid referral code")
	}

	// Cannot refer yourself
	if referrer.ID == referredUserID {
		return fmt.Errorf("cannot refer yourself")
	}

	ref := &domain.Referral{
		ID:            uuid.New(),
		ReferrerID:    referrer.ID,
		ReferredID:    referredUserID,
		Status:        domain.ReferralStatusSignedUp,
		RewardGranted: false,
	}

	if err := s.referralRepo.Create(ctx, ref); err != nil {
		return fmt.Errorf("create referral: %w", err)
	}

	return nil
}

// UpdateReferralStatus updates the status of a referral for a referred user.
func (s *ReferralService) UpdateReferralStatus(ctx context.Context, referredUserID uuid.UUID, status string) error {
	ref, err := s.referralRepo.FindByReferred(ctx, referredUserID)
	if err != nil {
		return fmt.Errorf("find referral: %w", err)
	}
	if ref == nil {
		return nil // No referral to update
	}

	rewardGranted := ref.RewardGranted
	if status == domain.ReferralStatusFirstSession && !ref.RewardGranted {
		rewardGranted = true
	}

	if err := s.referralRepo.UpdateStatus(ctx, referredUserID, status, rewardGranted); err != nil {
		return fmt.Errorf("update referral status: %w", err)
	}

	return nil
}

// GetReferralStats returns referral statistics for a user.
func (s *ReferralService) GetReferralStats(ctx context.Context, userID uuid.UUID) (*domain.ReferralStats, error) {
	stats, err := s.referralRepo.CountByReferrer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get referral stats: %w", err)
	}
	return stats, nil
}

// GetReferralCode returns the referral code for a user.
func (s *ReferralService) GetReferralCode(ctx context.Context, userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("find user: %w", err)
	}
	if user == nil {
		return "", fmt.Errorf("user not found")
	}
	return user.ReferralCode, nil
}
