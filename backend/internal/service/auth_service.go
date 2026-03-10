package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	jwtpkg "github.com/xex-exchange/xexplay-api/internal/pkg/jwt"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type AuthService struct {
	userRepo        *postgres.UserRepo
	referralService *ReferralService
	jwtSecret       string
}

func NewAuthService(userRepo *postgres.UserRepo, referralService *ReferralService, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		referralService: referralService,
		jwtSecret:       jwtSecret,
	}
}

// Login parses the Exchange JWT, upserts the user (creating on first login), and returns the user.
// If referralCode is non-empty on a new user, it applies the referral.
// deviceID and ip are stored for device fingerprinting.
func (s *AuthService) Login(ctx context.Context, tokenString, referralCode, deviceID, ip string) (*domain.User, error) {
	claims, err := jwtpkg.Parse(tokenString, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if user already exists
	existing, err := s.userRepo.FindByXexUserID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if existing != nil {
		// Update email from token in case it changed
		existing.Email = claims.Email
		if err := s.userRepo.Upsert(ctx, existing); err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
		// Update device info on every login
		now := time.Now().UTC()
		if err := s.userRepo.UpdateDeviceInfo(ctx, existing.ID, deviceID, ip, now); err != nil {
			log.Warn().Err(err).Str("user_id", existing.ID.String()).Msg("failed to update device info")
		}
		return existing, nil
	}

	// First login — create new user
	refCode, err := generateReferralCode()
	if err != nil {
		return nil, fmt.Errorf("generate referral code: %w", err)
	}

	user := &domain.User{
		ID:           uuid.New(),
		XexUserID:    claims.UserID,
		DisplayName:  "",
		Email:        claims.Email,
		Role:         domain.RoleUser,
		ReferralCode: refCode,
		Language:     "en",
		TotalPoints:  0,
		IsActive:     true,
	}

	if err := s.userRepo.Upsert(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Update device info for the new user
	now := time.Now().UTC()
	if err := s.userRepo.UpdateDeviceInfo(ctx, user.ID, deviceID, ip, now); err != nil {
		log.Warn().Err(err).Str("user_id", user.ID.String()).Msg("failed to update device info on new user")
	}

	// Apply referral code if provided
	if referralCode != "" && s.referralService != nil {
		if err := s.referralService.ApplyReferral(ctx, referralCode, user.ID); err != nil {
			log.Warn().Err(err).
				Str("referral_code", referralCode).
				Str("user_id", user.ID.String()).
				Msg("failed to apply referral code during login")
		}
	}

	return user, nil
}

// generateReferralCode generates a cryptographically random 8-character alphanumeric string.
func generateReferralCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	const length = 8

	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
