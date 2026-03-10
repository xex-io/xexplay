package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type LeagueHandler struct {
	leagueService *service.MiniLeagueService
}

func NewLeagueHandler(leagueService *service.MiniLeagueService) *LeagueHandler {
	return &LeagueHandler{leagueService: leagueService}
}

type createLeagueRequest struct {
	Name    string `json:"name" binding:"required"`
	EventID string `json:"event_id,omitempty"`
}

type joinLeagueRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// CreateLeague handles POST /leagues — creates a new mini league.
func (h *LeagueHandler) CreateLeague(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req createLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	var eventID *uuid.UUID
	if req.EventID != "" {
		parsed, err := uuid.Parse(req.EventID)
		if err != nil {
			response.BadRequest(c, "invalid event_id")
			return
		}
		eventID = &parsed
	}

	league, err := h.leagueService.CreateLeague(c.Request.Context(), req.Name, uid, eventID)
	if err != nil {
		response.InternalError(c, "failed to create league")
		return
	}

	response.Created(c, league)
}

// JoinLeague handles POST /leagues/join — joins a mini league by invite code.
func (h *LeagueHandler) JoinLeague(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req joinLeagueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	league, err := h.leagueService.JoinLeague(c.Request.Context(), req.InviteCode, uid)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, league)
}

// GetMyLeagues handles GET /leagues — returns all leagues the user belongs to.
func (h *LeagueHandler) GetMyLeagues(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	leagues, err := h.leagueService.GetUserLeagues(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get leagues")
		return
	}

	response.OK(c, leagues)
}

// GetLeague handles GET /leagues/:id — returns a single league.
func (h *LeagueHandler) GetLeague(c *gin.Context) {
	idParam := c.Param("id")
	leagueID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid league id")
		return
	}

	league, err := h.leagueService.GetLeague(c.Request.Context(), leagueID)
	if err != nil {
		response.InternalError(c, "failed to get league")
		return
	}
	if league == nil {
		response.NotFound(c, "league not found")
		return
	}

	response.OK(c, league)
}

// GetLeagueLeaderboard handles GET /leaderboards/league/:leagueId — returns the league leaderboard.
func (h *LeagueHandler) GetLeagueLeaderboard(c *gin.Context) {
	idParam := c.Param("leagueId")
	leagueID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid league id")
		return
	}

	entries, err := h.leagueService.GetLeagueLeaderboard(c.Request.Context(), leagueID)
	if err != nil {
		response.InternalError(c, "failed to get league leaderboard")
		return
	}

	response.OK(c, entries)
}
