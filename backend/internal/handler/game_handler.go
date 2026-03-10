package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type GameHandler struct {
	gameService *service.GameService
}

func NewGameHandler(gameService *service.GameService) *GameHandler {
	return &GameHandler{gameService: gameService}
}

// StartSession handles POST /sessions/start
func (h *GameHandler) StartSession(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	session, err := h.gameService.StartSession(c.Request.Context(), uid)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, session.ToView())
}

// GetCurrentSession handles GET /sessions/current
func (h *GameHandler) GetCurrentSession(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Start session will return existing if one exists
	session, err := h.gameService.StartSession(c.Request.Context(), uid)
	if err != nil {
		response.NotFound(c, "no active session")
		return
	}

	response.OK(c, session.ToView())
}

// GetCurrentCard handles GET /sessions/current/card
func (h *GameHandler) GetCurrentCard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Get active session first
	session, err := h.gameService.StartSession(c.Request.Context(), uid)
	if err != nil {
		response.NotFound(c, "no active session")
		return
	}

	card, err := h.gameService.GetCurrentCard(c.Request.Context(), session)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, card)
}

type submitAnswerRequest struct {
	Answer bool `json:"answer"`
}

// SubmitAnswer handles POST /sessions/current/answer
func (h *GameHandler) SubmitAnswer(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req submitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Get active session
	session, err := h.gameService.StartSession(c.Request.Context(), uid)
	if err != nil {
		response.NotFound(c, "no active session")
		return
	}

	result, err := h.gameService.SubmitAnswer(c.Request.Context(), session, uid, req.Answer)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, result)
}

// SkipCard handles POST /sessions/current/skip
func (h *GameHandler) SkipCard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Get active session
	session, err := h.gameService.StartSession(c.Request.Context(), uid)
	if err != nil {
		response.NotFound(c, "no active session")
		return
	}

	result, err := h.gameService.SkipCard(c.Request.Context(), session, uid)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, result)
}
