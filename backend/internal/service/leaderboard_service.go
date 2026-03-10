package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

type LeaderboardService struct {
	repo  *postgres.LeaderboardRepo
	cache *redis.LeaderboardCache
	users *postgres.UserRepo
	hub   *ws.Hub
}

func NewLeaderboardService(repo *postgres.LeaderboardRepo, cache *redis.LeaderboardCache, users *postgres.UserRepo, hub *ws.Hub) *LeaderboardService {
	return &LeaderboardService{
		repo:  repo,
		cache: cache,
		users: users,
		hub:   hub,
	}
}

// UpdateLeaderboard updates all applicable leaderboard periods for a user after an answer is resolved.
func (s *LeaderboardService) UpdateLeaderboard(ctx context.Context, userID uuid.UUID, points int, isCorrect bool, eventID *uuid.UUID) error {
	correct := 0
	wrong := 0
	if isCorrect {
		correct = 1
	} else {
		wrong = 1
	}

	now := time.Now().UTC()

	// Periods to update: daily, weekly, all-time, and optionally tournament
	periods := []struct {
		periodType string
		periodKey  string
		eventID    *uuid.UUID
	}{
		{domain.PeriodDaily, GetDailyKey(now), nil},
		{domain.PeriodWeekly, GetWeeklyKey(now), nil},
		{domain.PeriodAllTime, "all", nil},
	}

	if eventID != nil {
		periods = append(periods, struct {
			periodType string
			periodKey  string
			eventID    *uuid.UUID
		}{domain.PeriodTournament, eventID.String(), eventID})
	}

	userIDStr := userID.String()

	for _, p := range periods {
		// Update PostgreSQL
		entry := &domain.LeaderboardEntry{
			ID:             uuid.New(),
			UserID:         userID,
			EventID:        p.eventID,
			PeriodType:     p.periodType,
			PeriodKey:      p.periodKey,
			TotalPoints:    points,
			CorrectAnswers: correct,
			WrongAnswers:   wrong,
			TotalAnswers:   1,
		}
		if err := s.repo.UpsertEntry(ctx, entry); err != nil {
			return fmt.Errorf("update leaderboard %s/%s: %w", p.periodType, p.periodKey, err)
		}

		// Update Redis sorted set
		key := redis.LeaderboardKey(p.periodType, p.periodKey)
		if err := s.cache.UpdateScore(ctx, key, userIDStr, float64(points)); err != nil {
			// Log but don't fail — Redis is best-effort cache
			log.Warn().Err(err).
				Str("period_type", p.periodType).
				Str("period_key", p.periodKey).
				Msg("failed to update leaderboard cache")
		}

		// Set expiry for daily and weekly keys
		switch p.periodType {
		case domain.PeriodDaily:
			_ = s.cache.SetExpiry(ctx, key, 48*time.Hour)
		case domain.PeriodWeekly:
			_ = s.cache.SetExpiry(ctx, key, 10*24*time.Hour)
		}

		// Broadcast leaderboard_update to all connected users
		if s.hub != nil {
			s.hub.Broadcast(ws.Message{
				Type: "leaderboard_update",
				Data: map[string]interface{}{
					"period_type": p.periodType,
					"period_key":  p.periodKey,
				},
			})
		}
	}

	return nil
}

// GetLeaderboard returns a leaderboard page. It tries Redis first for fast reads, falling back to PostgreSQL.
// It always includes the requesting user's own rank.
func (s *LeaderboardService) GetLeaderboard(ctx context.Context, periodType, periodKey string, limit, offset int, userID uuid.UUID) (*domain.LeaderboardResponse, error) {
	var entries []domain.LeaderboardRow
	var total int
	var userRank *domain.LeaderboardRow

	// Try Redis for the top entries (only when offset is 0 and it's a simple top-N query)
	if offset == 0 {
		key := redis.LeaderboardKey(periodType, periodKey)
		cached, err := s.cache.GetTopN(ctx, key, int64(limit))
		if err == nil && len(cached) > 0 {
			// Enrich cached rows with user display info from DB
			entries, err = s.enrichRows(ctx, cached)
			if err != nil {
				// Fall back to PostgreSQL
				entries = nil
			}
		}
	}

	// Fall back to PostgreSQL
	if entries == nil {
		var err error
		entries, err = s.repo.GetRanking(ctx, periodType, periodKey, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("get leaderboard ranking: %w", err)
		}
	}

	// Get total count
	var err error
	total, err = s.repo.CountEntries(ctx, periodType, periodKey)
	if err != nil {
		return nil, fmt.Errorf("count leaderboard entries: %w", err)
	}

	// Get requesting user's rank
	userRank, err = s.repo.GetUserRank(ctx, userID, periodType, periodKey)
	if err != nil {
		// Log but don't fail
		log.Warn().Err(err).Str("user_id", userID.String()).Msg("failed to get user rank")
	}

	return &domain.LeaderboardResponse{
		PeriodType: periodType,
		PeriodKey:  periodKey,
		Entries:    entries,
		UserRank:   userRank,
		Total:      total,
	}, nil
}

// GetFriendsLeaderboard returns a leaderboard filtered to a user's friends (referral connections + mini-league members).
func (s *LeaderboardService) GetFriendsLeaderboard(ctx context.Context, friendIDs []uuid.UUID, periodType, periodKey string, limit, offset int, userID uuid.UUID) (*domain.LeaderboardResponse, error) {
	// Always include the requesting user
	idSet := make(map[uuid.UUID]bool, len(friendIDs)+1)
	idSet[userID] = true
	for _, id := range friendIDs {
		idSet[id] = true
	}

	allIDs := make([]uuid.UUID, 0, len(idSet))
	for id := range idSet {
		allIDs = append(allIDs, id)
	}

	entries, err := s.repo.GetRankingForUsers(ctx, periodType, periodKey, allIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get friends leaderboard: %w", err)
	}

	// Find the requesting user's entry in results
	var userRank *domain.LeaderboardRow
	for _, e := range entries {
		if e.UserID == userID {
			row := e
			userRank = &row
			break
		}
	}

	return &domain.LeaderboardResponse{
		PeriodType: periodType,
		PeriodKey:  periodKey,
		Entries:    entries,
		UserRank:   userRank,
		Total:      len(allIDs),
	}, nil
}

// enrichRows fills in DisplayName and AvatarURL for Redis-sourced rows by looking up each user.
func (s *LeaderboardService) enrichRows(ctx context.Context, rows []domain.LeaderboardRow) ([]domain.LeaderboardRow, error) {
	enriched := make([]domain.LeaderboardRow, 0, len(rows))
	for _, row := range rows {
		user, err := s.users.FindByID(ctx, row.UserID)
		if err != nil {
			return nil, fmt.Errorf("enrich leaderboard row: %w", err)
		}
		if user != nil {
			row.DisplayName = user.DisplayName
			row.AvatarURL = user.AvatarURL
		}
		enriched = append(enriched, row)
	}
	return enriched, nil
}

// GetDailyKey returns the date string for a given time (e.g., "2026-03-10").
func GetDailyKey(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// GetWeeklyKey returns the Monday date of the week containing t (e.g., "2026-03-09").
func GetWeeklyKey(t time.Time) string {
	t = t.UTC()
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -int(weekday-time.Monday))
	return monday.Format("2006-01-02")
}
