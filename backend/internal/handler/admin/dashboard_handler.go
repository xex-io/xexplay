package admin

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

// DashboardHandler handles admin dashboard and extended admin endpoints.
type DashboardHandler struct {
	leaderboardService  *service.LeaderboardService
	auditService        *service.AuditService
	userRepo            *postgres.UserRepo
	sessionRepo         *postgres.SessionRepo
	answerRepo          *postgres.AnswerRepo
	abuseRepo           *postgres.AbuseRepo
	referralRepo        *postgres.ReferralRepo
	rewardRepo          *postgres.RewardRepo
	notifHistoryRepo    *postgres.NotificationHistoryRepo
	prizePoolRepo       *postgres.PrizePoolRepo
}

func NewDashboardHandler(
	leaderboardService *service.LeaderboardService,
	auditService *service.AuditService,
	userRepo *postgres.UserRepo,
	sessionRepo *postgres.SessionRepo,
	answerRepo *postgres.AnswerRepo,
	abuseRepo *postgres.AbuseRepo,
	referralRepo *postgres.ReferralRepo,
	rewardRepo *postgres.RewardRepo,
	notifHistoryRepo *postgres.NotificationHistoryRepo,
	prizePoolRepo *postgres.PrizePoolRepo,
) *DashboardHandler {
	return &DashboardHandler{
		leaderboardService:  leaderboardService,
		auditService:        auditService,
		userRepo:            userRepo,
		sessionRepo:         sessionRepo,
		answerRepo:          answerRepo,
		abuseRepo:           abuseRepo,
		referralRepo:        referralRepo,
		rewardRepo:          rewardRepo,
		notifHistoryRepo:    notifHistoryRepo,
		prizePoolRepo:       prizePoolRepo,
	}
}

// --- 1. GET /admin/leaderboards/:type ---

// AdminGetLeaderboard returns leaderboard data for the admin panel.
func (h *DashboardHandler) AdminGetLeaderboard(c *gin.Context) {
	lbType := c.Param("type")
	limit, offset := parsePagination(c)

	var periodType, periodKey string
	switch lbType {
	case "daily":
		periodType = domain.PeriodDaily
		periodKey = c.DefaultQuery("period_key", service.GetDailyKey(time.Now().UTC()))
	case "weekly":
		periodType = domain.PeriodWeekly
		periodKey = c.DefaultQuery("period_key", service.GetWeeklyKey(time.Now().UTC()))
	case "tournament":
		periodType = domain.PeriodTournament
		eventID := c.Query("event_id")
		if eventID == "" {
			response.BadRequest(c, "event_id is required for tournament leaderboard")
			return
		}
		if _, err := uuid.Parse(eventID); err != nil {
			response.BadRequest(c, "invalid event_id")
			return
		}
		periodKey = eventID
	case "all-time":
		periodType = domain.PeriodAllTime
		periodKey = "all"
	default:
		response.BadRequest(c, "invalid leaderboard type: must be daily, weekly, tournament, or all-time")
		return
	}

	result, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), periodType, periodKey, limit, offset, uuid.Nil)
	if err != nil {
		response.InternalError(c, "failed to fetch leaderboard")
		return
	}

	response.OK(c, result)
}

// --- 2. GET /admin/analytics/overview ---

// AdminGetAnalytics returns platform analytics overview.
func (h *DashboardHandler) AdminGetAnalytics(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now().UTC()

	totalUsers, err := h.userRepo.Count(ctx)
	if err != nil {
		response.InternalError(c, "failed to count users")
		return
	}

	dau, err := h.userRepo.CountActiveUsers(ctx, now.Truncate(24*time.Hour))
	if err != nil {
		response.InternalError(c, "failed to count DAU")
		return
	}

	wau, err := h.userRepo.CountActiveUsers(ctx, now.AddDate(0, 0, -7))
	if err != nil {
		response.InternalError(c, "failed to count WAU")
		return
	}

	mau, err := h.userRepo.CountActiveUsers(ctx, now.AddDate(0, 0, -30))
	if err != nil {
		response.InternalError(c, "failed to count MAU")
		return
	}

	totalSessions, err := h.sessionRepo.CountTotal(ctx)
	if err != nil {
		response.InternalError(c, "failed to count sessions")
		return
	}

	completedSessions, err := h.sessionRepo.CountCompleted(ctx)
	if err != nil {
		response.InternalError(c, "failed to count completed sessions")
		return
	}

	var completionRate float64
	if totalSessions > 0 {
		completionRate = float64(completedSessions) / float64(totalSessions) * 100
	}

	correct, incorrect, err := h.answerRepo.CountCorrectIncorrect(ctx)
	if err != nil {
		response.InternalError(c, "failed to count answers")
		return
	}

	response.OK(c, gin.H{
		"total_users":            totalUsers,
		"dau":                    dau,
		"wau":                    wau,
		"mau":                    mau,
		"total_sessions":         totalSessions,
		"completed_sessions":     completedSessions,
		"session_completion_rate": completionRate,
		"answers": gin.H{
			"correct":   correct,
			"incorrect": incorrect,
		},
	})
}

// --- 3. GET /admin/exchange/metrics ---

// AdminGetExchangeMetrics returns exchange-related statistics.
func (h *DashboardHandler) AdminGetExchangeMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	linkedCount, err := h.userRepo.CountLinkedExchangeUsers(ctx)
	if err != nil {
		response.InternalError(c, "failed to count linked users")
		return
	}

	tierDist, err := h.userRepo.TradingTierDistribution(ctx)
	if err != nil {
		response.InternalError(c, "failed to get tier distribution")
		return
	}

	claimsByType, err := h.rewardRepo.CountClaimsByType(ctx)
	if err != nil {
		response.InternalError(c, "failed to get reward claims")
		return
	}

	response.OK(c, gin.H{
		"linked_exchange_users":   linkedCount,
		"trading_tier_distribution": tierDist,
		"reward_claims_by_type":   claimsByType,
	})
}

// --- 4. GET /admin/users/search?q= ---

// AdminSearchUsers searches users by email or display name.
func (h *DashboardHandler) AdminSearchUsers(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		response.BadRequest(c, "search query 'q' is required")
		return
	}

	limit, offset := parsePagination(c)

	users, err := h.userRepo.Search(c.Request.Context(), q, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to search users")
		return
	}

	response.OK(c, users)
}

// --- 5. GET /admin/users/:id/activity ---

// AdminGetUserActivity returns a user's recent sessions, answers, and rewards.
func (h *DashboardHandler) AdminGetUserActivity(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	ctx := c.Request.Context()

	sessions, err := h.sessionRepo.FindByUserID(ctx, userID, 20)
	if err != nil {
		response.InternalError(c, "failed to fetch sessions")
		return
	}

	answers, err := h.answerRepo.FindByUserID(ctx, userID, 50)
	if err != nil {
		response.InternalError(c, "failed to fetch answers")
		return
	}

	rewards, err := h.rewardRepo.FindByUser(ctx, userID, 20, 0)
	if err != nil {
		response.InternalError(c, "failed to fetch rewards")
		return
	}

	response.OK(c, gin.H{
		"sessions": sessions,
		"answers":  answers,
		"rewards":  rewards,
	})
}

// --- 6. POST /admin/users/:id/moderate ---

type moderateRequest struct {
	Action string `json:"action" binding:"required,oneof=suspend ban activate"`
	Reason string `json:"reason"`
}

// AdminModerateUser moderates a user (suspend/ban/activate).
func (h *DashboardHandler) AdminModerateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req moderateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: action must be suspend, ban, or activate")
		return
	}

	ctx := c.Request.Context()

	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		response.InternalError(c, "failed to fetch user")
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	var isActive bool
	switch req.Action {
	case "activate":
		isActive = true
	case "suspend", "ban":
		isActive = false
	}

	if err := h.userRepo.UpdateIsActive(ctx, userID, isActive); err != nil {
		response.InternalError(c, "failed to update user status")
		return
	}

	// Create audit log
	adminUserID, _ := c.Get("user_id")
	adminID := adminUserID.(uuid.UUID)

	h.auditService.LogAction(ctx, adminID, domain.AuditActionUserBanned, domain.EntityTypeUser, userID.String(),
		map[string]string{"action": req.Action, "reason": req.Reason}, c.ClientIP())

	response.OK(c, gin.H{
		"message": "user moderation applied",
		"action":  req.Action,
		"user_id": userID,
	})
}

// --- 7. GET /admin/abuse-flags/stats ---

// AdminGetAbuseFlagStats returns abuse flag statistics.
func (h *DashboardHandler) AdminGetAbuseFlagStats(c *gin.Context) {
	ctx := c.Request.Context()

	byStatus, err := h.abuseRepo.CountByStatus(ctx)
	if err != nil {
		response.InternalError(c, "failed to get abuse flag stats by status")
		return
	}

	byType, err := h.abuseRepo.CountByType(ctx)
	if err != nil {
		response.InternalError(c, "failed to get abuse flag stats by type")
		return
	}

	response.OK(c, gin.H{
		"by_status": byStatus,
		"by_type":   byType,
	})
}

// --- 8. GET /admin/notifications ---

// AdminListNotifications returns notification send history.
func (h *DashboardHandler) AdminListNotifications(c *gin.Context) {
	limit, offset := parsePagination(c)

	history, err := h.notifHistoryRepo.FindRecent(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to fetch notification history")
		return
	}

	response.OK(c, history)
}

// --- 9. GET /admin/referrals/stats ---

// AdminGetReferralStats returns platform-wide referral statistics.
func (h *DashboardHandler) AdminGetReferralStats(c *gin.Context) {
	ctx := c.Request.Context()

	total, err := h.referralRepo.CountAll(ctx)
	if err != nil {
		response.InternalError(c, "failed to count referrals")
		return
	}

	converted, err := h.referralRepo.CountConverted(ctx)
	if err != nil {
		response.InternalError(c, "failed to count converted referrals")
		return
	}

	activeReferrers, err := h.referralRepo.CountActiveReferrers(ctx)
	if err != nil {
		response.InternalError(c, "failed to count active referrers")
		return
	}

	var conversionRate float64
	if total > 0 {
		conversionRate = float64(converted) / float64(total) * 100
	}

	response.OK(c, gin.H{
		"total_referrals":   total,
		"converted":         converted,
		"conversion_rate":   conversionRate,
		"active_referrers":  activeReferrers,
	})
}

// --- 10. GET /admin/referrals/top ---

// AdminGetTopReferrers returns the top referrers by referral count.
func (h *DashboardHandler) AdminGetTopReferrers(c *gin.Context) {
	limit, _ := parsePagination(c)

	topReferrers, err := h.referralRepo.TopReferrers(c.Request.Context(), limit)
	if err != nil {
		response.InternalError(c, "failed to fetch top referrers")
		return
	}

	response.OK(c, topReferrers)
}

// --- 11. GET /admin/prize-pools ---

// AdminListPrizePools returns active/completed prize pools.
func (h *DashboardHandler) AdminListPrizePools(c *gin.Context) {
	limit, offset := parsePagination(c)
	status := c.Query("status")

	var pools []domain.PrizePool
	var err error

	if status != "" {
		pools, err = h.prizePoolRepo.FindByStatus(c.Request.Context(), status, limit, offset)
	} else {
		pools, err = h.prizePoolRepo.FindAll(c.Request.Context(), limit, offset)
	}

	if err != nil {
		response.InternalError(c, "failed to fetch prize pools")
		return
	}

	response.OK(c, pools)
}

// --- 12. GET /admin/prize-pools/history ---

// AdminGetPrizePoolHistory returns prize pool distribution history.
func (h *DashboardHandler) AdminGetPrizePoolHistory(c *gin.Context) {
	limit, offset := parsePagination(c)

	distributions, err := h.prizePoolRepo.FindAllDistributions(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to fetch prize pool history")
		return
	}

	response.OK(c, distributions)
}

// --- 13. POST /admin/prize-pools ---

type createPrizePoolRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	TotalAmount float64  `json:"total_amount" binding:"required,min=0"`
	Currency    string   `json:"currency"`
	StartDate   *string  `json:"start_date"`
	EndDate     *string  `json:"end_date"`
}

// AdminCreatePrizePool creates a new prize pool.
func (h *DashboardHandler) AdminCreatePrizePool(c *gin.Context) {
	var req createPrizePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	currency := req.Currency
	if currency == "" {
		currency = "XEX"
	}

	adminUserID, _ := c.Get("user_id")
	adminID := adminUserID.(uuid.UUID)

	pool := &domain.PrizePool{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		TotalAmount: req.TotalAmount,
		Currency:    currency,
		Status:      domain.PrizePoolStatusActive,
		CreatedBy:   &adminID,
	}

	if req.StartDate != nil {
		t, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			response.BadRequest(c, "invalid start_date format, use RFC3339")
			return
		}
		pool.StartDate = &t
	}

	if req.EndDate != nil {
		t, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			response.BadRequest(c, "invalid end_date format, use RFC3339")
			return
		}
		pool.EndDate = &t
	}

	if err := h.prizePoolRepo.Create(c.Request.Context(), pool); err != nil {
		response.InternalError(c, "failed to create prize pool")
		return
	}

	// Audit log
	h.auditService.LogAction(c.Request.Context(), adminID, "prize_pool_created", "prize_pool", pool.ID.String(),
		map[string]interface{}{"name": req.Name, "amount": req.TotalAmount}, c.ClientIP())

	response.Created(c, pool)
}

// --- helpers ---

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = 50
	offset = 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return limit, offset
}

// Ensure json import is used (for moderateRequest audit details).
var _ = json.RawMessage{}
