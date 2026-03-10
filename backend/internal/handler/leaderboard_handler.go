package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
	maxLimit      = 100
)

type LeaderboardHandler struct {
	leaderboardService *service.LeaderboardService
}

func NewLeaderboardHandler(leaderboardService *service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboardService: leaderboardService}
}

// GetDaily handles GET /leaderboards/daily
func (h *LeaderboardHandler) GetDaily(c *gin.Context) {
	userID := getUserID(c)
	limit, offset := parsePagination(c)

	periodKey := c.Query("date")
	if periodKey == "" {
		periodKey = service.GetDailyKey(time.Now().UTC())
	}

	result, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), "daily", periodKey, limit, offset, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

// GetWeekly handles GET /leaderboards/weekly
func (h *LeaderboardHandler) GetWeekly(c *gin.Context) {
	userID := getUserID(c)
	limit, offset := parsePagination(c)

	periodKey := c.Query("week")
	if periodKey == "" {
		periodKey = service.GetWeeklyKey(time.Now().UTC())
	}

	result, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), "weekly", periodKey, limit, offset, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

// GetTournament handles GET /leaderboards/tournament/:eventId
func (h *LeaderboardHandler) GetTournament(c *gin.Context) {
	userID := getUserID(c)
	limit, offset := parsePagination(c)

	eventIDParam := c.Param("eventId")
	eventID, err := uuid.Parse(eventIDParam)
	if err != nil {
		response.BadRequest(c, "invalid event id")
		return
	}

	result, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), "tournament", eventID.String(), limit, offset, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

// GetAllTime handles GET /leaderboards/all-time
func (h *LeaderboardHandler) GetAllTime(c *gin.Context) {
	userID := getUserID(c)
	limit, offset := parsePagination(c)

	result, err := h.leaderboardService.GetLeaderboard(c.Request.Context(), "all_time", "all", limit, offset, userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, result)
}

func getUserID(c *gin.Context) uuid.UUID {
	userID, _ := c.Get("user_id")
	uid, ok := userID.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return uid
}

func parsePagination(c *gin.Context) (limit, offset int) {
	limit = defaultLimit
	offset = defaultOffset

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > maxLimit {
				limit = maxLimit
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
