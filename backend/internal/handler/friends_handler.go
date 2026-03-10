package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

// FriendsHandler handles friend-related leaderboard endpoints.
type FriendsHandler struct {
	leaderboardService *service.LeaderboardService
	referralRepo       *postgres.ReferralRepo
	miniLeagueRepo     *postgres.MiniLeagueRepo
}

// NewFriendsHandler creates a new FriendsHandler.
func NewFriendsHandler(
	leaderboardService *service.LeaderboardService,
	referralRepo *postgres.ReferralRepo,
	miniLeagueRepo *postgres.MiniLeagueRepo,
) *FriendsHandler {
	return &FriendsHandler{
		leaderboardService: leaderboardService,
		referralRepo:       referralRepo,
		miniLeagueRepo:     miniLeagueRepo,
	}
}

// GetFriendsLeaderboard handles GET /leaderboards/friends.
// It combines referral connections and mini-league members into a friends leaderboard.
func (h *FriendsHandler) GetFriendsLeaderboard(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		response.Unauthorized(c, "invalid user")
		return
	}
	limit, offset := parsePagination(c)
	ctx := c.Request.Context()

	// Collect friend user IDs from referral connections
	friendSet := make(map[uuid.UUID]bool)

	// 1. Who referred this user
	referredBy, err := h.referralRepo.FindByReferred(ctx, userID)
	if err == nil && referredBy != nil {
		friendSet[referredBy.ReferrerID] = true
	}

	// 2. Who this user referred
	referrals, err := h.referralRepo.FindByReferrer(ctx, userID)
	if err == nil {
		for _, ref := range referrals {
			friendSet[ref.ReferredID] = true
		}
	}

	// 3. Mini-league members from all user's leagues
	leagues, err := h.miniLeagueRepo.FindByUser(ctx, userID)
	if err == nil {
		for _, league := range leagues {
			members, err := h.miniLeagueRepo.GetMembers(ctx, league.ID)
			if err == nil {
				for _, m := range members {
					if m.UserID != userID {
						friendSet[m.UserID] = true
					}
				}
			}
		}
	}

	// Convert set to slice
	friendIDs := make([]uuid.UUID, 0, len(friendSet))
	for id := range friendSet {
		friendIDs = append(friendIDs, id)
	}

	// Default to weekly period
	periodType := domain.PeriodWeekly
	periodKey := c.Query("week")
	if periodKey == "" {
		periodKey = service.GetWeeklyKey(time.Now().UTC())
	}

	result, err := h.leaderboardService.GetFriendsLeaderboard(ctx, friendIDs, periodType, periodKey, limit, offset, userID)
	if err != nil {
		response.InternalError(c, "failed to get friends leaderboard")
		return
	}

	response.OK(c, result)
}
