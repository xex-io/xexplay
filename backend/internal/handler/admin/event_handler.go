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

type EventHandler struct {
	eventRepo *postgres.EventRepo
}

func NewEventHandler(eventRepo *postgres.EventRepo) *EventHandler {
	return &EventHandler{eventRepo: eventRepo}
}

// List handles GET /admin/events
func (h *EventHandler) List(c *gin.Context) {
	events, err := h.eventRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch events")
		return
	}

	response.OK(c, events)
}

type createEventRequest struct {
	Name              json.RawMessage `json:"name" binding:"required"`
	Description       json.RawMessage `json:"description"`
	Slug              string          `json:"slug" binding:"required"`
	StartDate         time.Time       `json:"start_date" binding:"required"`
	EndDate           time.Time       `json:"end_date" binding:"required"`
	ScoringMultiplier float64         `json:"scoring_multiplier"`
}

// Create handles POST /admin/events
func (h *EventHandler) Create(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	if req.EndDate.Before(req.StartDate) {
		response.BadRequest(c, "end_date must be after start_date")
		return
	}

	if req.ScoringMultiplier == 0 {
		req.ScoringMultiplier = 1.0
	}

	event := &domain.Event{
		ID:                uuid.New(),
		Name:              req.Name,
		Slug:              req.Slug,
		Description:       req.Description,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		IsActive:          true,
		ScoringMultiplier: req.ScoringMultiplier,
	}

	if err := h.eventRepo.Create(c.Request.Context(), event); err != nil {
		response.InternalError(c, "failed to create event: "+err.Error())
		return
	}

	response.Created(c, event)
}

type updateEventRequest struct {
	Name              json.RawMessage `json:"name"`
	Description       json.RawMessage `json:"description"`
	Slug              string          `json:"slug"`
	StartDate         *time.Time      `json:"start_date"`
	EndDate           *time.Time      `json:"end_date"`
	IsActive          *bool           `json:"is_active"`
	ScoringMultiplier *float64        `json:"scoring_multiplier"`
}

// Update handles PUT /admin/events/:id
func (h *EventHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid event id")
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Fetch existing event
	event, err := h.eventRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch event")
		return
	}
	if event == nil {
		response.NotFound(c, "event not found")
		return
	}

	// Apply partial updates
	if req.Name != nil {
		event.Name = req.Name
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.Slug != "" {
		event.Slug = req.Slug
	}
	if req.StartDate != nil {
		event.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		event.EndDate = *req.EndDate
	}
	if req.IsActive != nil {
		event.IsActive = *req.IsActive
	}
	if req.ScoringMultiplier != nil {
		event.ScoringMultiplier = *req.ScoringMultiplier
	}

	if event.EndDate.Before(event.StartDate) {
		response.BadRequest(c, "end_date must be after start_date")
		return
	}

	if err := h.eventRepo.Update(c.Request.Context(), event); err != nil {
		response.InternalError(c, "failed to update event: "+err.Error())
		return
	}

	response.OK(c, event)
}
