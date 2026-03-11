package admin

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type MatchHandler struct {
	matchRepo *postgres.MatchRepo
}

func NewMatchHandler(matchRepo *postgres.MatchRepo) *MatchHandler {
	return &MatchHandler{matchRepo: matchRepo}
}

// List handles GET /admin/matches
func (h *MatchHandler) List(c *gin.Context) {
	eventIDParam := c.Query("event_id")
	if eventIDParam != "" {
		eventID, err := uuid.Parse(eventIDParam)
		if err != nil {
			response.BadRequest(c, "invalid event_id")
			return
		}
		matches, err := h.matchRepo.FindByEventID(c.Request.Context(), eventID)
		if err != nil {
			response.InternalError(c, "failed to fetch matches")
			return
		}
		response.OK(c, matches)
		return
	}

	matches, err := h.matchRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch matches")
		return
	}
	response.OK(c, matches)
}

type createMatchRequest struct {
	EventID     uuid.UUID       `json:"event_id" binding:"required"`
	HomeTeam    string          `json:"home_team" binding:"required"`
	AwayTeam    string          `json:"away_team" binding:"required"`
	KickoffTime time.Time       `json:"kickoff_time" binding:"required"`
	Status      string          `json:"status"`
	ResultData  json.RawMessage `json:"result_data"`
}

// Create handles POST /admin/matches
func (h *MatchHandler) Create(c *gin.Context) {
	var req createMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	status := req.Status
	if status == "" {
		status = domain.MatchStatusUpcoming
	}

	match := &domain.Match{
		ID:          uuid.New(),
		EventID:     req.EventID,
		HomeTeam:    req.HomeTeam,
		AwayTeam:    req.AwayTeam,
		KickoffTime: req.KickoffTime,
		Status:      status,
		ResultData:  req.ResultData,
	}

	if err := h.matchRepo.Create(c.Request.Context(), match); err != nil {
		response.InternalError(c, "failed to create match: "+err.Error())
		return
	}

	response.Created(c, match)
}

type updateMatchRequest struct {
	EventID     *uuid.UUID      `json:"event_id"`
	HomeTeam    string          `json:"home_team"`
	AwayTeam    string          `json:"away_team"`
	KickoffTime *time.Time      `json:"kickoff_time"`
	Status      string          `json:"status"`
	HomeScore   *int            `json:"home_score"`
	AwayScore   *int            `json:"away_score"`
	ResultData  json.RawMessage `json:"result_data"`
}

// Update handles PUT /admin/matches/:id
func (h *MatchHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid match id")
		return
	}

	var req updateMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	match, err := h.matchRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch match")
		return
	}
	if match == nil {
		response.NotFound(c, "match not found")
		return
	}

	// Apply partial updates
	if req.EventID != nil {
		match.EventID = *req.EventID
	}
	if req.HomeTeam != "" {
		match.HomeTeam = req.HomeTeam
	}
	if req.AwayTeam != "" {
		match.AwayTeam = req.AwayTeam
	}
	if req.KickoffTime != nil {
		match.KickoffTime = *req.KickoffTime
	}
	if req.Status != "" {
		match.Status = req.Status
	}
	if req.HomeScore != nil {
		match.HomeScore = req.HomeScore
	}
	if req.AwayScore != nil {
		match.AwayScore = req.AwayScore
	}
	if req.ResultData != nil {
		match.ResultData = req.ResultData
	}

	if err := h.matchRepo.Update(c.Request.Context(), match); err != nil {
		response.InternalError(c, "failed to update match: "+err.Error())
		return
	}

	response.OK(c, match)
}
