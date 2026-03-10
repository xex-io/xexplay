package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

type MiniLeagueService struct {
	leagueRepo      *postgres.MiniLeagueRepo
	leaderboardRepo *postgres.LeaderboardRepo
	userRepo        *postgres.UserRepo
	lbCache         *redis.LeaderboardCache
}

func NewMiniLeagueService(
	leagueRepo *postgres.MiniLeagueRepo,
	leaderboardRepo *postgres.LeaderboardRepo,
	userRepo *postgres.UserRepo,
	lbCache *redis.LeaderboardCache,
) *MiniLeagueService {
	return &MiniLeagueService{
		leagueRepo:      leagueRepo,
		leaderboardRepo: leaderboardRepo,
		userRepo:        userRepo,
		lbCache:         lbCache,
	}
}

// CreateLeague creates a new mini league and adds the creator as the first member.
func (s *MiniLeagueService) CreateLeague(ctx context.Context, name string, creatorID uuid.UUID, eventID *uuid.UUID) (*domain.MiniLeague, error) {
	code, err := generateInviteCode()
	if err != nil {
		return nil, fmt.Errorf("generate invite code: %w", err)
	}

	league := &domain.MiniLeague{
		ID:         uuid.New(),
		Name:       name,
		CreatorID:  creatorID,
		InviteCode: code,
		EventID:    eventID,
		MaxMembers: 50,
	}

	if err := s.leagueRepo.Create(ctx, league); err != nil {
		return nil, fmt.Errorf("create league: %w", err)
	}

	// Add creator as the first member
	if err := s.leagueRepo.AddMember(ctx, league.ID, creatorID); err != nil {
		return nil, fmt.Errorf("add creator to league: %w", err)
	}

	league.MemberCount = 1
	return league, nil
}

// JoinLeague joins a user to a mini league using an invite code.
func (s *MiniLeagueService) JoinLeague(ctx context.Context, inviteCode string, userID uuid.UUID) (*domain.MiniLeague, error) {
	league, err := s.leagueRepo.FindByInviteCode(ctx, inviteCode)
	if err != nil {
		return nil, fmt.Errorf("find league: %w", err)
	}
	if league == nil {
		return nil, fmt.Errorf("invalid invite code")
	}

	count, err := s.leagueRepo.CountMembers(ctx, league.ID)
	if err != nil {
		return nil, fmt.Errorf("count members: %w", err)
	}
	if count >= league.MaxMembers {
		return nil, fmt.Errorf("league is full")
	}

	if err := s.leagueRepo.AddMember(ctx, league.ID, userID); err != nil {
		return nil, fmt.Errorf("join league: %w", err)
	}

	league.MemberCount = count + 1
	return league, nil
}

// GetUserLeagues returns all mini leagues a user belongs to.
func (s *MiniLeagueService) GetUserLeagues(ctx context.Context, userID uuid.UUID) ([]domain.MiniLeague, error) {
	leagues, err := s.leagueRepo.FindByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user leagues: %w", err)
	}
	if leagues == nil {
		leagues = []domain.MiniLeague{}
	}
	return leagues, nil
}

// GetLeague returns a mini league by its ID.
func (s *MiniLeagueService) GetLeague(ctx context.Context, leagueID uuid.UUID) (*domain.MiniLeague, error) {
	league, err := s.leagueRepo.FindByID(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get league: %w", err)
	}
	return league, nil
}

// GetLeagueLeaderboard returns the leaderboard for members of a mini league.
// It uses the weekly period by default and filters to only league members.
func (s *MiniLeagueService) GetLeagueLeaderboard(ctx context.Context, leagueID uuid.UUID) ([]domain.LeaderboardRow, error) {
	members, err := s.leagueRepo.GetMembers(ctx, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get league members: %w", err)
	}

	if len(members) == 0 {
		return []domain.LeaderboardRow{}, nil
	}

	// Get the current weekly period key
	now := time.Now().UTC()
	periodKey := GetWeeklyKey(now)

	// Get full weekly leaderboard
	allEntries, err := s.leaderboardRepo.GetRanking(ctx, domain.PeriodWeekly, periodKey, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("get weekly leaderboard: %w", err)
	}

	// Filter to only league members
	memberSet := make(map[uuid.UUID]bool, len(members))
	for _, m := range members {
		memberSet[m.UserID] = true
	}

	var leagueEntries []domain.LeaderboardRow
	rank := 0
	for _, entry := range allEntries {
		if memberSet[entry.UserID] {
			rank++
			entry.Rank = rank
			leagueEntries = append(leagueEntries, entry)
		}
	}

	if leagueEntries == nil {
		leagueEntries = []domain.LeaderboardRow{}
	}

	return leagueEntries, nil
}

// generateInviteCode generates a random 8-character hex invite code.
func generateInviteCode() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
