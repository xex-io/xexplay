package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type CardHandler struct {
	cardService *service.CardService
}

func NewCardHandler(cardService *service.CardService) *CardHandler {
	return &CardHandler{cardService: cardService}
}

// List handles GET /admin/cards
func (h *CardHandler) List(c *gin.Context) {
	// TODO: implement with pagination and optional match_id filter
	response.OK(c, []interface{}{})
}

// Create handles POST /admin/cards
func (h *CardHandler) Create(c *gin.Context) {
	// TODO: implement full create with validation
	response.OK(c, gin.H{"message": "card creation placeholder"})
}

// Update handles PUT /admin/cards/:id
func (h *CardHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid card id")
		return
	}

	// TODO: implement full update with validation
	response.OK(c, gin.H{"message": "card update placeholder"})
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
