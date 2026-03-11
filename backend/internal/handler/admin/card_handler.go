package admin

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type CardHandler struct {
	cardService *service.CardService
	cardRepo    *postgres.CardRepo
}

func NewCardHandler(cardService *service.CardService, cardRepo *postgres.CardRepo) *CardHandler {
	return &CardHandler{
		cardService: cardService,
		cardRepo:    cardRepo,
	}
}

// List handles GET /admin/cards
func (h *CardHandler) List(c *gin.Context) {
	dateParam := c.Query("date")
	if dateParam != "" {
		date, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			response.BadRequest(c, "invalid date format, use YYYY-MM-DD")
			return
		}
		cards, err := h.cardRepo.FindByAvailableDate(c.Request.Context(), date)
		if err != nil {
			response.InternalError(c, "failed to fetch cards")
			return
		}
		response.OK(c, cards)
		return
	}

	cards, err := h.cardRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch cards")
		return
	}
	response.OK(c, cards)
}

type createCardRequest struct {
	MatchID         uuid.UUID       `json:"match_id" binding:"required"`
	QuestionText    json.RawMessage `json:"question_text" binding:"required"`
	Tier            string          `json:"tier" binding:"required"`
	HighAnswerIsYes *bool           `json:"high_answer_is_yes"`
	AvailableDate   time.Time       `json:"available_date" binding:"required"`
	ExpiresAt       time.Time       `json:"expires_at" binding:"required"`
}

// Create handles POST /admin/cards
func (h *CardHandler) Create(c *gin.Context) {
	var req createCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Validate tier
	switch req.Tier {
	case domain.TierGold, domain.TierSilver, domain.TierWhite, domain.TierVIP:
		// valid
	default:
		response.BadRequest(c, "invalid tier: must be gold, silver, white, or vip")
		return
	}

	if req.ExpiresAt.Before(req.AvailableDate) {
		response.BadRequest(c, "expires_at must be after available_date")
		return
	}

	card := &domain.Card{
		ID:              uuid.New(),
		MatchID:         req.MatchID,
		QuestionText:    req.QuestionText,
		Tier:            req.Tier,
		HighAnswerIsYes: req.HighAnswerIsYes,
		IsResolved:      false,
		AvailableDate:   req.AvailableDate,
		ExpiresAt:       req.ExpiresAt,
	}

	if err := h.cardRepo.Create(c.Request.Context(), card); err != nil {
		response.InternalError(c, "failed to create card: "+err.Error())
		return
	}

	response.Created(c, card)
}

type updateCardRequest struct {
	MatchID         *uuid.UUID      `json:"match_id"`
	QuestionText    json.RawMessage `json:"question_text"`
	Tier            string          `json:"tier"`
	HighAnswerIsYes *bool           `json:"high_answer_is_yes"`
	AvailableDate   *time.Time      `json:"available_date"`
	ExpiresAt       *time.Time      `json:"expires_at"`
}

// Update handles PUT /admin/cards/:id
func (h *CardHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid card id")
		return
	}

	var req updateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	card, err := h.cardRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch card")
		return
	}
	if card == nil {
		response.NotFound(c, "card not found")
		return
	}

	if card.IsResolved {
		response.BadRequest(c, "cannot update a resolved card")
		return
	}

	// Apply partial updates
	if req.MatchID != nil {
		card.MatchID = *req.MatchID
	}
	if req.QuestionText != nil {
		card.QuestionText = req.QuestionText
	}
	if req.Tier != "" {
		switch req.Tier {
		case domain.TierGold, domain.TierSilver, domain.TierWhite, domain.TierVIP:
			card.Tier = req.Tier
		default:
			response.BadRequest(c, "invalid tier: must be gold, silver, white, or vip")
			return
		}
	}
	if req.HighAnswerIsYes != nil {
		card.HighAnswerIsYes = req.HighAnswerIsYes
	}
	if req.AvailableDate != nil {
		card.AvailableDate = *req.AvailableDate
	}
	if req.ExpiresAt != nil {
		card.ExpiresAt = *req.ExpiresAt
	}

	if card.ExpiresAt.Before(card.AvailableDate) {
		response.BadRequest(c, "expires_at must be after available_date")
		return
	}

	if err := h.cardRepo.Update(c.Request.Context(), card); err != nil {
		response.InternalError(c, "failed to update card: "+err.Error())
		return
	}

	response.OK(c, card)
}

type resolveCardRequest struct {
	CorrectAnswer bool `json:"correct_answer"`
}

// Resolve handles POST /admin/cards/:id/resolve
func (h *CardHandler) Resolve(c *gin.Context) {
	idParam := c.Param("id")
	cardID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid card id")
		return
	}

	var req resolveCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.cardService.ResolveCard(c.Request.Context(), cardID, req.CorrectAnswer); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "card resolved successfully"})
}
