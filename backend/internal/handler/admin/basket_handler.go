package admin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type BasketHandler struct {
	basketRepo *postgres.BasketRepo
}

func NewBasketHandler(basketRepo *postgres.BasketRepo) *BasketHandler {
	return &BasketHandler{basketRepo: basketRepo}
}

// List handles GET /admin/baskets
func (h *BasketHandler) List(c *gin.Context) {
	dateParam := c.Query("date")
	eventIDParam := c.Query("event_id")

	// If both date and event_id are provided, use the specific finder
	if dateParam != "" && eventIDParam != "" {
		date, err := time.Parse("2006-01-02", dateParam)
		if err != nil {
			response.BadRequest(c, "invalid date format, use YYYY-MM-DD")
			return
		}
		eventID, err := uuid.Parse(eventIDParam)
		if err != nil {
			response.BadRequest(c, "invalid event_id")
			return
		}
		basket, err := h.basketRepo.FindByDateAndEvent(c.Request.Context(), date, eventID)
		if err != nil {
			response.InternalError(c, "failed to fetch basket")
			return
		}
		if basket == nil {
			response.OK(c, []interface{}{})
			return
		}
		response.OK(c, []*domain.DailyBasket{basket})
		return
	}

	baskets, err := h.basketRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch baskets")
		return
	}
	response.OK(c, baskets)
}

type createBasketRequest struct {
	EventID    uuid.UUID   `json:"event_id" binding:"required"`
	BasketDate time.Time   `json:"basket_date" binding:"required"`
	CardIDs    []uuid.UUID `json:"card_ids"`
}

// Create handles POST /admin/baskets
func (h *BasketHandler) Create(c *gin.Context) {
	var req createBasketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	basket := &domain.DailyBasket{
		ID:          uuid.New(),
		BasketDate:  req.BasketDate,
		EventID:     req.EventID,
		IsPublished: false,
	}

	if err := h.basketRepo.Create(c.Request.Context(), basket); err != nil {
		response.InternalError(c, "failed to create basket: "+err.Error())
		return
	}

	// Add cards if provided
	if len(req.CardIDs) > 0 {
		if err := h.basketRepo.AddCards(c.Request.Context(), basket.ID, req.CardIDs); err != nil {
			response.InternalError(c, "basket created but failed to add cards: "+err.Error())
			return
		}
	}

	response.Created(c, basket)
}

type updateBasketRequest struct {
	EventID    *uuid.UUID  `json:"event_id"`
	BasketDate *time.Time  `json:"basket_date"`
	CardIDs    []uuid.UUID `json:"card_ids"`
}

// Update handles PUT /admin/baskets/:id
func (h *BasketHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid basket id")
		return
	}

	var req updateBasketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	basket, err := h.basketRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch basket")
		return
	}
	if basket == nil {
		response.NotFound(c, "basket not found")
		return
	}

	if basket.IsPublished {
		response.BadRequest(c, "cannot update a published basket")
		return
	}

	// Apply partial updates
	if req.EventID != nil {
		basket.EventID = *req.EventID
	}
	if req.BasketDate != nil {
		basket.BasketDate = *req.BasketDate
	}

	if err := h.basketRepo.Update(c.Request.Context(), basket); err != nil {
		response.InternalError(c, "failed to update basket: "+err.Error())
		return
	}

	// Replace cards if provided
	if req.CardIDs != nil {
		if err := h.basketRepo.RemoveAllCards(c.Request.Context(), basket.ID); err != nil {
			response.InternalError(c, "failed to remove existing cards: "+err.Error())
			return
		}
		if len(req.CardIDs) > 0 {
			if err := h.basketRepo.AddCards(c.Request.Context(), basket.ID, req.CardIDs); err != nil {
				response.InternalError(c, "failed to add cards: "+err.Error())
				return
			}
		}
	}

	response.OK(c, basket)
}

// Delete handles DELETE /admin/baskets/:id
func (h *BasketHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid basket id")
		return
	}

	basket, err := h.basketRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch basket")
		return
	}
	if basket == nil {
		response.NotFound(c, "basket not found")
		return
	}

	if basket.IsPublished {
		response.BadRequest(c, "cannot delete a published basket")
		return
	}

	if err := h.basketRepo.Delete(c.Request.Context(), id); err != nil {
		response.InternalError(c, "failed to delete basket: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "basket deleted"})
}

// Publish handles POST /admin/baskets/:id/publish
func (h *BasketHandler) Publish(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid basket id")
		return
	}

	if err := h.basketRepo.Publish(c.Request.Context(), id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "basket published successfully"})
}
